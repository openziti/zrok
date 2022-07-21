package util

import (
	"fmt"
	"net/http"
	"sort"
)

func DumpHeaders(headers http.Header, in bool) string {
	out := "headers {\n"
	keys := make([]string, len(headers))
	i := 0
	for k, _ := range headers {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range headers[k] {
			indicator := "->"
			if !in {
				indicator = "<-"
			}
			out += fmt.Sprintf("\t%v %v: %v\n", indicator, k, v)
		}
	}
	out += "}"
	return out
}
