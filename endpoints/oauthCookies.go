package endpoints

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/pkg/errors"
)

// OAuthCookieConfig defines the interface that OAuth configurations must implement
// to work with the shared cookie utilities
type OAuthCookieConfig interface {
	GetCookieName() string
	GetCookieDomain() string
	GetMaxCookieSize() int
	GetSessionLifetime() time.Duration
}

// compressToken compresses a JWT token string using gzip and returns a base64-encoded string
func CompressToken(token string) (string, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if _, err := gzWriter.Write([]byte(token)); err != nil {
		return "", errors.Wrap(err, "failed to write to gzip writer")
	}

	if err := gzWriter.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close gzip writer")
	}

	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

// decompressToken decompresses a base64-encoded gzip string back to the original JWT token
func DecompressToken(compressed string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(compressed)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode base64")
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", errors.Wrap(err, "failed to create gzip reader")
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return "", errors.Wrap(err, "failed to read decompressed data")
	}

	return string(decompressed), nil
}

// GetSessionCookie retrieves and reassembles a session cookie, handling both single and striped cookies
func GetSessionCookie(r *http.Request, cookieName string) (*http.Cookie, error) {
	// get the first cookie
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	var compressedValue string

	// check if the cookie value has the stripe count prefix: {count}|{data}
	if strings.Contains(cookie.Value, "|") {
		parts := strings.SplitN(cookie.Value, "|", 2)
		if len(parts) == 2 {
			count, err := strconv.Atoi(parts[0])
			if err == nil && count > 0 {
				// this is a striped cookie
				chunks := make([]string, count)
				chunks[0] = parts[1]

				// fetch the remaining chunks
				for i := 1; i < count; i++ {
					chunkCookie, err := r.Cookie(fmt.Sprintf("%s_%d", cookieName, i))
					if err != nil {
						return nil, errors.Errorf("missing cookie chunk '%s_%d'", cookieName, i)
					}
					chunks[i] = chunkCookie.Value
				}

				// reassemble the compressed value
				compressedValue = strings.Join(chunks, "")
			} else {
				// not a valid stripe prefix, treat as single cookie
				compressedValue = cookie.Value
			}
		} else {
			compressedValue = cookie.Value
		}
	} else {
		// single cookie, no striping
		compressedValue = cookie.Value
	}

	// decompress the value
	decompressedValue, err := DecompressToken(compressedValue)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decompress cookie value")
	}

	// return a cookie with the decompressed JWT value
	return &http.Cookie{
		Name:  cookieName,
		Value: decompressedValue,
	}, nil
}

// SetSessionCookie sets a session cookie, compressing and striping it if necessary
func SetSessionCookie(w http.ResponseWriter, cookieName string, tokenValue string, cfg OAuthCookieConfig) error {
	// compress the JWT token
	compressedToken, err := CompressToken(tokenValue)
	if err != nil {
		return errors.Wrap(err, "failed to compress token")
	}

	// use default max cookie size if not configured
	maxCookieSize := cfg.GetMaxCookieSize()
	if maxCookieSize == 0 {
		maxCookieSize = 2048
	}

	// common cookie attributes
	cookieAttrs := &http.Cookie{
		MaxAge:  int(cfg.GetSessionLifetime().Seconds()),
		Domain:  cfg.GetCookieDomain(),
		Path:    "/",
		Expires: time.Now().Add(cfg.GetSessionLifetime()),
		// Secure:  true, // pending server tls feature https://github.com/openziti/zrok/issues/24
		HttpOnly: true,                 // enabled because zrok frontend is the only intended consumer of this cookie, not client-side scripts
		SameSite: http.SameSiteLaxMode, // explicitly set to the default Lax mode which allows the zrok share to be navigated to from another site and receive the cookie
	}

	// check if we need to stripe the cookie
	if len(compressedToken) > maxCookieSize {
		// calculate number of chunks needed
		chunkCount := (len(compressedToken) + maxCookieSize - 1) / maxCookieSize

		// calculate chunk size (leaving room for the count prefix in first cookie)
		prefixLen := len(fmt.Sprintf("%d|", chunkCount))
		firstChunkSize := maxCookieSize - prefixLen
		if firstChunkSize <= 0 {
			return errors.New("max cookie size too small for striping")
		}

		// set first cookie with count prefix
		firstChunk := compressedToken[:min(firstChunkSize, len(compressedToken))]
		firstCookie := *cookieAttrs
		firstCookie.Name = cookieName
		firstCookie.Value = fmt.Sprintf("%d|%s", chunkCount, firstChunk)
		http.SetCookie(w, &firstCookie)

		// set remaining chunks
		offset := firstChunkSize
		for i := 1; i < chunkCount; i++ {
			end := min(offset+maxCookieSize, len(compressedToken))
			chunk := compressedToken[offset:end]

			chunkCookie := *cookieAttrs
			chunkCookie.Name = fmt.Sprintf("%s_%d", cookieName, i)
			chunkCookie.Value = chunk
			http.SetCookie(w, &chunkCookie)

			offset = end
		}

		dl.Debugf("striped session cookie into '%d' chunks", chunkCount)
	} else {
		// single cookie is sufficient
		cookie := *cookieAttrs
		cookie.Name = cookieName
		cookie.Value = compressedToken
		http.SetCookie(w, &cookie)
	}

	return nil
}

// ClearSessionCookies clears all session cookies including striped cookie chunks
func ClearSessionCookies(w http.ResponseWriter, r *http.Request, cookieName string, cfg OAuthCookieConfig) {
	// try to determine if we have striped cookies by checking the first cookie
	cookie, err := r.Cookie(cookieName)
	if err == nil && strings.Contains(cookie.Value, "|") {
		parts := strings.SplitN(cookie.Value, "|", 2)
		if len(parts) == 2 {
			count, err := strconv.Atoi(parts[0])
			if err == nil && count > 0 {
				// clear all striped cookie chunks
				for i := 1; i < count; i++ {
					http.SetCookie(w, &http.Cookie{
						Name:     fmt.Sprintf("%s_%d", cookieName, i),
						Value:    "",
						MaxAge:   -1,
						Domain:   cfg.GetCookieDomain(),
						Path:     "/",
						HttpOnly: true,
						SameSite: http.SameSiteLaxMode,
					})
				}
			}
		}
	}

	// always clear the base cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		MaxAge:   -1,
		Domain:   cfg.GetCookieDomain(),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// FilterSessionCookies filters out session cookies and their striped chunks from a cookie list
func FilterSessionCookies(cookies []*http.Cookie, cookieName string) []*http.Cookie {
	filtered := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		// filter out the base session cookie
		if cookie.Name == cookieName {
			continue
		}
		// filter out striped session cookie chunks (e.g., cookieName_1, cookieName_2)
		if strings.HasPrefix(cookie.Name, cookieName+"_") {
			// check if the suffix is a number
			suffix := strings.TrimPrefix(cookie.Name, cookieName+"_")
			if _, err := strconv.Atoi(suffix); err == nil {
				continue
			}
		}
		// filter out pkce cookie
		if cookie.Name == "pkce" {
			continue
		}
		filtered = append(filtered, cookie)
	}
	return filtered
}
