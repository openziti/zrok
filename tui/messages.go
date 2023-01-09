package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
)

var ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D90166"))
var ErrorLabel = ErrorStyle.Render("ERROR")
var WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
var WarningLabel = WarningStyle.Render("WARNING")
var CodeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))

func Error(msg string, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v: %v (%v)\n", ErrorLabel, msg, strings.TrimSpace(err.Error()))
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "%v %v\n", ErrorLabel, msg)
	}
	os.Exit(1)
}

func Warning(msg string, v ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, "%v: "+msg+"\n", append([]interface{}{WarningLabel}, v...))
}
