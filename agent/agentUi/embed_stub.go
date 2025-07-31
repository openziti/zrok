//go:build no_zrok_ui

package agentUi

import "embed"

// FS is a stub embed.FS that contains no files when built with the no_zrok_ui tag
var FS embed.FS
