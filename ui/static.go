package ui

import (
	"github.com/sirupsen/logrus"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func StaticBuilder(handler http.Handler) http.Handler {
	logrus.Infof("building")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1") {
			logrus.Debugf("directing '%v' to api handler", r.URL.Path)
			handler.ServeHTTP(w, r)
			return
		}

		logrus.Debugf("directing '%v' to static handler", r.URL.Path)

		staticPath := "build"
		indexPath := "index.html"

		// get the absolute path to prevent directory traversal
		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			// if we failed to get the absolute path respond with a 400 bad request and stop
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// prepend the path with the path to the static directory
		path = filepath.Join(staticPath, path)

		_, err = FS.Open(path)
		if os.IsNotExist(err) {
			// file does not exist, serve index.gohtml
			index, err := FS.ReadFile(filepath.Join(staticPath, indexPath))
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

		// get the subdirectory of the static dir
		if statics, err := fs.Sub(FS, staticPath); err == nil {
			// otherwise, use http.FileServer to serve the static dir
			http.FileServer(http.FS(statics)).ServeHTTP(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
