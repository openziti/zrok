package driveClient

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Depth indicates whether a request applies to the resource's members. It's
// defined in RFC 4918 section 10.2.
type Depth int

const (
	// DepthZero indicates that the request applies only to the resource.
	DepthZero Depth = 0
	// DepthOne indicates that the request applies to the resource and its
	// internal members only.
	DepthOne Depth = 1
	// DepthInfinity indicates that the request applies to the resource and all
	// of its members.
	DepthInfinity Depth = -1
)

// ParseDepth parses a Depth header.
func ParseDepth(s string) (Depth, error) {
	switch s {
	case "0":
		return DepthZero, nil
	case "1":
		return DepthOne, nil
	case "infinity":
		return DepthInfinity, nil
	}
	return 0, fmt.Errorf("webdav: invalid Depth value")
}

// String formats the depth.
func (d Depth) String() string {
	switch d {
	case DepthZero:
		return "0"
	case DepthOne:
		return "1"
	case DepthInfinity:
		return "infinity"
	}
	panic("webdav: invalid Depth value")
}

// ParseOverwrite parses an Overwrite header.
func ParseOverwrite(s string) (bool, error) {
	switch s {
	case "T":
		return true, nil
	case "F":
		return false, nil
	}
	return false, fmt.Errorf("webdav: invalid Overwrite value")
}

// FormatOverwrite formats an Overwrite header.
func FormatOverwrite(overwrite bool) string {
	if overwrite {
		return "T"
	} else {
		return "F"
	}
}

type HTTPError struct {
	Code int
	Err  error
}

func HTTPErrorFromError(err error) *HTTPError {
	if err == nil {
		return nil
	}
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr
	} else {
		return &HTTPError{http.StatusInternalServerError, err}
	}
}

func IsNotFound(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Code == http.StatusNotFound
	}
	return false
}

func HTTPErrorf(code int, format string, a ...interface{}) *HTTPError {
	return &HTTPError{code, fmt.Errorf(format, a...)}
}

func (err *HTTPError) Error() string {
	s := fmt.Sprintf("%v %v", err.Code, http.StatusText(err.Code))
	if err.Err != nil {
		return fmt.Sprintf("%v: %v", s, err.Err)
	} else {
		return s
	}
}

func (err *HTTPError) Unwrap() error {
	return err.Err
}

type FileInfo struct {
	Path     string
	Size     int64
	ModTime  time.Time
	IsDir    bool
	MIMEType string
	ETag     string
}
