package util

import "strings"

func ExpandUrlTemplate(token, template string) string {
	return strings.Replace(template, "{name}", token, -1)
}
