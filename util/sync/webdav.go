package sync

import (
	"fmt"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/util/sync/webdavClient"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type WebDAVTargetConfig struct {
	URL      *url.URL
	Username string
	Password string
	Root     env_core.Root
}

type WebDAVTarget struct {
	cfg   *WebDAVTargetConfig
	c     *webdavClient.Client
	isDir bool
}

func NewWebDAVTarget(cfg *WebDAVTargetConfig) (*WebDAVTarget, error) {
	c, err := webdavClient.NewZrokClient(cfg.URL, cfg.Root, webdavClient.NewAutoAuth(cfg.Username, cfg.Password))
	if err != nil {
		return nil, err
	}
	if err := c.Connect(); err != nil {
		return nil, errors.Wrap(err, "error connecting to webdav target")
	}
	return &WebDAVTarget{cfg: cfg, c: c}, nil
}

func (t *WebDAVTarget) Inventory() ([]*Object, error) {
	fi, err := t.c.Stat("")
	if !fi.IsDir() {
		t.isDir = false
		return []*Object{{
			Path:     fi.Name(),
			Size:     fi.Size(),
			Modified: fi.ModTime(),
		}}, nil
	}

	t.isDir = true
	tree, err := t.recurse("", nil)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (t *WebDAVTarget) IsDir() bool {
	return t.isDir
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
			if v, ok := f.(webdavClient.File); ok {
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
	if t.isDir {
		return t.c.ReadStream(path)
	}
	return t.c.ReadStream("")
}

func (t *WebDAVTarget) WriteStream(path string, stream io.Reader, mode os.FileMode) error {
	return t.c.WriteStream(path, stream, mode)
}

func (t *WebDAVTarget) SetModificationTime(path string, mtime time.Time) error {
	modtimeUnix := mtime.Unix()
	body := "<?xml version=\"1.0\" encoding=\"utf-8\" ?>" +
		"<propertyupdate xmlns=\"DAV:\" xmlns:z=\"zrok:\"><set><prop><z:lastmodified>" +
		fmt.Sprintf("%d", modtimeUnix) +
		"</z:lastmodified></prop></set></propertyupdate>"
	if err := t.c.Proppatch(path, body, nil, nil); err != nil {
		return err
	}
	return nil
}
