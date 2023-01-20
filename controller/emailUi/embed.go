package emailUi

import "embed"

//go:embed verify.gohtml verify.gotext forgotPassword.gohtml forgotPassword.gotext
var FS embed.FS
