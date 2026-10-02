package automation

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_model"
)

// authenticateLimiter names the ziti controller's /authenticate rate limiter, whose generated 429 is
// typed rather than seen by the session transport.
const authenticateLimiter = "authenticate"

// rateLimitedBodyLimit bounds how much of a 429 body is read for its error envelope.
const rateLimitedBodyLimit = 64 * 1024

// RateLimitedError reports that a ziti rate limiter refused a management request. Limiter is the
// error code from ziti's envelope (SERVER_TOO_MANY_REQUESTS for the command rate limiter), or
// 'authenticate' for the authentication limiter.
type RateLimitedError struct {
	Limiter string
	Method  string
	Path    string
}

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("ziti rate limited '%s' on '%s %s'", e.Limiter, e.Method, e.Path)
}

// newRateLimitedError counts and logs one rate-limited answer from ziti.
func newRateLimitedError(limiter, method, path string) *RateLimitedError {
	zitiRateLimited.Add(1)
	dl.Warnf("ziti rate limited by '%s' on '%s %s', rate limited count '%d'", limiter, method, path, zitiRateLimited.Value())
	return &RateLimitedError{Limiter: limiter, Method: method, Path: path}
}

// rateLimitedResponse consumes a 429 response and returns its error; the limiter is the code in
// ziti's error envelope, when the body carries one.
func rateLimitedResponse(req *http.Request, resp *http.Response) *RateLimitedError {
	defer resp.Body.Close()
	limiter := http.StatusText(http.StatusTooManyRequests)
	var envelope rest_model.APIErrorEnvelope
	if body, err := io.ReadAll(io.LimitReader(resp.Body, rateLimitedBodyLimit)); err == nil {
		if json.Unmarshal(body, &envelope) == nil && envelope.Error != nil && envelope.Error.Code != "" {
			limiter = envelope.Error.Code
		}
	}
	return newRateLimitedError(limiter, req.Method, req.URL.Path)
}
