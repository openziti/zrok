package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"time"
)

type shareModel struct {
	shareToken        string
	frontendEndpoints []string
	shareMode         string
	backendMode       string
	requests          []*shareRequestModel
	logMessages       []string
	width             int
	height            int
}

type shareRequestModel struct {
	stamp      time.Time
	remoteAddr string
	verb       string
	path       string
}

func newShareModel(shareToken string, frontendEndpoints []string, shareMode, backendMode string) *shareModel {
	return &shareModel{
		shareToken:        shareToken,
		frontendEndpoints: frontendEndpoints,
		shareMode:         shareMode,
		backendMode:       backendMode,
	}
}

func (m *shareModel) Init() tea.Cmd { return nil }

func (m *shareModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		shareHeaderStyle.Width(m.width - 30)
		configHeaderStyle.Width(26)
		m.height = msg.Height
		requestsStyle.Width(m.width - 2)
		requestsStyle.Height(m.height - (len(m.frontendEndpoints) + 6))

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+l":
			return m, tea.ClearScreen
		}
	}

	return m, nil
}

func (m *shareModel) View() string {
	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		shareHeaderStyle.Render(strings.Join(m.frontendEndpoints, "\n")),
		configHeaderStyle.Render(m.renderConfig()),
	)
	requests := requestsStyle.Render("hello")
	all := lipgloss.JoinVertical(lipgloss.Left, topRow, requests)
	return all
}

func (m *shareModel) renderConfig() string {
	out := "["
	if m.shareMode == "public" {
		out += shareModePublicStyle.Render(strings.ToUpper(m.shareMode))
	} else {
		out += shareModePrivateStyle.Render(strings.ToUpper(m.shareMode))
	}
	out += "] ["
	if m.backendMode == "proxy" {
		out += backendModeProxyStyle.Render(strings.ToUpper(m.backendMode))
	} else {
		out += backendModeWebStyle.Render(strings.ToUpper(m.backendMode))
	}
	out += "]"
	return out
}

var shareHeaderStyle = lipgloss.NewStyle().
	Height(3).
	PaddingTop(1).PaddingLeft(2).PaddingBottom(1).PaddingRight(2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Align(lipgloss.Center)

var configHeaderStyle = lipgloss.NewStyle().
	Height(3).
	PaddingTop(1).PaddingLeft(2).PaddingBottom(1).PaddingRight(2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Align(lipgloss.Center)

var requestsStyle = lipgloss.NewStyle().
	Height(3).
	PaddingTop(1).PaddingLeft(2).PaddingBottom(1).PaddingRight(2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63"))

var shareModePublicStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0F0"))
var shareModePrivateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F00"))
var backendModeProxyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
var backendModeWebStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0CC"))
