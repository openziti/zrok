package sync

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	pathpkg "path"
	"strings"
	"time"

	"github.com/openziti/zrok/v2/drives/davClient"
	"github.com/openziti/zrok/v2/environment/env_core"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/pkg/errors"
)

type ZrokTargetConfig struct {
	URL  *url.URL
	Root env_core.Root
}

type ZrokTarget struct {
	cfg *ZrokTargetConfig
	dc  *davClient.Client
}

type zrokDialContext struct {
	root env_core.Root
}

func (zdc *zrokDialContext) Dial(_ context.Context, _, addr string) (net.Conn, error) {
	share := strings.Split(addr, ":")[0]
	return sdk.NewDialer(share, zdc.root)
}

func NewZrokTarget(cfg *ZrokTargetConfig) (*ZrokTarget, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&zrokDialContext{cfg.Root}).Dial
	transport.TLSClientConfig.InsecureSkipVerify = true
	httpUrl := strings.Replace(cfg.URL.String(), "zrok:", "http:", 1)
	dc, err := davClient.NewClient(&http.Client{Transport: transport}, httpUrl)
	if err != nil {
		return nil, err
	}
	return &ZrokTarget{cfg: cfg, dc: dc}, nil
}

func (t *ZrokTarget) Inventory() ([]*Object, error) {
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
				ETag:     fi.ETag,
			})
		}
	}
	return objects, nil
}

func (t *ZrokTarget) Dir(path string) ([]*Object, error) {
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

func (t *ZrokTarget) Mkdir(path string) error {
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

func (t *ZrokTarget) ReadStream(path string) (io.ReadCloser, error) {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return nil, err
	}
	return t.dc.Open(context.Background(), targetPath)
}

func (t *ZrokTarget) WriteStream(path string, rs io.Reader, _ os.FileMode) error {
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

func (t *ZrokTarget) WriteStreamWithModTime(path string, rs io.Reader, _ os.FileMode, modTime time.Time) error {
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

func (t *ZrokTarget) Move(src, dest string) error {
	sourcePath, err := joinRemotePath(t.cfg.URL.Path, src)
	if err != nil {
		return err
	}
	return t.dc.MoveAll(context.Background(), sourcePath, dest, true)
}

func (t *ZrokTarget) Rm(path string) error {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return err
	}
	return t.dc.RemoveAll(context.Background(), targetPath)
}

func (t *ZrokTarget) SetModificationTime(path string, mtime time.Time) error {
	targetPath, err := joinRemotePath(t.cfg.URL.Path, path)
	if err != nil {
		return err
	}
	return t.dc.Touch(context.Background(), targetPath, mtime)
}
