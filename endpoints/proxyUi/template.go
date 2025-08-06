package proxyUi

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/sirupsen/logrus"
)

var externalTemplate []byte

func WriteBadGateway(w http.ResponseWriter, variableData map[string]interface{}, templatePath string) {
	WriteTemplate(w, http.StatusBadGateway, variableData, templatePath)
}

func TemplateData(title, banner string) map[string]interface{} {
	return map[string]interface{}{
		"title":  title,
		"banner": banner,
	}
}

func NotFoundData(shareToken string) map[string]interface{} {
	return TemplateData(
		fmt.Sprintf("'%v' not found!", shareToken),
		fmt.Sprintf("share <code>%v</code> not found!", shareToken),
	)
}

func WriteNotFound(w http.ResponseWriter, variableData map[string]interface{}, templatePath string) {
	WriteTemplate(w, http.StatusNotFound, variableData, templatePath)
}

func WriteTemplate(w http.ResponseWriter, statusCode int, variableData map[string]interface{}, templatePath string) {
	if templatePath != "" && externalTemplate == nil {
		if f, err := os.ReadFile(templatePath); err == nil {
			externalTemplate = f
		} else {
			logrus.Errorf("error reading proxyUi template from '%v': %v", templatePath, err)
		}
	}
	var templateData = externalTemplate
	if templateData == nil {
		if f, err := FS.ReadFile("template.html"); err == nil {
			templateData = f
		} else {
			logrus.Errorf("error reading embedded proxyUi template 'template.html': %v", err)
		}
	}

	tmpl, err := template.New("template").Parse(string(templateData))
	if err != nil {
		logrus.Errorf("failed to parse template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Warnf("merging variable data: %v", variableData)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variableData); err != nil {
		logrus.Errorf("failed to execute template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	n, err := w.Write(buf.Bytes())
	if n != buf.Len() {
		logrus.Errorf("short write: wrote %d bytes, expected %d", n, buf.Len())
		return
	}
	if err != nil {
		logrus.Errorf("failed to write response: %v", err)
		return
	}
}
