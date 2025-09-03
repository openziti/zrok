package util

import "strings"

func ExpandUrlTemplate(token, template string) string {
	return strings.Replace(template, "{token}", token, -1)
}
