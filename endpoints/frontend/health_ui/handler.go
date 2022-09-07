package health_ui

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

func WriteHealthOk(w http.ResponseWriter) {
	if data, err := FS.ReadFile("index.html"); err == nil {
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
