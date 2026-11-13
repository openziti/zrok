package basicauth

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeUsersFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "users")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return path
}

func TestReadFileEntries(t *testing.T) {
	tests := map[string]struct {
		content string
		entries []string
	}{
		"colon in password":         {"admin:s3:cret\n", []string{"admin:s3:cret"}},
		"crlf":                      {"admin:one\r\nuser:two\r\n", []string{"admin:one", "user:two"}},
		"mixed line endings":        {"admin:one\r\nuser:two\nthird:three", []string{"admin:one", "user:two", "third:three"}},
		"blank lines":               {"\n\nadmin:one\n\n \t\n\r\nuser:two\n\n", []string{"admin:one", "user:two"}},
		"trailing newline absent":   {"admin:one\nuser:two", []string{"admin:one", "user:two"}},
		"no comment syntax":         {"#admin:one\n;user:two\n", []string{"#admin:one", ";user:two"}},
		"entries returned as given": {"  admin : one  \n", []string{"  admin : one  "}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			entries, err := ReadFile(writeUsersFile(t, tc.content))
			require.NoError(t, err)
			require.Equal(t, tc.entries, entries)
		})
	}
}

func TestReadFileNoEntries(t *testing.T) {
	for name, content := range map[string]string{"empty": "", "newlines": "\n\r\n\n", "whitespace only": "  \t\n   \r\n"} {
		t.Run(name, func(t *testing.T) {
			path := writeUsersFile(t, content)
			_, err := ReadFile(path)
			var fileErr *FileError
			require.ErrorAs(t, err, &fileErr)
			require.ErrorIs(t, err, ErrNoEntries)
			require.Contains(t, err.Error(), path)
		})
	}
}

func TestReadFileMalformedLineIsNotQuoted(t *testing.T) {
	path := writeUsersFile(t, "admin:one\n\nuser:two\n"+sentinel+"\nthird:three\n")
	_, err := ReadFile(path)
	var fileErr *FileError
	require.ErrorAs(t, err, &fileErr)
	require.Equal(t, 4, fileErr.Line)
	require.ErrorIs(t, err, ErrMalformedEntry)
	require.Contains(t, err.Error(), path)
	require.Contains(t, err.Error(), "line '4'")
	require.NotContains(t, err.Error(), sentinel)
}

func TestReadFileSizeLimit(t *testing.T) {
	line := "admin:secret\n"
	atLimit := strings.Repeat(line, MaxFileSize/len(line))
	atLimit += strings.Repeat("x", MaxFileSize-len(atLimit)-1) + ":"
	require.Len(t, atLimit, MaxFileSize)
	entries, err := ReadFile(writeUsersFile(t, atLimit))
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	path := writeUsersFile(t, atLimit+"\n")
	_, err = ReadFile(path)
	var fileErr *FileError
	require.ErrorAs(t, err, &fileErr)
	require.Contains(t, err.Error(), path)
	require.Contains(t, err.Error(), "1 MiB")
	require.Contains(t, err.Error(), "1048576")
}

func TestReadFileUnusablePaths(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")
	_, err := ReadFile(missing)
	var fileErr *FileError
	require.ErrorAs(t, err, &fileErr)
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.Contains(t, err.Error(), missing)

	_, err = ReadFile(dir)
	require.ErrorAs(t, err, &fileErr)
	require.Contains(t, err.Error(), dir)
	require.Contains(t, err.Error(), "is a directory")

	if runtime.GOOS != "windows" && os.Geteuid() != 0 {
		unreadable := writeUsersFile(t, "admin:"+sentinel+"\n")
		require.NoError(t, os.Chmod(unreadable, 0))
		_, err = ReadFile(unreadable)
		require.ErrorAs(t, err, &fileErr)
		require.True(t, errors.Is(err, fs.ErrPermission))
		require.Contains(t, err.Error(), unreadable)
		require.NotContains(t, err.Error(), sentinel)
	}
}
