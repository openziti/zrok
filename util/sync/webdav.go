package sync

import (
	"github.com/pkg/errors"
	"github.com/studio-b12/gowebdav"
	"io"
	"os"
	"path/filepath"
)

type WebDAVTargetConfig struct {
	URL      string
	Username string
	Password string
}

type WebDAVTarget struct {
	c *gowebdav.Client
}

func NewWebDAVTarget(cfg *WebDAVTargetConfig) (*WebDAVTarget, error) {
	c := gowebdav.NewClient(cfg.URL, cfg.Username, cfg.Password)
	if err := c.Connect(); err != nil {
		return nil, errors.Wrap(err, "error connecting to webdav target")
	}
	return &WebDAVTarget{c: c}, nil
}

func (t *WebDAVTarget) Inventory() ([]*Object, error) {
	tree, err := t.recurse("", nil)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (t *WebDAVTarget) recurse(path string, tree []*Object) ([]*Object, error) {
	files, err := t.c.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		sub := filepath.ToSlash(filepath.Join(path, f.Name()))
		if f.IsDir() {
			tree, err = t.recurse(sub, tree)
			if err != nil {
				return nil, err
			}
		} else {
			if v, ok := f.(gowebdav.File); ok {
				tree = append(tree, &Object{
					Path:     filepath.ToSlash(filepath.Join(path, f.Name())),
					Size:     v.Size(),
					Modified: v.ModTime(),
					ETag:     v.ETag(),
				})
			}
		}
	}
	return tree, nil
}

func (t *WebDAVTarget) ReadStream(path string) (io.ReadCloser, error) {
	return t.c.ReadStream(path)
}

func (t *WebDAVTarget) WriteStream(path string, stream io.Reader, mode os.FileMode) error {
	return t.c.WriteStream(path, stream, mode)
}
