package sync

import (
	"io"
	"os"
	"time"
)

type Object struct {
	Path     string
	Size     int64
	Modified time.Time
	ETag     string
}

type Target interface {
	Inventory() ([]*Object, error)
	IsDir() bool
	ReadStream(path string) (io.ReadCloser, error)
	WriteStream(path string, stream io.Reader, mode os.FileMode) error
	SetModificationTime(path string, mtime time.Time) error
}
