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

	"github.com/sirupsen/logrus"
)

// OAuthCookieConfig defines the interface for OAuth cookie configuration
// This allows different proxy types to implement their own config structs while sharing cookie utilities
type OAuthCookieConfig interface {
	GetCookieName() string
	GetCookieDomain() string
	GetMaxCookieSize() int
	GetSessionLifetime() time.Duration
}

// CompressToken compresses a token string using gzip and returns base64-encoded result
func CompressToken(token string) (string, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write([]byte(token)); err != nil {
		return "", fmt.Errorf("error writing to gzip writer: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return "", fmt.Errorf("error closing gzip writer: %w", err)
	}
	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}

// DecompressToken decompresses a base64-encoded, gzip-compressed token string
func DecompressToken(compressed string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(compressed)
	if err != nil {
		return "", fmt.Errorf("error decoding base64: %w", err)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gzipReader.Close()

	decompressed, err := io.ReadAll(gzipReader)
	if err != nil {
		return "", fmt.Errorf("error reading from gzip reader: %w", err)
	}

	return string(decompressed), nil
}

// GetSessionCookie retrieves and reassembles a session cookie that may be striped across multiple cookies
func GetSessionCookie(r *http.Request, cookieName string) (*http.Cookie, error) {
	baseCookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, err
	}

	// check if this is a striped cookie by looking for the count prefix
	parts := strings.SplitN(baseCookie.Value, "|", 2)
	if len(parts) != 2 {
		// not striped, decompress and return
		decompressed, err := DecompressToken(baseCookie.Value)
		if err != nil {
			return nil, fmt.Errorf("error decompressing cookie: %w", err)
		}
		return &http.Cookie{
			Name:  cookieName,
			Value: decompressed,
		}, nil
	}

	// striped cookie - reassemble all chunks
	count, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid cookie count prefix: %w", err)
	}

	// start with the data from the base cookie
	var reassembled strings.Builder
	reassembled.WriteString(parts[1])

	// retrieve and append the numbered chunks
	for i := 1; i < count; i++ {
		chunkName := fmt.Sprintf("%s_%d", cookieName, i)
		chunk, err := r.Cookie(chunkName)
		if err != nil {
			return nil, fmt.Errorf("missing cookie chunk %s: %w", chunkName, err)
		}
		reassembled.WriteString(chunk.Value)
	}

	// decompress the reassembled data
	decompressed, err := DecompressToken(reassembled.String())
	if err != nil {
		return nil, fmt.Errorf("error decompressing reassembled cookie: %w", err)
	}

	return &http.Cookie{
		Name:  cookieName,
		Value: decompressed,
	}, nil
}

// SetSessionCookie compresses and stripes a session cookie across multiple cookies if needed
func SetSessionCookie(w http.ResponseWriter, cookieName string, tokenValue string, cfg OAuthCookieConfig) error {
	// compress the token
	compressed, err := CompressToken(tokenValue)
	if err != nil {
		return fmt.Errorf("error compressing token: %w", err)
	}

	maxSize := cfg.GetMaxCookieSize()
	if maxSize == 0 {
		maxSize = 3072
	}

	// if compressed data fits in a single cookie, set it directly
	if len(compressed) <= maxSize {
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    compressed,
			MaxAge:   int(cfg.GetSessionLifetime().Seconds()),
			Domain:   cfg.GetCookieDomain(),
			Path:     "/",
			Expires:  time.Now().Add(cfg.GetSessionLifetime()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		return nil
	}

	// need to stripe across multiple cookies
	logrus.Debugf("cookie size %d exceeds max %d, striping across multiple cookies", len(compressed), maxSize)

	// calculate how many cookies we need
	// account for the count prefix in the first cookie (e.g., "3|")
	countPrefixSize := len(fmt.Sprintf("%d|", (len(compressed)/maxSize)+2)) // estimate
	firstChunkSize := maxSize - countPrefixSize
	remainingSize := len(compressed) - firstChunkSize
	additionalChunks := (remainingSize + maxSize - 1) / maxSize // ceiling division
	totalCookies := additionalChunks + 1

	// set the base cookie with count prefix
	firstChunkData := compressed[:firstChunkSize]
	baseValue := fmt.Sprintf("%d|%s", totalCookies, firstChunkData)

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    baseValue,
		MaxAge:   int(cfg.GetSessionLifetime().Seconds()),
		Domain:   cfg.GetCookieDomain(),
		Path:     "/",
		Expires:  time.Now().Add(cfg.GetSessionLifetime()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// set the numbered chunks
	offset := firstChunkSize
	for i := 1; i < totalCookies; i++ {
		chunkName := fmt.Sprintf("%s_%d", cookieName, i)
		end := offset + maxSize
		if end > len(compressed) {
			end = len(compressed)
		}
		chunkData := compressed[offset:end]

		http.SetCookie(w, &http.Cookie{
			Name:     chunkName,
			Value:    chunkData,
			MaxAge:   int(cfg.GetSessionLifetime().Seconds()),
			Domain:   cfg.GetCookieDomain(),
			Path:     "/",
			Expires:  time.Now().Add(cfg.GetSessionLifetime()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		offset = end
	}

	return nil
}

// ClearSessionCookies clears all session cookies including any striped chunks
func ClearSessionCookies(w http.ResponseWriter, r *http.Request, cookieName string, cfg OAuthCookieConfig) {
	// iterate through all cookies and clear any that match the session cookie pattern
	for _, cookie := range r.Cookies() {
		// clear base cookie or any numbered chunks (cookieName_1, cookieName_2, etc.)
		if cookie.Name == cookieName || strings.HasPrefix(cookie.Name, cookieName+"_") {
			http.SetCookie(w, &http.Cookie{
				Name:     cookie.Name,
				Value:    "",
				MaxAge:   -1,
				Domain:   cfg.GetCookieDomain(),
				Path:     "/",
				HttpOnly: true,
			})
		}
	}
}

// FilterSessionCookies filters out session cookies (including striped chunks) from a cookie list
func FilterSessionCookies(cookies []*http.Cookie, cookieName string) []*http.Cookie {
	var filtered []*http.Cookie
	for _, cookie := range cookies {
		// skip the base cookie
		if cookie.Name == cookieName {
			continue
		}
		// skip numbered chunks (e.g., "cookieName_1", "cookieName_2", etc.)
		if strings.HasPrefix(cookie.Name, cookieName+"_") {
			continue
		}
		filtered = append(filtered, cookie)
	}
	return filtered
}
