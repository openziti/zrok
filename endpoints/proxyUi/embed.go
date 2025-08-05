package proxyUi

import "embed"

//go:embed health.html intersititial.html notFound.html unauthorized.html
var FS embed.FS
