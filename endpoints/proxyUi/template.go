package proxyUi

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var tmpl *template.Template

func init() {
	if data, err := FS.ReadFile("template.html"); err == nil {
		tmpl, err = template.New("template").Parse(string(data))
		if err != nil {
			panic(errors.Wrap(err, "unable to parse embedded template 'template.html'"))
		}
	} else {
		panic(errors.Wrap(err, "unable to load embedded template 'template.html'"))
	}
}

type VariableData map[string]interface{}

func ReplaceTemplate(path string) error {
	if f, err := os.ReadFile(path); err == nil {
		tmpl, err = template.New("template").Parse(string(f))
		if err != nil {
			panic(errors.Wrapf(err, "unable to parse template '%v'", path))
		}
	} else {
		return errors.Wrapf(err, "error reading template from '%v'", path)
	}
	return nil
}

func WriteHealthOk(w http.ResponseWriter) {
	WriteTemplate(w, http.StatusOK, RequiredData("healthy", "healthy"))
}

func WriteBadGateway(w http.ResponseWriter, variableData VariableData) {
	WriteTemplate(w, http.StatusBadGateway, variableData)
}

func RequiredData(title, banner string) VariableData {
	return map[string]interface{}{
		"title":  title,
		"banner": banner,
	}
}

func (vd VariableData) WithError(err error) VariableData {
	vd["error"] = err.Error()
	return vd
}

func NotFoundData(shareToken string) VariableData {
	return RequiredData(
		fmt.Sprintf("'%v' not found!", shareToken),
		fmt.Sprintf("share <code>%v</code> not found!", shareToken),
	)
}

func WriteNotFound(w http.ResponseWriter, variableData VariableData) {
	WriteTemplate(w, http.StatusNotFound, variableData)
}

func UnauthorizedData() VariableData {
	return RequiredData(
		"unauthorized!",
		"user not authorized!",
	)
}

func UnauthorizedUser(user string) VariableData {
	return RequiredData(
		"unauthorized!",
		fmt.Sprintf("user <code>%v</code> not authorized to access share!", user),
	)
}

func WriteUnauthorized(w http.ResponseWriter, variableData VariableData) {
	WriteTemplate(w, http.StatusUnauthorized, variableData)
}

func WriteTemplate(w http.ResponseWriter, statusCode int, variableData VariableData) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variableData); err != nil {
		logrus.Errorf("failed to execute template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

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
