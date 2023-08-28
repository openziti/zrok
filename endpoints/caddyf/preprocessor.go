package caddyf

import (
	"github.com/openziti/zrok/sdk"
	"os"
	"strings"
	"text/template"
)

func preprocessCaddyfile(inF string, shr *sdk.Share) (string, error) {
	input, err := os.ReadFile(inF)
	if err != nil {
		return "", err
	}
	tmpl, err := template.New(inF).Parse(string(input))
	if err != nil {
		return "", err
	}
	output := new(strings.Builder)
	if err := tmpl.Execute(output, shr); err != nil {
		return "", err
	}
	return output.String(), nil
}
