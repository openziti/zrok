package driveClient

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/util/sync/driveClient/internal"
	"io"
	"net/http"
	"time"
)

// HTTPClient performs HTTP requests. It's implemented by *http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type basicAuthHTTPClient struct {
	c                  HTTPClient
	username, password string
}

func (c *basicAuthHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(c.username, c.password)
	return c.c.Do(req)
}

// HTTPClientWithBasicAuth returns an HTTP client that adds basic
// authentication to all outgoing requests. If c is nil, http.DefaultClient is
// used.
func HTTPClientWithBasicAuth(c HTTPClient, username, password string) HTTPClient {
	if c == nil {
		c = http.DefaultClient
	}
	return &basicAuthHTTPClient{c, username, password}
}

// Client provides access to a remote WebDAV filesystem.
type Client struct {
	ic *internal.Client
}

func NewClient(c HTTPClient, endpoint string) (*Client, error) {
	ic, err := internal.NewClient(c, endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{ic}, nil
}

func (c *Client) FindCurrentUserPrincipal(ctx context.Context) (string, error) {
	propfind := internal.NewPropNamePropFind(internal.CurrentUserPrincipalName)

	// TODO: consider retrying on the root URI "/" if this fails, as suggested
	// by the RFC?
	resp, err := c.ic.PropFindFlat(ctx, "", propfind)
	if err != nil {
		return "", err
	}

	var prop internal.CurrentUserPrincipal
	if err := resp.DecodeProp(&prop); err != nil {
		return "", err
	}
	if prop.Unauthenticated != nil {
		return "", fmt.Errorf("webdav: unauthenticated")
	}

	return prop.Href.Path, nil
}

var fileInfoPropFind = internal.NewPropNamePropFind(
	internal.ResourceTypeName,
	internal.GetContentLengthName,
	internal.GetLastModifiedName,
	internal.GetContentTypeName,
	internal.GetETagName,
)

func fileInfoFromResponse(resp *internal.Response) (*FileInfo, error) {
	path, err := resp.Path()
	if err != nil {
		return nil, err
	}

	fi := &FileInfo{Path: path}

	var resType internal.ResourceType
	if err := resp.DecodeProp(&resType); err != nil {
		return nil, err
	}

	if resType.Is(internal.CollectionName) {
		fi.IsDir = true
	} else {
		var getLen internal.GetContentLength
		if err := resp.DecodeProp(&getLen); err != nil {
			return nil, err
		}

		var getType internal.GetContentType
		if err := resp.DecodeProp(&getType); err != nil && !internal.IsNotFound(err) {
			return nil, err
		}

		var getETag internal.GetETag
		if err := resp.DecodeProp(&getETag); err != nil && !internal.IsNotFound(err) {
			return nil, err
		}

		fi.Size = getLen.Length
		fi.MIMEType = getType.Type
		fi.ETag = string(getETag.ETag)
	}

	var getMod internal.GetLastModified
	if err := resp.DecodeProp(&getMod); err != nil && !internal.IsNotFound(err) {
		return nil, err
	}
	fi.ModTime = time.Time(getMod.LastModified)

	return fi, nil
}

func (c *Client) Stat(ctx context.Context, name string) (*FileInfo, error) {
	resp, err := c.ic.PropFindFlat(ctx, name, fileInfoPropFind)
	if err != nil {
		return nil, err
	}
	return fileInfoFromResponse(resp)
}

func (c *Client) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	req, err := c.ic.NewRequest(http.MethodGet, name, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.ic.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) Readdir(ctx context.Context, name string, recursive bool) ([]FileInfo, error) {
	depth := internal.DepthOne
	if recursive {
		depth = internal.DepthInfinity
	}

	ms, err := c.ic.PropFind(ctx, name, depth, fileInfoPropFind)
	if err != nil {
		return nil, err
	}

	l := make([]FileInfo, 0, len(ms.Responses))
	for _, resp := range ms.Responses {
		fi, err := fileInfoFromResponse(&resp)
		if err != nil {
			return l, err
		}
		l = append(l, *fi)
	}

	return l, nil
}

type fileWriter struct {
	pw   *io.PipeWriter
	done <-chan error
}

func (fw *fileWriter) Write(b []byte) (int, error) {
	return fw.pw.Write(b)
}

func (fw *fileWriter) Close() error {
	if err := fw.pw.Close(); err != nil {
		return err
	}
	return <-fw.done
}

func (c *Client) Create(ctx context.Context, name string) (io.WriteCloser, error) {
	pr, pw := io.Pipe()

	req, err := c.ic.NewRequest(http.MethodPut, name, pr)
	if err != nil {
		pw.Close()
		return nil, err
	}

	done := make(chan error, 1)
	go func() {
		resp, err := c.ic.Do(req.WithContext(ctx))
		if err != nil {
			done <- err
			return
		}
		resp.Body.Close()
		done <- nil
	}()

	return &fileWriter{pw, done}, nil
}

func (c *Client) RemoveAll(ctx context.Context, name string) error {
	req, err := c.ic.NewRequest(http.MethodDelete, name, nil)
	if err != nil {
		return err
	}

	resp, err := c.ic.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) Mkdir(ctx context.Context, name string) error {
	req, err := c.ic.NewRequest("MKCOL", name, nil)
	if err != nil {
		return err
	}

	resp, err := c.ic.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) CopyAll(ctx context.Context, name, dest string, overwrite bool) error {
	req, err := c.ic.NewRequest("COPY", name, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Destination", c.ic.ResolveHref(dest).String())
	req.Header.Set("Overwrite", internal.FormatOverwrite(overwrite))

	resp, err := c.ic.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) MoveAll(ctx context.Context, name, dest string, overwrite bool) error {
	req, err := c.ic.NewRequest("MOVE", name, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Destination", c.ic.ResolveHref(dest).String())
	req.Header.Set("Overwrite", internal.FormatOverwrite(overwrite))

	resp, err := c.ic.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
