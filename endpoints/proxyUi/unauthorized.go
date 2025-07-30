package proxyUi

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func WriteUnauthorized(w http.ResponseWriter) {
	if data, err := FS.ReadFile("unauthorized.html"); err == nil {
		w.WriteHeader(http.StatusUnauthorized)
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
