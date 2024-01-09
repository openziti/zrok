package sync

import (
	"io"
	"os"
	"time"
)

type Object struct {
	Path     string
	IsDir    bool
	Size     int64
	Modified time.Time
	ETag     string
}

type Target interface {
	Inventory() ([]*Object, error)
	Mkdir(path string) error
	ReadStream(path string) (io.ReadCloser, error)
	WriteStream(path string, stream io.Reader, mode os.FileMode) error
	SetModificationTime(path string, mtime time.Time) error
}
