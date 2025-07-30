package proxyUi

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func WriteNotFound(w http.ResponseWriter) {
	if data, err := FS.ReadFile("notFound.html"); err == nil {
		w.WriteHeader(http.StatusNotFound)
		n, err := w.Write(data)
		if n != len(data) {
			logrus.Errorf("short write")
			return
		}
		if err != nil {
			logrus.Error(err)
			return
		}
	}
}
