package sync

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type davTestEntry struct {
	href    string
	isDir   bool
	content string
}

func newDAVTestServer(rootHref string, entries []davTestEntry) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PROPFIND":
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusMultiStatus)
			if r.Header.Get("Depth") == "0" {
				writeDAVMultiStatus(w, []davTestEntry{{href: rootHref, isDir: true}})
				return
			}
			writeDAVMultiStatus(w, append([]davTestEntry{{href: rootHref, isDir: true}}, entries...))

		case "GET":
			for _, entry := range entries {
				if !entry.isDir && r.URL.Path == entry.href {
					_, _ = w.Write([]byte(entry.content))
					return
				}
			}
			http.NotFound(w, r)

		default:
			http.Error(w, "unexpected method", http.StatusInternalServerError)
		}
	}))
}

func writeDAVMultiStatus(w http.ResponseWriter, entries []davTestEntry) {
	_, _ = fmt.Fprint(w, `<?xml version="1.0" encoding="utf-8"?><d:multistatus xmlns:d="DAV:">`)
	for _, entry := range entries {
		if entry.isDir {
			_, _ = fmt.Fprintf(w, `<d:response><d:href>%s</d:href><d:propstat><d:prop><d:resourcetype><d:collection/></d:resourcetype><d:getlastmodified>Mon, 01 Jan 2024 00:00:00 GMT</d:getlastmodified></d:prop><d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response>`, entry.href)
		} else {
			_, _ = fmt.Fprintf(w, `<d:response><d:href>%s</d:href><d:propstat><d:prop><d:resourcetype/><d:getcontentlength>%d</d:getcontentlength><d:getlastmodified>Mon, 01 Jan 2024 00:00:00 GMT</d:getlastmodified></d:prop><d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response>`, entry.href, len(entry.content))
		}
	}
	_, _ = fmt.Fprint(w, `</d:multistatus>`)
}

func TestWebDAVInventoryRejectsUnsafePaths(t *testing.T) {
	tests := []struct {
		name     string
		urlPath  string
		rootHref string
		entry    davTestEntry
	}{
		{
			name:     "parent segment",
			urlPath:  "/",
			rootHref: "/",
			entry:    davTestEntry{href: "/../outside.txt", content: "owned"},
		},
		{
			name:     "encoded parent segment",
			urlPath:  "/",
			rootHref: "/",
			entry:    davTestEntry{href: "/%2e%2e/outside.txt", content: "owned"},
		},
		{
			name:     "nested parent segment",
			urlPath:  "/",
			rootHref: "/",
			entry:    davTestEntry{href: "/a/../outside.txt", content: "owned"},
		},
		{
			name:     "unsafe directory",
			urlPath:  "/",
			rootHref: "/",
			entry:    davTestEntry{href: "/../outside/", isDir: true},
		},
		{
			name:     "outside requested root",
			urlPath:  "/base",
			rootHref: "/base/",
			entry:    davTestEntry{href: "/other/outside.txt", content: "owned"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newDAVTestServer(tt.rootHref, []davTestEntry{tt.entry})
			defer srv.Close()

			u, err := url.Parse(srv.URL + tt.urlPath)
			if err != nil {
				t.Fatal(err)
			}
			src, err := NewWebDAVTarget(&WebDAVTargetConfig{URL: u})
			if err != nil {
				t.Fatal(err)
			}

			base := t.TempDir()
			root := filepath.Join(base, "victim-root")
			dst := NewFilesystemTarget(&FilesystemTargetConfig{Root: root})

			if err := OneWay(src, dst, false); err == nil {
				t.Fatal("OneWay succeeded, want unsafe path error")
			}
			if _, err := os.Stat(filepath.Join(base, "outside.txt")); !os.IsNotExist(err) {
				t.Fatalf("outside file: got %v, want not exist", err)
			}
			if _, err := os.Stat(filepath.Join(base, "outside")); !os.IsNotExist(err) {
				t.Fatalf("outside directory: got %v, want not exist", err)
			}
		})
	}
}

func TestWebDAVInventoryCopiesSafePathsUnderRequestedRoot(t *testing.T) {
	srv := newDAVTestServer("/base/", []davTestEntry{
		{href: "/base/file.txt", content: "safe"},
	})
	defer srv.Close()

	u, err := url.Parse(srv.URL + "/base")
	if err != nil {
		t.Fatal(err)
	}
	src, err := NewWebDAVTarget(&WebDAVTargetConfig{URL: u})
	if err != nil {
		t.Fatal(err)
	}

	base := t.TempDir()
	root := filepath.Join(base, "dest")
	dst := NewFilesystemTarget(&FilesystemTargetConfig{Root: root})
	if err := OneWay(src, dst, false); err != nil {
		t.Fatalf("OneWay failed: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(root, "file.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got := string(b); got != "safe" {
		t.Fatalf("ReadFile: got %q, want %q", got, "safe")
	}
}

func TestFilesystemTargetRejectsTraversalPaths(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	target := NewFilesystemTarget(&FilesystemTargetConfig{Root: root})

	victim := filepath.Join(base, "victim.txt")
	if err := os.WriteFile(victim, []byte("victim"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.txt"), []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := target.WriteStream("/../outside.txt", strings.NewReader("owned"), 0755); err == nil {
		t.Fatal("WriteStream succeeded, want error")
	}
	if err := target.Mkdir("/../outside-dir"); err == nil {
		t.Fatal("Mkdir succeeded, want error")
	}
	if err := target.Rm("/../victim.txt"); err == nil {
		t.Fatal("Rm succeeded, want error")
	}
	if err := target.SetModificationTime("/../victim.txt", time.Unix(123, 0)); err == nil {
		t.Fatal("SetModificationTime succeeded, want error")
	}
	if err := target.Move("/source.txt", "../moved.txt"); err == nil {
		t.Fatal("Move child succeeded, want error")
	}
	if err := target.Move("/", "../moved-root"); err == nil {
		t.Fatal("Move root succeeded, want error")
	}

	if _, err := os.Stat(filepath.Join(base, "outside.txt")); !os.IsNotExist(err) {
		t.Fatalf("outside file: got %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(base, "outside-dir")); !os.IsNotExist(err) {
		t.Fatalf("outside dir: got %v, want not exist", err)
	}
	b, err := os.ReadFile(victim)
	if err != nil {
		t.Fatalf("victim retained: %v", err)
	}
	if got := string(b); got != "victim" {
		t.Fatalf("victim content: got %q, want %q", got, "victim")
	}
	if _, err := os.Stat(filepath.Join(root, "source.txt")); err != nil {
		t.Fatalf("source retained: %v", err)
	}
}

func TestFilesystemTargetMkdirCreatesRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "new-root")
	target := NewFilesystemTarget(&FilesystemTargetConfig{Root: root})

	if err := target.Mkdir("/"); err != nil {
		t.Fatalf("Mkdir root: %v", err)
	}
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		t.Fatalf("root directory: info=%v err=%v, want directory", fi, err)
	}
}

func TestFilesystemTargetRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	target := NewFilesystemTarget(&FilesystemTargetConfig{Root: root})

	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if err := target.WriteStream("/escape/created.txt", strings.NewReader("owned"), 0755); err == nil {
		t.Fatal("WriteStream succeeded, want error")
	}
	if _, err := os.Stat(filepath.Join(outside, "created.txt")); !os.IsNotExist(err) {
		t.Fatalf("outside create: got %v, want not exist", err)
	}

	victim := filepath.Join(outside, "victim.txt")
	if err := os.WriteFile(victim, []byte("victim"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := target.Rm("/escape/victim.txt"); err == nil {
		t.Fatal("Rm succeeded, want error")
	}
	if _, err := os.Stat(victim); err != nil {
		t.Fatalf("victim retained: %v", err)
	}
}

func TestFilesystemTargetAllowsInternalSymlink(t *testing.T) {
	root := t.TempDir()
	actual := filepath.Join(root, "actual")
	if err := os.Mkdir(actual, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("actual", filepath.Join(root, "nested")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	target := NewFilesystemTarget(&FilesystemTargetConfig{Root: root})
	if err := target.WriteStream("/nested/created.txt", strings.NewReader("safe"), 0755); err != nil {
		t.Fatalf("WriteStream: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(actual, "created.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got := string(b); got != "safe" {
		t.Fatalf("ReadFile: got %q, want %q", got, "safe")
	}
}
