package basicauth

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// MaxFileSize bounds a basic authentication file at one mebibyte, far above any real list of users.
const MaxFileSize = 1 << 20

// ErrNoEntries reports a file that yields no entries; asking for basic authentication with no users is refused.
var ErrNoEntries = errors.New("no <username:password> entries")

// FileError is every failure of ReadFile. it names the path and, for a malformed entry, the line number. it never
// carries the text of a line.
type FileError struct {
	Path string
	Line int
	Err  error
}

func (e *FileError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("invalid basic authentication file '%v': line '%d': %v", e.Path, e.Line, e.Err)
	}
	return fmt.Sprintf("invalid basic authentication file '%v': %v", e.Path, e.Err)
}

func (e *FileError) Unwrap() error { return e.Err }

// ReadFile returns the entries of a basic authentication file: one <username:password> per line, either line
// ending, blank and whitespace-only lines skipped, no comment syntax. every entry is validated by ParseEntry and
// returned as written. a file with no entries, over MaxFileSize, or that cannot be read is an error.
func ReadFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, &FileError{Path: path, Err: describeOpenError(err)}
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, &FileError{Path: path, Err: describeOpenError(err)}
	}
	if info.IsDir() {
		return nil, &FileError{Path: path, Err: errors.New("is a directory")}
	}
	data, err := io.ReadAll(io.LimitReader(f, MaxFileSize+1))
	if err != nil {
		return nil, &FileError{Path: path, Err: describeOpenError(err)}
	}
	if len(data) > MaxFileSize {
		return nil, &FileError{Path: path, Err: fmt.Errorf("larger than the limit of 1 MiB (%d bytes)", MaxFileSize)}
	}
	var entries []string
	for i, line := range bytes.Split(data, []byte("\n")) {
		entry := strings.TrimSuffix(string(line), "\r")
		if strings.TrimSpace(entry) == "" {
			continue
		}
		if _, _, err := ParseEntry(entry); err != nil {
			return nil, &FileError{Path: path, Line: i + 1, Err: err}
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil, &FileError{Path: path, Err: ErrNoEntries}
	}
	return entries, nil
}

// describeOpenError keeps the cause for errors.Is while dropping the path the os package repeats in its text.
func describeOpenError(err error) error {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Err
	}
	return err
}
