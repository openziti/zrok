package sync

import (
	"context"
	"github.com/openziti/zrok/drives/davClient"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type WebDAVTargetConfig struct {
	URL      *url.URL
	Username string
	Password string
}

type WebDAVTarget struct {
	cfg *WebDAVTargetConfig
	dc  *davClient.Client
}

func NewWebDAVTarget(cfg *WebDAVTargetConfig) (*WebDAVTarget, error) {
	dc, err := davClient.NewClient(http.DefaultClient, cfg.URL.String())
	if err != nil {
		return nil, err
	}
	return &WebDAVTarget{cfg: cfg, dc: dc}, nil
}

func (t *WebDAVTarget) Inventory() ([]*Object, error) {
	rootFi, err := t.dc.Stat(context.Background(), t.cfg.URL.Path)
	if err != nil {
		return nil, err
	}

	if !rootFi.IsDir {
		base := filepath.Base(t.cfg.URL.Path)
		t.cfg.URL.Path = filepath.Dir(t.cfg.URL.Path)
		return []*Object{{
			Path:     "/" + base,
			IsDir:    false,
			Size:     rootFi.Size,
			Modified: rootFi.ModTime,
		}}, nil
	}

	fis, err := t.dc.Readdir(context.Background(), "", true)
	if err != nil {
		return nil, err
	}
	var objects []*Object
	for _, fi := range fis {
		if fi.Path != "/" {
			objects = append(objects, &Object{
				Path:     fi.Path,
				IsDir:    fi.IsDir,
				Size:     fi.Size,
				Modified: fi.ModTime,
				ETag:     fi.ETag,
			})
		}
	}
	return objects, nil
}

func (t *WebDAVTarget) Mkdir(path string) error {
	return t.dc.Mkdir(context.Background(), filepath.Join(t.cfg.URL.Path, path))
}

func (t *WebDAVTarget) ReadStream(path string) (io.ReadCloser, error) {
	return t.dc.Open(context.Background(), filepath.Join(t.cfg.URL.Path, path))
}

func (t *WebDAVTarget) WriteStream(path string, rs io.Reader, _ os.FileMode) error {
	ws, err := t.dc.Create(context.Background(), filepath.Join(t.cfg.URL.Path, path))
	if err != nil {
		return err
	}
	defer func() { _ = ws.Close() }()
	_, err = io.Copy(ws, rs)
	if err != nil {
		return err
	}
	return nil
}

func (t *WebDAVTarget) SetModificationTime(path string, mtime time.Time) error {
	return t.dc.Touch(context.Background(), filepath.Join(t.cfg.URL.Path, path), mtime)
}