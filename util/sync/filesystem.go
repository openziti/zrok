package sync

import (
	"context"
	"fmt"
	"golang.org/x/net/webdav"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type FilesystemTargetConfig struct {
	Root string
}

type FilesystemTarget struct {
	cfg  *FilesystemTargetConfig
	root fs.FS
	tree []*Object
}

func NewFilesystemTarget(cfg *FilesystemTargetConfig) *FilesystemTarget {
	root := os.DirFS(cfg.Root)
	return &FilesystemTarget{cfg: cfg, root: root}
}

func (t *FilesystemTarget) Inventory() ([]*Object, error) {
	if _, err := os.Stat(t.cfg.Root); os.IsNotExist(err) {
		return nil, nil
	}
	t.tree = nil
	if err := fs.WalkDir(t.root, ".", t.recurse); err != nil {
		return nil, err
	}
	return t.tree, nil
}

func (t *FilesystemTarget) recurse(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		fi, err := d.Info()
		if err != nil {
			return err
		}
		etag := ""
		if v, ok := fi.(webdav.ETager); ok {
			etag, err = v.ETag(context.Background())
			if err != nil {
				return err
			}
		} else {
			etag = fmt.Sprintf(`"%x%x"`, fi.ModTime().UTC().UnixNano(), fi.Size())
		}
		t.tree = append(t.tree, &Object{path, fi.Size(), fi.ModTime(), etag})
	}
	return nil
}

func (t *FilesystemTarget) ReadStream(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (t *FilesystemTarget) WriteStream(path string, stream io.Reader, mode os.FileMode) error {
	targetPath := filepath.Join(t.cfg.Root, path)

	if err := os.MkdirAll(filepath.Dir(targetPath), mode); err != nil {
		return err
	}
	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, stream)
	if err != nil {
		return err
	}
	return nil
}

func (t *FilesystemTarget) SetModificationTime(path string, mtime time.Time) error {
	targetPath := filepath.Join(t.cfg.Root, path)
	if err := os.Chtimes(targetPath, time.Now(), mtime); err != nil {
		return err
	}
	return nil
}
