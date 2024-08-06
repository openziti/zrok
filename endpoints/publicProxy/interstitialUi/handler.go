package interstitialUi

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var externalFile []byte

func WriteInterstitialAnnounce(w http.ResponseWriter, htmlPath string) {
	if htmlPath != "" && externalFile == nil {
		if data, err := os.ReadFile(htmlPath); err == nil {
			externalFile = data
		} else {
			logrus.Errorf("error reading external interstitial file '%v': %v", htmlPath, err)
		}
	}
	var htmlData = externalFile
	if htmlData == nil {
		if data, err := FS.ReadFile("index.html"); err == nil {
			htmlData = data
		} else {
			logrus.Errorf("error reading embedded interstitial html 'index.html': %v", err)
		}
	}
	w.WriteHeader(http.StatusOK)
	n, err := w.Write(htmlData)
	if n != len(htmlData) {
		logrus.Errorf("short write")
		return
	}
	if err != nil {
		logrus.Error(err)
		return
	}
}
