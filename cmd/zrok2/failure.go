package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/openziti/zrok/v2/environment"
	restEnvironment "github.com/openziti/zrok/v2/rest_client_zrok/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/tui"
	"github.com/pkg/errors"
)

const (
	// exitTransient is the exit status for a failure a later retry may fix: the controller was unreachable,
	// busy or failing (a connection failure, any 5xx, a 429), or the failure was not a controller answer.
	exitTransient = 1

	// exitPermanent is the exit status for a failure a retry cannot fix: the controller refused the request
	// with a 4xx other than 429.
	exitPermanent = 2
)

// permanentAdvice ends every permanent failure's message.
const permanentAdvice = "retrying will not change the answer; 'zrok2 agent' retries transient failures with backoff"

// exitCodesHelp documents the exit statuses in the help of the commands that classify their failures.
const exitCodesHelp = `Exit status:
  0  success
  1  a failure a later retry may fix: the zrok controller was unreachable, busy (503 or 429) or failing (5xx); also any failure that is not an answer from the controller
  2  a failure a retry cannot fix: the zrok controller refused the request (401, 400, 404, 409 and other 4xx)`

// statusError is satisfied by every response type the generated zrok client returns as an error.
type statusError interface {
	error
	Code() int
	IsCode(int) bool
}

// permanentError is a failure decided on the client side that a retry cannot fix.
type permanentError struct {
	msg string
}

func (e *permanentError) Error() string {
	return e.msg
}

func newPermanentError(format string, v ...interface{}) error {
	return &permanentError{msg: fmt.Sprintf(format, v...)}
}

// failure is an error classified by whether a retry can change its outcome.
type failure struct {
	exitCode int
	detail   string
}

// classifyFailure decides whether err is transient or permanent and describes it for a user, against the
// environment's zrok api endpoint.
func classifyFailure(err error) failure {
	return classifyFailureAgainst(err, configuredApiEndpoint)
}

// classifyFailureAgainst is classifyFailure with the zrok api endpoint given, looked up only for a transport
// failure; one is described as the controller being unreachable only when it was a request to that endpoint.
func classifyFailureAgainst(err error, apiEndpoint func() string) failure {
	if err == nil {
		return failure{exitCode: exitTransient}
	}

	var permanent *permanentError
	if errors.As(err, &permanent) {
		return failure{exitCode: exitPermanent, detail: permanent.msg}
	}

	if status, payload, found := controllerAnswer(err); found {
		switch {
		case status == http.StatusServiceUnavailable || status == http.StatusTooManyRequests:
			return failure{exitCode: exitTransient, detail: busyMessage(retryAfterSeconds(err))}
		case status >= 500:
			return failure{exitCode: exitTransient, detail: answerDetail(status, payload)}
		case status >= 400:
			return failure{exitCode: exitPermanent, detail: answerDetail(status, payload)}
		}
	}

	if isApiTransportFailure(err, apiEndpoint) {
		return failure{exitCode: exitTransient, detail: fmt.Sprintf("unable to reach the zrok controller: %v", err)}
	}

	return failure{exitCode: exitTransient, detail: strings.TrimSpace(err.Error())}
}

// message renders the failure under msg, the description of what the command was doing, with the advice
// a permanent failure ends with.
func (f failure) message(msg string) string {
	if f.exitCode == exitPermanent {
		return f.describe(msg) + "; " + permanentAdvice
	}
	return f.describe(msg)
}

// describe renders the failure under msg without advice.
func (f failure) describe(msg string) string {
	if f.detail != "" {
		return fmt.Sprintf("%v (%v)", msg, f.detail)
	}
	return msg
}

// exitWithFailure reports err under msg and exits with the status its classification calls for; under
// --panic it panics instead.
func exitWithFailure(msg string, err error) {
	if !panicInstead {
		f := classifyFailure(err)
		_, _ = fmt.Fprintf(os.Stderr, "[%v]: %v\n", tui.ErrorLabel, f.message(msg))
		os.Exit(f.exitCode)
	}
	if err == nil {
		panic(msg)
	}
	panic(errors.Wrap(err, msg))
}

// subordinateError reports err under msg to the agent as an error message and exits with the status its
// classification calls for. the advice is left out; the agent is the reader, and it does its own pacing.
func subordinateError(msg string, err error) {
	f := classifyFailure(err)
	out := make(map[string]interface{})
	out[subordinate.MessageKey] = subordinate.ErrorMessage
	out[subordinate.ErrorMessage] = f.describe(msg)
	if data, err := json.Marshal(out); err == nil {
		fmt.Println(string(data))
	} else {
		fmt.Println("{\"" + subordinate.MessageKey + "\":\"" + subordinate.ErrorMessage + "\",\"" + subordinate.ErrorMessage + "\":\"internal error\"}")
	}
	os.Exit(f.exitCode)
}

// isApiTransportFailure reports whether err is a dial or transport failure of a request to the zrok api
// endpoint. the http client reports those as a *url.Error naming the request's url; the ziti sdk's own
// requests from a backend fail the same way, so the url's host has to be the api endpoint's.
func isApiTransportFailure(err error, apiEndpoint func() string) bool {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return false
	}
	reqUrl, err := url.Parse(urlErr.URL)
	if err != nil {
		return false
	}
	apiUrl, err := url.Parse(apiEndpoint())
	if err != nil || apiUrl.Host == "" {
		return false
	}
	return strings.EqualFold(reqUrl.Hostname(), apiUrl.Hostname()) && urlPort(reqUrl) == urlPort(apiUrl)
}

// urlPort returns u's port, defaulting from its scheme.
func urlPort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	if strings.EqualFold(u.Scheme, "http") {
		return "80"
	}
	return "443"
}

// configuredApiEndpoint returns the zrok api endpoint the environment uses, or "" when there is no
// environment to ask.
func configuredApiEndpoint() string {
	root, err := environment.LoadRoot()
	if err != nil {
		return ""
	}
	apiEndpoint, _ := root.ApiEndpoint()
	return apiEndpoint
}

// controllerAnswer returns the status and any error message of a controller response carried by err.
func controllerAnswer(err error) (status int, payload string, found bool) {
	var answer statusError
	if errors.As(err, &answer) {
		if withPayload, ok := answer.(interface {
			GetPayload() rest_model_zrok.ErrorMessage
		}); ok {
			payload = string(withPayload.GetPayload())
		}
		return answer.Code(), payload, true
	}
	var apiErr *runtime.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code, "", true
	}
	return 0, "", false
}

// retryAfterSeconds returns the Retry-After the controller answered with err, or zero when it gave none.
func retryAfterSeconds(err error) int64 {
	var shareErr *share.ShareServiceUnavailable
	if errors.As(err, &shareErr) {
		return shareErr.RetryAfter
	}
	var unshareErr *share.UnshareServiceUnavailable
	if errors.As(err, &unshareErr) {
		return unshareErr.RetryAfter
	}
	var accessErr *share.AccessServiceUnavailable
	if errors.As(err, &accessErr) {
		return accessErr.RetryAfter
	}
	var unaccessErr *share.UnaccessServiceUnavailable
	if errors.As(err, &unaccessErr) {
		return unaccessErr.RetryAfter
	}
	var enableErr *restEnvironment.EnableServiceUnavailable
	if errors.As(err, &enableErr) {
		return enableErr.RetryAfter
	}
	var disableErr *restEnvironment.DisableServiceUnavailable
	if errors.As(err, &disableErr) {
		return disableErr.RetryAfter
	}
	// an operation without a documented 503 surfaces as an APIError carrying the raw response
	var apiErr *runtime.APIError
	if errors.As(err, &apiErr) {
		if resp, ok := apiErr.Response.(runtime.ClientResponse); ok {
			if v, err := strconv.ParseInt(resp.GetHeader("Retry-After"), 10, 64); err == nil {
				return v
			}
		}
	}
	return 0
}

func busyMessage(retryAfter int64) string {
	switch {
	case retryAfter == 1:
		return "the zrok service is busy, retry in 1 second"
	case retryAfter > 1:
		return fmt.Sprintf("the zrok service is busy, retry in %d seconds", retryAfter)
	default:
		return "the zrok service is busy, retry shortly"
	}
}

func answerDetail(status int, payload string) string {
	if payload != "" {
		return payload
	}
	detail := fmt.Sprintf("the zrok controller answered '%d %v'", status, strings.ToLower(http.StatusText(status)))
	switch status {
	case http.StatusUnauthorized:
		detail += "; check the account token"
	case http.StatusNotFound:
		detail += "; check the token or resource"
	}
	return detail
}
