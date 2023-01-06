package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"os"
	"strings"
)

type backendHandler interface {
	Requests() func() int32
}

func mustGetAdminAuth() runtime.ClientAuthInfoWriter {
	adminToken := os.Getenv("ZROK_ADMIN_TOKEN")
	if adminToken == "" {
		panic("please set ZROK_ADMIN_TOKEN to a valid admin token for your zrok instance")
	}
	return httptransport.APIKeyAuth("X-TOKEN", "header", adminToken)
}

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D90166"))
var errorLabel = errorStyle.Render("ERROR")
var warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
var warningLabel = warningStyle.Render("WARNING")
var codeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))

func showError(msg string, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v: %v (%v)\n", errorLabel, msg, strings.TrimSpace(err.Error()))
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "%v %v\n", errorLabel, msg)
	}
	os.Exit(1)
}
