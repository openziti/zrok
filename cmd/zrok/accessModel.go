package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/openziti-test-kitchen/zrok/endpoints"
	"strings"
	"time"
)

type accessModel struct {
	shrToken      string
	localEndpoint string
	requests      []*endpoints.BackendRequest
	width         int
	height        int
}

func newAccessModel(shrToken, localEndpoint string) *accessModel {
	return &accessModel{
		shrToken:      shrToken,
		localEndpoint: localEndpoint,
	}
}

func (m *accessModel) Init() tea.Cmd { return nil }

func (m *accessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *endpoints.BackendRequest:
		m.requests = append([]*endpoints.BackendRequest{msg}, m.requests...)
		if len(m.requests) > 2048 {
			m.requests = m.requests[:2048]
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		accessHeaderStyle.Width(m.width - 2)
		accessRequestsStyle.Width(m.width - 2)

		m.height = msg.Height
		accessRequestsStyle.Height(m.height - 7)

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

func (m *accessModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		accessHeaderStyle.Render(fmt.Sprintf("%v -> %v", m.localEndpoint, m.shrToken)),
		accessRequestsStyle.Render(m.renderRequests()),
	)
}

func (m *accessModel) renderRequests() string {
	out := ""
	maxRows := accessRequestsStyle.GetHeight()
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

func (m *accessModel) renderMethod(method string) string {
	switch strings.ToLower(method) {
	case "get":
		return getStyle.Render(method)
	case "post":
		return postStyle.Render(method)
	default:
		return otherMethodStyle.Render(method)
	}
}

var accessHeaderStyle = lipgloss.NewStyle().
	Height(3).
	PaddingTop(1).PaddingLeft(2).PaddingBottom(1).PaddingRight(2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Align(lipgloss.Center)

var accessRequestsStyle = lipgloss.NewStyle().
	Height(3).
	PaddingLeft(2).PaddingRight(2).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63"))

type accessLogWriter struct{}

func (w accessLogWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
