package sync

import (
	"context"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/util/sync/driveClient"
	"io"
	"net/http"
	"net/url"
	"os"
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
	dc    *driveClient.Client
	isDir bool
}

func NewWebDAVTarget(cfg *WebDAVTargetConfig) (*WebDAVTarget, error) {
	dc, err := driveClient.NewClient(http.DefaultClient, cfg.URL.String())
	if err != nil {
		return nil, err
	}
	return &WebDAVTarget{cfg: cfg, dc: dc}, nil
}

func (t *WebDAVTarget) Inventory() ([]*Object, error) {
	fis, err := t.dc.Readdir(context.Background(), "", true)
	if err != nil {
		return nil, err
	}
	var objects []*Object
	for _, fi := range fis {
		if !fi.IsDir {
			objects = append(objects, &Object{
				Path:     fi.Path,
				Size:     fi.Size,
				Modified: fi.ModTime,
				ETag:     fi.ETag,
			})
		}
	}
	return objects, nil
}

func (t *WebDAVTarget) ReadStream(path string) (io.ReadCloser, error) {
	return t.dc.Open(context.Background(), path)
}

func (t *WebDAVTarget) WriteStream(path string, rs io.Reader, _ os.FileMode) error {
	ws, err := t.dc.Create(context.Background(), path)
	if err != nil {
		return err
	}
	defer ws.Close()
	_, err = io.Copy(ws, rs)
	if err != nil {
		return err
	}
	return nil
}

func (t *WebDAVTarget) SetModificationTime(path string, mtime time.Time) error {
	return t.dc.Touch(context.Background(), path, mtime)
}
