package webdavClient

import (
	"errors"
	"fmt"
	"os"
)

// ErrAuthChanged must be returned from the Verify method as an error
// to trigger a re-authentication / negotiation with a new authenticator.
var ErrAuthChanged = errors.New("authentication failed, change algorithm")

// ErrTooManyRedirects will be used as return error if a request exceeds 10 redirects.
var ErrTooManyRedirects = errors.New("stopped after 10 redirects")

// StatusError implements error and wraps
// an erroneous status code.
type StatusError struct {
	Status int
}

func (se StatusError) Error() string {
	return fmt.Sprintf("%d", se.Status)
}

// IsErrCode returns true if the given error
// is an os.PathError wrapping a StatusError
// with the given status code.
func IsErrCode(err error, code int) bool {
	if pe, ok := err.(*os.PathError); ok {
		se, ok := pe.Err.(StatusError)
		return ok && se.Status == code
	}
	return false
}

// IsErrNotFound is shorthand for IsErrCode
// for status 404.
func IsErrNotFound(err error) bool {
	return IsErrCode(err, 404)
}

func NewPathError(op string, path string, statusCode int) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  StatusError{statusCode},
	}
}

func NewPathErrorErr(op string, path string, err error) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}
