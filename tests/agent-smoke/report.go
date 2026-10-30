//go:build agent_smoke

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/build"
)

const reportDir = "report"

// report is the allowlisted record a run or cleanup leaves for upload: step outcomes, resource identifiers and
// bounded child logs. every byte written is passed through redact first.
type report struct {
	mu      sync.Mutex
	dir     string
	mode    string
	secrets []string
	logs    map[string]*boundedLog
	order   []string
	summary summary
}

type summary struct {
	Mode       string       `dd:"mode"`
	Result     string       `dd:"result"`
	Failure    string       `dd:"failure"`
	Build      string       `dd:"build"`
	Commit     string       `dd:"commit"`
	Runner     string       `dd:"runner"`
	StopEvent  string       `dd:"stop_event"`
	RunID      string       `dd:"run_id"`
	RunAttempt string       `dd:"run_attempt"`
	Resources  resources    `dd:"resources"`
	Steps      []stepResult `dd:"steps"`
	Cleanup    *stepResult  `dd:"cleanup"`
	Started    string       `dd:"started"`
	Finished   string       `dd:"finished"`
}

type resources struct {
	APIEndpoint string   `dd:"api_endpoint"`
	Namespace   string   `dd:"namespace"`
	Name        string   `dd:"name"`
	Description string   `dd:"description"`
	EnvZID      string   `dd:"env_zid"`
	ShareTokens []string `dd:"share_tokens"`
}

type stepResult struct {
	Name     string `dd:"name"`
	Result   string `dd:"result"`
	Detail   string `dd:"detail"`
	Duration string `dd:"duration"`
}

func newReport(stateDir, mode string, secrets ...string) *report {
	return &report{
		dir:     filepath.Join(stateDir, reportDir),
		mode:    mode,
		secrets: slices.DeleteFunc(slices.Clone(secrets), func(s string) bool { return s == "" }),
		logs:    make(map[string]*boundedLog),
		summary: summary{
			Mode:    mode,
			Result:  "fail",
			Build:   build.String(),
			Commit:  os.Getenv("GITHUB_SHA"),
			Runner:  runtime.GOOS + "/" + runtime.GOARCH,
			Started: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func (r *report) redact(text string) string {
	for _, secret := range r.secrets {
		text = strings.ReplaceAll(text, secret, "[redacted]")
	}
	return text
}

func (r *report) log(label string) *boundedLog {
	r.mu.Lock()
	defer r.mu.Unlock()
	l := &boundedLog{}
	r.logs[label] = l
	r.order = append(r.order, label)
	return l
}

func (r *report) logPath(label string) string {
	return filepath.Join(reportDir, "logs", r.mode+"-"+label+".log")
}

// step runs one observation and records its outcome. the first failure is the run's failure; later steps are
// not run by the caller.
func (r *report) step(name string, fn func() (string, error)) error {
	dl.Infof("step '%v' started", name)
	started := time.Now()
	detail, err := fn()
	result := stepResult{Name: name, Result: "pass", Detail: r.redact(detail), Duration: time.Since(started).Round(time.Millisecond).String()}
	if err != nil {
		result.Result = "fail"
		result.Detail = r.redact(err.Error())
		dl.Errorf("step '%v' failed: %v", name, result.Detail)
	} else {
		dl.Infof("step '%v' passed %v", name, result.Detail)
	}
	r.mu.Lock()
	r.summary.Steps = append(r.summary.Steps, result)
	if err != nil && r.summary.Failure == "" {
		r.summary.Failure = fmt.Sprintf("%v: %v", name, result.Detail)
	}
	r.mu.Unlock()
	return err
}

// stopEvent records the graceful-stop event the run delivered and who received it.
func (r *report) stopEvent(description string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.summary.StopEvent = description
}

// fail records a failure that did not come from a step, such as cancellation, unless one is already recorded.
func (r *report) fail(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.summary.Failure == "" {
		r.summary.Failure = r.redact(err.Error())
	}
}

func (r *report) cleanup(detail string, err error, duration time.Duration) {
	result := &stepResult{Name: "cleanup", Result: "pass", Detail: r.redact(detail), Duration: duration.Round(time.Millisecond).String()}
	if err != nil {
		result.Result = "fail"
		result.Detail = r.redact(err.Error())
		dl.Errorf("cleanup failed: %v", result.Detail)
	} else {
		dl.Infof("cleanup passed %v", result.Detail)
	}
	r.mu.Lock()
	r.summary.Cleanup = result
	r.mu.Unlock()
}

func (r *report) failed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.summary.Failure != "" || (r.summary.Cleanup != nil && r.summary.Cleanup.Result != "pass")
}

// write finalizes the result and writes the summary and logs. the result passes only with no failure and, when
// cleanup ran, a passing cleanup.
func (r *report) write(m manifest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.summary.Finished = time.Now().UTC().Format(time.RFC3339)
	r.summary.RunID = m.RunID
	r.summary.RunAttempt = m.RunAttempt
	r.summary.Resources = resources{
		APIEndpoint: m.APIEndpoint,
		Namespace:   m.Namespace,
		Name:        m.Name,
		Description: m.Description,
		EnvZID:      m.EnvZID,
		ShareTokens: slices.Clone(m.ShareTokens),
	}
	r.summary.Result = "pass"
	if r.summary.Failure != "" || r.summary.Cleanup == nil || r.summary.Cleanup.Result != "pass" {
		r.summary.Result = "fail"
	}
	if r.summary.Result == "pass" {
		dl.Infof("'%v' passed; report in '%v'", r.mode, r.dir)
	} else if r.summary.Failure != "" {
		dl.Errorf("'%v' failed: %v; report in '%v'", r.mode, r.summary.Failure, r.dir)
	} else {
		dl.Errorf("'%v' failed in cleanup; report in '%v'", r.mode, r.dir)
	}

	if err := os.MkdirAll(filepath.Join(r.dir, "logs"), 0o700); err != nil {
		return err
	}
	for _, label := range r.order {
		data := []byte(r.redact(string(r.logs[label].contents())))
		if err := os.WriteFile(filepath.Join(filepath.Dir(r.dir), r.logPath(label)), data, 0o600); err != nil {
			return err
		}
	}
	data, err := marshalJSON(&r.summary)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(r.dir, r.mode+"-summary.json"), append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(r.dir, r.mode+"-summary.md"), []byte(r.redact(r.markdown())), 0o600)
}

func (r *report) markdown() string {
	s := r.summary
	var out bytes.Buffer
	_, _ = fmt.Fprintf(&out, "### agent smoke %v: %v\n\n", s.Mode, s.Result)
	_, _ = fmt.Fprintf(&out, "- build: `%v`\n- commit: `%v`\n- runner: `%v`\n- run: `%v` attempt `%v`\n", s.Build, s.Commit, s.Runner, s.RunID, s.RunAttempt)
	if s.StopEvent != "" {
		_, _ = fmt.Fprintf(&out, "- graceful stop: %v\n", s.StopEvent)
	}
	_, _ = fmt.Fprintf(&out, "- endpoint: `%v`, name: `%v:%v`\n", s.Resources.APIEndpoint, s.Resources.Namespace, s.Resources.Name)
	_, _ = fmt.Fprintf(&out, "- environment: `%v` (`%v`)\n- share tokens: `%v`\n", s.Resources.EnvZID, s.Resources.Description, strings.Join(s.Resources.ShareTokens, ", "))
	if s.Failure != "" {
		_, _ = fmt.Fprintf(&out, "\nfailure: %v\n", markdownCell(s.Failure))
	}
	out.WriteString("\n| step | result | duration | detail |\n| --- | --- | --- | --- |\n")
	rows := slices.Clone(s.Steps)
	if s.Cleanup != nil {
		rows = append(rows, *s.Cleanup)
	}
	for _, step := range rows {
		_, _ = fmt.Fprintf(&out, "| %v | %v | %v | %v |\n", step.Name, step.Result, step.Duration, markdownCell(step.Detail))
	}
	return out.String()
}

func markdownCell(text string) string {
	return strings.NewReplacer("|", "\\|", "\n", " ").Replace(text)
}
