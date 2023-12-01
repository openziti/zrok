package webdavClient

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DigestAuth structure holds our credentials
type DigestAuth struct {
	user        string
	pw          string
	digestParts map[string]string
}

// NewDigestAuth creates a new instance of our Digest Authenticator
func NewDigestAuth(login, secret string, rs *http.Response) (Authenticator, error) {
	return &DigestAuth{user: login, pw: secret, digestParts: digestParts(rs)}, nil
}

// Authorize the current request
func (d *DigestAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	d.digestParts["uri"] = path
	d.digestParts["method"] = rq.Method
	d.digestParts["username"] = d.user
	d.digestParts["password"] = d.pw
	rq.Header.Set("Authorization", getDigestAuthorization(d.digestParts))
	return nil
}

// Verify checks for authentication issues and may trigger a re-authentication
func (d *DigestAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	if rs.StatusCode == 401 {
		err = NewPathError("Authorize", path, rs.StatusCode)
	}
	return
}

// Close cleans up all resources
func (d *DigestAuth) Close() error {
	return nil
}

// Clone creates a copy of itself
func (d *DigestAuth) Clone() Authenticator {
	parts := make(map[string]string, len(d.digestParts))
	for k, v := range d.digestParts {
		parts[k] = v
	}
	return &DigestAuth{user: d.user, pw: d.pw, digestParts: parts}
}

// String toString
func (d *DigestAuth) String() string {
	return fmt.Sprintf("DigestAuth login: %s", d.user)
}

func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop", "opaque", "algorithm", "entityBody"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Trim(
						strings.SplitN(r, `=`, 2)[1],
						`"`,
					)
				}
			}
		}
	}
	return result
}

func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func getDigestAuthorization(digestParts map[string]string) string {
	d := digestParts
	// These are the correct ha1 and ha2 for qop=auth. We should probably check for other types of qop.

	var (
		ha1        string
		ha2        string
		nonceCount = 00000001
		cnonce     = getCnonce()
		response   string
	)

	// 'ha1' value depends on value of "algorithm" field
	switch d["algorithm"] {
	case "MD5", "":
		ha1 = getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	case "MD5-sess":
		ha1 = getMD5(
			fmt.Sprintf("%s:%v:%s",
				getMD5(d["username"]+":"+d["realm"]+":"+d["password"]),
				nonceCount,
				cnonce,
			),
		)
	}

	// 'ha2' value depends on value of "qop" field
	switch d["qop"] {
	case "auth", "":
		ha2 = getMD5(d["method"] + ":" + d["uri"])
	case "auth-int":
		if d["entityBody"] != "" {
			ha2 = getMD5(d["method"] + ":" + d["uri"] + ":" + getMD5(d["entityBody"]))
		}
	}

	// 'response' value depends on value of "qop" field
	switch d["qop"] {
	case "":
		response = getMD5(
			fmt.Sprintf("%s:%s:%s",
				ha1,
				d["nonce"],
				ha2,
			),
		)
	case "auth", "auth-int":
		response = getMD5(
			fmt.Sprintf("%s:%s:%v:%s:%s:%s",
				ha1,
				d["nonce"],
				nonceCount,
				cnonce,
				d["qop"],
				ha2,
			),
		)
	}

	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", nc=%v, cnonce="%s", response="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], nonceCount, cnonce, response)

	if d["qop"] != "" {
		authorization += fmt.Sprintf(`, qop=%s`, d["qop"])
	}

	if d["opaque"] != "" {
		authorization += fmt.Sprintf(`, opaque="%s"`, d["opaque"])
	}

	return authorization
}
