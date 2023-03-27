package emailUi

import "embed"

//go:embed verify.gohtml verify.gotext resetPassword.gohtml resetPassword.gotext limitWarning.gohtml limitWarning.gotext
var FS embed.FS
