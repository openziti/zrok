package webdavClient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// PassportAuth structure holds our credentials
type PassportAuth struct {
	user            string
	pw              string
	cookies         []http.Cookie
	inhibitRedirect bool
}

// constructor for PassportAuth creates a new PassportAuth object and
// automatically authenticates against the given partnerURL
func NewPassportAuth(c *http.Client, user, pw, partnerURL string, header *http.Header) (Authenticator, error) {
	p := &PassportAuth{
		user:            user,
		pw:              pw,
		inhibitRedirect: true,
	}
	err := p.genCookies(c, partnerURL, header)
	return p, err
}

// Authorize the current request
func (p *PassportAuth) Authorize(c *http.Client, rq *http.Request, path string) error {
	// prevent redirects to detect subsequent authentication requests
	if p.inhibitRedirect {
		rq.Header.Set(XInhibitRedirect, "1")
	} else {
		p.inhibitRedirect = true
	}
	for _, cookie := range p.cookies {
		rq.AddCookie(&cookie)
	}
	return nil
}

// Verify verifies if the authentication is good
func (p *PassportAuth) Verify(c *http.Client, rs *http.Response, path string) (redo bool, err error) {
	switch rs.StatusCode {
	case 301, 302, 307, 308:
		redo = true
		if rs.Header.Get("Www-Authenticate") != "" {
			// re-authentication required as we are redirected to the login page
			err = p.genCookies(c, rs.Request.URL.String(), &rs.Header)
		} else {
			// just a redirect, follow it
			p.inhibitRedirect = false
		}
	case 401:
		err = NewPathError("Authorize", path, rs.StatusCode)
	}
	return
}

// Close cleans up all resources
func (p *PassportAuth) Close() error {
	return nil
}

// Clone creates a Copy of itself
func (p *PassportAuth) Clone() Authenticator {
	// create a copy to allow independent cookie updates
	clonedCookies := make([]http.Cookie, len(p.cookies))
	copy(clonedCookies, p.cookies)

	return &PassportAuth{
		user:            p.user,
		pw:              p.pw,
		cookies:         clonedCookies,
		inhibitRedirect: true,
	}
}

// String toString
func (p *PassportAuth) String() string {
	return fmt.Sprintf("PassportAuth login: %s", p.user)
}

func (p *PassportAuth) genCookies(c *http.Client, partnerUrl string, header *http.Header) error {
	// For more details refer to:
	// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-pass/2c80637d-438c-4d4b-adc5-903170a779f3
	// Skipping step 1 and 2 as we already have the partner server challenge

	baseAuthenticationServer := header.Get("Location")
	baseAuthenticationServerURL, err := url.Parse(baseAuthenticationServer)
	if err != nil {
		return err
	}

	// Skipping step 3 and 4 as we already know that we need and have the user's credentials
	// Step 5 (Sign-in request)
	authenticationServerUrl := url.URL{
		Scheme: baseAuthenticationServerURL.Scheme,
		Host:   baseAuthenticationServerURL.Host,
		Path:   "/login2.srf",
	}

	partnerServerChallenge := strings.Split(header.Get("Www-Authenticate"), " ")[1]

	req := http.Request{
		Method: "GET",
		URL:    &authenticationServerUrl,
		Header: http.Header{
			"Authorization": []string{"Passport1.4 sign-in=" + url.QueryEscape(p.user) + ",pwd=" + url.QueryEscape(p.pw) + ",OrgVerb=GET,OrgUrl=" + partnerUrl + "," + partnerServerChallenge},
		},
	}

	rs, err := c.Do(&req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, rs.Body)
	rs.Body.Close()
	if rs.StatusCode != 200 {
		return NewPathError("Authorize", "/", rs.StatusCode)
	}

	// Step 6 (Token Response from Authentication Server)
	tokenResponseHeader := rs.Header.Get("Authentication-Info")
	if tokenResponseHeader == "" {
		return NewPathError("Authorize", "/", 401)
	}
	tokenResponseHeaderList := strings.Split(tokenResponseHeader, ",")
	token := ""
	for _, tokenResponseHeader := range tokenResponseHeaderList {
		if strings.HasPrefix(tokenResponseHeader, "from-PP='") {
			token = tokenResponseHeader
			break
		}
	}
	if token == "" {
		return NewPathError("Authorize", "/", 401)
	}

	// Step 7 (First Authentication Request to Partner Server)
	origUrl, err := url.Parse(partnerUrl)
	if err != nil {
		return err
	}
	req = http.Request{
		Method: "GET",
		URL:    origUrl,
		Header: http.Header{
			"Authorization": []string{"Passport1.4 " + token},
		},
	}

	rs, err = c.Do(&req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, rs.Body)
	rs.Body.Close()
	if rs.StatusCode != 200 && rs.StatusCode != 302 {
		return NewPathError("Authorize", "/", rs.StatusCode)
	}

	// Step 8 (Set Token Message from Partner Server)
	cookies := rs.Header.Values("Set-Cookie")
	p.cookies = make([]http.Cookie, len(cookies))
	for i, cookie := range cookies {
		cookieParts := strings.Split(cookie, ";")
		cookieName := strings.Split(cookieParts[0], "=")[0]
		cookieValue := strings.Split(cookieParts[0], "=")[1]

		p.cookies[i] = http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		}
	}

	return nil
}
