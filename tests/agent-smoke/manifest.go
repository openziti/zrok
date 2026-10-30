//go:build agent_smoke

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
)

const manifestName = "manifest.json"

// manifest records this run's identity and every resource it acquired, so either cleanup invocation can remove
// exactly those resources. it never holds the account token or identity material.
type manifest struct {
	RunID       string `dd:"run_id"`
	RunAttempt  string `dd:"run_attempt"`
	APIEndpoint string `dd:"api_endpoint"`
	Namespace   string `dd:"namespace"`
	Name        string `dd:"name"`
	Description string `dd:"description"`

	// NameForeign marks a name already listed before this run tried to reserve it; cleanup never touches it
	NameForeign bool `dd:"name_foreign"`
	// NameAttempted marks that a reservation was requested; an unconfirmed attempt is reconciled by exact name
	NameAttempted bool   `dd:"name_attempted"`
	NameReserved  bool   `dd:"name_reserved"`
	EnvAttempted  bool   `dd:"env_attempted"`
	EnvZID        string `dd:"env_zid"`

	ShareTokens []string        `dd:"share_tokens"`
	Processes   []processRecord `dd:"processes"`
}

// processRecord identifies an agent process by pid and start time, so a later invocation can refuse a reused pid.
// Platform says which identity format Started and Command hold: 'posix' (ps start time and command line, also
// assumed when empty) or 'windows' (creation time and executable path). cleanup acts only on its own platform's
// records.
type processRecord struct {
	PID      int    `dd:"pid"`
	Platform string `dd:"platform"`
	Started  string `dd:"started"`
	Command  string `dd:"command"`
	Reaped   bool   `dd:"reaped"`
}

// runState serializes manifest updates and persists each one before the caller proceeds.
type runState struct {
	mu   sync.Mutex
	path string
	m    manifest
}

func newRunState(dir string, m manifest) (*runState, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, manifestName)
	if _, err := os.Lstat(path); err == nil {
		return nil, fmt.Errorf("manifest '%v' already exists; a run needs a new state directory", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	st := &runState{path: path, m: m}
	if err := st.update(func(*manifest) {}); err != nil {
		return nil, err
	}
	return st, nil
}

func loadRunState(dir string) (*runState, error) {
	path := filepath.Join(dir, manifestName)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	st := &runState{path: path}
	if err := dd.BindFromJSON(&st.m, path); err != nil {
		return nil, fmt.Errorf("unable to parse manifest '%v': %w", path, err)
	}
	return st, nil
}

func (st *runState) snapshot() manifest {
	st.mu.Lock()
	defer st.mu.Unlock()
	m := st.m
	m.ShareTokens = slices.Clone(st.m.ShareTokens)
	m.Processes = slices.Clone(st.m.Processes)
	return m
}

func (st *runState) update(change func(*manifest)) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	change(&st.m)
	data, err := marshalJSON(&st.m)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(st.path), ".manifest-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), st.path)
}

func (st *runState) addShareToken(token string) error {
	added := false
	err := st.update(func(m *manifest) {
		if !slices.Contains(m.ShareTokens, token) {
			m.ShareTokens = append(m.ShareTokens, token)
			added = true
		}
	})
	if err == nil && added {
		dl.Infof("recorded share '%v'", token)
	}
	return err
}

// marshalJSON unbinds a project-owned struct with dd and encodes the result as indented JSON.
func marshalJSON(source any) ([]byte, error) {
	values, err := dd.Unbind(source)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(values, "", "  ")
}
