package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
)

var SeriousBusiness = lipgloss.NewStyle().Foreground(lipgloss.Color("#D90166"))
var ErrorLabel = SeriousBusiness.Render("ERROR")
var Attention = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
var WarningLabel = Attention.Render("WARNING")
var Code = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))

func Error(msg string, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[%v]: %v (%v)\n", ErrorLabel, msg, strings.TrimSpace(err.Error()))
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "[%v] %v\n", ErrorLabel, msg)
	}
	os.Exit(1)
}

func Warning(msg string, v ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "%v: "+msg+"\n", append([]interface{}{WarningLabel}, v...))
}
