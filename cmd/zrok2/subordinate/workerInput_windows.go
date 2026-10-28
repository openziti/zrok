//go:build windows

package subordinate

import "io"

func workerInput(input io.ReadCloser) (io.ReadCloser, error) { return input, nil }
