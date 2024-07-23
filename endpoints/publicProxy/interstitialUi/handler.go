package interstitialUi

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

func WriteInterstitialAnnounce(w http.ResponseWriter) {
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
