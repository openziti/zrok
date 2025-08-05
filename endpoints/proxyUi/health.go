package proxyUi

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func WriteHealthOk(w http.ResponseWriter) {
	if data, err := FS.ReadFile("health.html"); err == nil {
		w.WriteHeader(http.StatusOK)
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
