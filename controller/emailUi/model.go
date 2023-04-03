package emailUi

import (
	"bytes"
	"github.com/pkg/errors"
	"text/template"
)

type WarningEmail struct {
	EmailAddress string
	Detail       string
	Version      string
}

func (we WarningEmail) MergeTemplate(filename string) (string, error) {
	t, err := template.ParseFS(FS, filename)
	if err != nil {
		return "", errors.Wrapf(err, "error parsing warning email template '%v'", filename)
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, we); err != nil {
		return "", errors.Wrapf(err, "error executing warning email template '%v'", filename)
	}
	return buf.String(), nil
}
