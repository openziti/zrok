package webdavClient

import (
	"io"
	"log"
	"net/http"
	"path"
	"strings"
)

func (c *Client) req(method, path string, body io.Reader, intercept func(*http.Request)) (rs *http.Response, err error) {
	var redo bool
	var r *http.Request
	var uri = PathEscape(Join(c.root, path))
	auth, body := c.auth.NewAuthenticator(body)
	defer auth.Close()

	for { // TODO auth.continue() strategy(true|n times|until)?
		if r, err = http.NewRequest(method, uri, body); err != nil {
			return
		}

		for k, vals := range c.headers {
			for _, v := range vals {
				r.Header.Add(k, v)
			}
		}

		if err = auth.Authorize(c.c, r, path); err != nil {
			return
		}

		if intercept != nil {
			intercept(r)
		}

		if c.interceptor != nil {
			c.interceptor(method, r)
		}

		if rs, err = c.c.Do(r); err != nil {
			return
		}

		if redo, err = auth.Verify(c.c, rs, path); err != nil {
			rs.Body.Close()
			return nil, err
		}
		if redo {
			rs.Body.Close()
			if body, err = r.GetBody(); err != nil {
				return nil, err
			}
			continue
		}
		break
	}

	return rs, err
}

func (c *Client) mkcol(path string) (status int, err error) {
	rs, err := c.req("MKCOL", path, nil, nil)
	if err != nil {
		return
	}
	defer rs.Body.Close()

	status = rs.StatusCode
	if status == 405 {
		status = 201
	}

	return
}

func (c *Client) options(path string) (*http.Response, error) {
	return c.req("OPTIONS", path, nil, func(rq *http.Request) {
		rq.Header.Add("Depth", "0")
	})
}

func (c *Client) propfind(path string, self bool, body string, resp interface{}, parse func(resp interface{}) error) error {
	rs, err := c.req("PROPFIND", path, strings.NewReader(body), func(rq *http.Request) {
		if self {
			rq.Header.Add("Depth", "0")
		} else {
			rq.Header.Add("Depth", "1")
		}
		rq.Header.Add("Content-Type", "application/xml;charset=UTF-8")
		rq.Header.Add("Accept", "application/xml,text/xml")
		rq.Header.Add("Accept-Charset", "utf-8")
		// TODO add support for 'gzip,deflate;q=0.8,q=0.7'
		rq.Header.Add("Accept-Encoding", "")
	})
	if err != nil {
		return err
	}
	defer rs.Body.Close()

	if rs.StatusCode != 207 {
		return NewPathError("PROPFIND", path, rs.StatusCode)
	}

	return parseXML(rs.Body, resp, parse)
}

func (c *Client) doCopyMove(
	method string,
	oldpath string,
	newpath string,
	overwrite bool,
) (
	status int,
	r io.ReadCloser,
	err error,
) {
	rs, err := c.req(method, oldpath, nil, func(rq *http.Request) {
		rq.Header.Add("Destination", PathEscape(Join(c.root, newpath)))
		if overwrite {
			rq.Header.Add("Overwrite", "T")
		} else {
			rq.Header.Add("Overwrite", "F")
		}
	})
	if err != nil {
		return
	}
	status = rs.StatusCode
	r = rs.Body
	return
}

func (c *Client) copymove(method string, oldpath string, newpath string, overwrite bool) (err error) {
	s, data, err := c.doCopyMove(method, oldpath, newpath, overwrite)
	if err != nil {
		return
	}
	if data != nil {
		defer data.Close()
	}

	switch s {
	case 201, 204:
		return nil

	case 207:
		// TODO handle multistat errors, worst case ...
		log.Printf("TODO handle %s - %s multistatus result %s\n", method, oldpath, String(data))

	case 409:
		err := c.createParentCollection(newpath)
		if err != nil {
			return err
		}

		return c.copymove(method, oldpath, newpath, overwrite)
	}

	return NewPathError(method, oldpath, s)
}

func (c *Client) put(path string, stream io.Reader) (status int, err error) {
	rs, err := c.req("PUT", path, stream, nil)
	if err != nil {
		return
	}
	defer rs.Body.Close()

	status = rs.StatusCode
	return
}

func (c *Client) createParentCollection(itemPath string) (err error) {
	parentPath := path.Dir(itemPath)
	if parentPath == "." || parentPath == "/" {
		return nil
	}

	return c.MkdirAll(parentPath, 0755)
}
