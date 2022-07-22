package sql

import "embed"

//go:embed *.sql
var Fs embed.FS
