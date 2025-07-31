//go:build !no_zrok_ui

package ui

import "embed"

//go:embed dist
var FS embed.FS
