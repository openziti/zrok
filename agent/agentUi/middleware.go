package agentUi

import (
	"github.com/sirupsen/logrus"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const staticPath = "dist"

func Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1") {
			handler.ServeHTTP(w, r)
			return
		}

		path := filepath.ToSlash(filepath.Join(staticPath, r.URL.Path))
		logrus.Debugf("path = %v", path)

		f, err := FS.Open(path)
		if os.IsNotExist(err) {
			// file does not exist, serve index.gohtml
			index, err := FS.ReadFile(filepath.ToSlash(filepath.Join(staticPath, "index.html")))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write(index)
			return

		} else if err != nil {
			// if we got an error (that wasn't that the file doesn't exist) stating the
			// file, return a 500 internal server error and stop
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = f.Close() }()

		// get the subdirectory of the static dir
		if statics, err := fs.Sub(FS, staticPath); err == nil {
			// otherwise, use http.FileServer to serve the static dir
			http.FileServer(http.FS(statics)).ServeHTTP(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
