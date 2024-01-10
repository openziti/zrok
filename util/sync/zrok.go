package sync

import (
	"context"
	"github.com/openziti/zrok/drives/davClient"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ZrokTargetConfig struct {
	URL      *url.URL
	Username string
	Password string
	Root     env_core.Root
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

func (t *ZrokTarget) Mkdir(path string) error {
	return t.dc.Mkdir(context.Background(), filepath.Join(t.cfg.URL.Path, path))
}

func (t *ZrokTarget) ReadStream(path string) (io.ReadCloser, error) {
	return t.dc.Open(context.Background(), filepath.Join(t.cfg.URL.Path, path))
}

func (t *ZrokTarget) WriteStream(path string, rs io.Reader, _ os.FileMode) error {
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

func (t *ZrokTarget) SetModificationTime(path string, mtime time.Time) error {
	return t.dc.Touch(context.Background(), filepath.Join(t.cfg.URL.Path, path), mtime)
}
