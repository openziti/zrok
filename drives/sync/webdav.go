package sync

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	pathpkg "path"
	"time"

	"github.com/openziti/zrok/v2/drives/davClient"
	"github.com/pkg/errors"
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
	var httpClient davClient.HTTPClient
	httpClient = http.DefaultClient
	if cfg.Username != "" || cfg.Password != "" {
		httpClient = davClient.HTTPClientWithBasicAuth(httpClient, cfg.Username, cfg.Password)
	}
	dc, err := davClient.NewClient(httpClient, cfg.URL.String())
	if err != nil {
		return nil, err
	}
	return &WebDAVTarget{cfg: cfg, dc: dc}, nil
}

func (t *WebDAVTarget) Inventory() ([]*Object, error) {
	rootPath, err := cleanVirtualPath(t.cfg.URL.Path)
	if err != nil {
		return nil, err
	}

	rootFi, err := t.dc.Stat(context.Background(), rootPath)
	if err != nil {
		return nil, err
	}

	if !rootFi.IsDir {
		objectPath, err := remoteFileObjectPath(rootPath, rootFi.Path)
		if err != nil {
			return nil, err
		}
		t.cfg.URL.Path = pathpkg.Dir(rootPath)
		return []*Object{{
			Path:     objectPath,
			IsDir:    false,
			Size:     rootFi.Size,
			Modified: rootFi.ModTime,
		}}, nil
	}

	if _, err := remoteObjectPath(rootPath, rootFi.Path, true); err != nil {
		return nil, err
	}

	fis, err := t.dc.Readdir(context.Background(), rootPath, true)
	if err != nil {
		return nil, err
	}
	var objects []*Object
	for _, fi := range fis {
		objectPath, err := remoteObjectPath(rootPath, fi.Path, fi.IsDir)
		if err != nil {
			return nil, err
		}
		if objectPath != "/" {
			objects = append(objects, &Object{
				Path:     objectPath,
				IsDir:    fi.IsDir,
				Size:     fi.Size,
				Modified: fi.ModTime,
			})
		}
	}
	return objects, nil
}

func (t *WebDAVTarget) Dir(path string) ([]*Object, error) {
	fis, err := t.dc.Readdir(context.Background(), t.cfg.URL.Path, false)
	if err != nil {
		return nil, err
	}
	var objects []*Object
	for _, fi := range fis {
		if fi.Path != "/" && fi.Path != t.cfg.URL.Path+"/" {
			objects = append(objects, &Object{
				Path:     pathpkg.Base(fi.Path),
				IsDir:    fi.IsDir,
				Size:     fi.Size,
				Modified: fi.ModTime,
			})
		}
	}
	return objects, nil
}

func (t *WebDAVTarget) Mkdir(path string) error {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return err
	}

	fi, err := t.dc.Stat(context.Background(), targetPath)
	if err == nil {
		if fi.IsDir {
			return nil
		}
		return errors.Errorf("'%v' already exists; not directory", path)
	}
	return t.dc.Mkdir(context.Background(), targetPath)
}

func (t *WebDAVTarget) ReadStream(path string) (io.ReadCloser, error) {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return nil, err
	}
	return t.dc.Open(context.Background(), targetPath)
}

func (t *WebDAVTarget) WriteStream(path string, rs io.Reader, _ os.FileMode) error {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return err
	}

	ws, err := t.dc.Create(context.Background(), targetPath)
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

func (t *WebDAVTarget) WriteStreamWithModTime(path string, rs io.Reader, _ os.FileMode, modTime time.Time) error {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return err
	}

	ws, err := t.dc.CreateWithModTime(context.Background(), targetPath, modTime)
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

func (t *WebDAVTarget) Move(src, dest string) error {
	sourcePath, err := joinRemotePath(t.cfg.URL.Path, src)
	if err != nil {
		return err
	}
	return t.dc.MoveAll(context.Background(), sourcePath, dest, true)
}

func (t *WebDAVTarget) Rm(path string) error {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return err
	}
	return t.dc.RemoveAll(context.Background(), targetPath)
}

func (t *WebDAVTarget) SetModificationTime(path string, mtime time.Time) error {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return err
	}
	return t.dc.Touch(context.Background(), targetPath, mtime)
}
