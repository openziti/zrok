package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/openziti-test-kitchen/zrok/endpoints"
	"strings"
	"time"
)

type shareModel struct {
	shareToken           string
	frontendDescriptions []string
	shareMode            string
	backendMode          string
	requests             []*endpoints.BackendRequest
	logMessages          []string
	width                int
	height               int
}

func newShareModel(shareToken string, frontendEndpoints []string, shareMode, backendMode string) *shareModel {
	return &shareModel{
		shareToken:           shareToken,
		frontendDescriptions: frontendEndpoints,
		shareMode:            shareMode,
		backendMode:          backendMode,
	}
}

func (m *shareModel) Init() tea.Cmd { return nil }

func (m *shareModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *endpoints.BackendRequest:
		m.requests = append([]*endpoints.BackendRequest{msg}, m.requests...)
		if len(m.requests) > 2048 {
			m.requests = m.requests[:2048]
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		shareHeaderStyle.Width(m.width - 30)
		configHeaderStyle.Width(26)
		m.height = msg.Height
		requestsStyle.Width(m.width - 2)
		requestsStyle.Height(m.height - (len(m.frontendDescriptions) + 6))

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
		shareHeaderStyle.Render(strings.Join(m.frontendDescriptions, "\n")),
		configHeaderStyle.Render(m.renderConfig()),
	)
	requests := requestsStyle.Render(m.renderBackendRequests())
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

func (m *shareModel) renderBackendRequests() string {
	out := ""
	maxRows := requestsStyle.GetHeight()
	for i := 0; i < maxRows && i < len(m.requests); i++ {
		req := m.requests[i]
		out += fmt.Sprintf("%v %v -> %v %v",
			timeStyle.Render(req.Stamp.Format(time.RFC850)),
			addressStyle.Render(req.RemoteAddr),
			m.renderMethod(req.Method),
			req.Path,
		)
		if i != maxRows-1 {
			out += "\n"
		}
	}
	return out
}

func (m *shareModel) renderMethod(method string) string {
	switch strings.ToLower(method) {
	case "get":
		return getStyle.Render(method)
	case "post":
		return postStyle.Render(method)
	default:
		return otherMethodStyle.Render(method)
	}
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
	PaddingLeft(2).PaddingRight(2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63"))

var shareModePublicStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0F0"))
var shareModePrivateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F00"))
var backendModeProxyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
var backendModeWebStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0CC"))
var timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#444"))
var addressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
var getStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("98"))
var postStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("101"))
var otherMethodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("166"))
