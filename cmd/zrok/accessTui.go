package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/openziti-test-kitchen/zrok/endpoints"
	"strings"
	"time"
)

const accessTuiBacklog = 256

type accessModel struct {
	shrToken      string
	localEndpoint string
	requests      []*endpoints.Request
	log           []string
	showLog       bool
	width         int
	height        int
	prg           *tea.Program
}

type accessLogLine string

func newAccessModel(shrToken, localEndpoint string) *accessModel {
	return &accessModel{
		shrToken:      shrToken,
		localEndpoint: localEndpoint,
	}
}

func (m *accessModel) Init() tea.Cmd { return nil }

func (m *accessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *endpoints.Request:
		m.requests = append(m.requests, msg)
		if len(m.requests) > accessTuiBacklog {
			m.requests = m.requests[1:]
		}

	case accessLogLine:
		m.showLog = true
		m.adjustPaneHeights()

		m.log = append(m.log, string(msg))
		if len(m.log) > accessTuiBacklog {
			m.log = m.log[1:]
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		accessHeaderStyle.Width(m.width - 2)
		accessRequestsStyle.Width(m.width - 2)
		accessLogStyle.Width(m.width - 2)

		m.height = msg.Height
		m.adjustPaneHeights()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+l":
			return m, tea.ClearScreen
		case "l":
			m.showLog = !m.showLog
			m.adjustPaneHeights()
		}
	}

	return m, nil
}

func (m *accessModel) View() string {
	var panes string
	if m.showLog {
		panes = lipgloss.JoinVertical(lipgloss.Left,
			accessRequestsStyle.Render(m.renderRequests()),
			accessLogStyle.Render(m.renderLog()),
		)
	} else {
		panes = accessRequestsStyle.Render(m.renderRequests())
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		accessHeaderStyle.Render(fmt.Sprintf("%v -> %v", m.localEndpoint, m.shrToken)),
		panes,
	)
}

func (m *accessModel) adjustPaneHeights() {
	if !m.showLog {
		accessRequestsStyle.Height(m.height - 5)
	} else {
		splitHeight := m.height - 5
		accessRequestsStyle.Height(splitHeight/2 - 1)
		accessLogStyle.Height(splitHeight/2 - 1)
	}
}

func (m *accessModel) renderRequests() string {
	var requestLines []string
	for _, req := range m.requests {
		reqLine := fmt.Sprintf("%v %v -> %v %v",
			timeStyle.Render(req.Stamp.Format(time.RFC850)),
			addressStyle.Render(req.RemoteAddr),
			m.renderMethod(req.Method),
			req.Path,
		)
		reqLineWrapped := wordwrap.String(reqLine, m.width-2)
		splitWrapped := strings.Split(reqLineWrapped, "\n")
		for _, splitLine := range splitWrapped {
			splitLine := strings.ReplaceAll(splitLine, "\n", "")
			if splitLine != "" {
				requestLines = append(requestLines, splitLine)
			}
		}
	}
	maxRows := accessRequestsStyle.GetHeight()
	startRow := 0
	if len(requestLines) > maxRows {
		startRow = len(requestLines) - maxRows
	}
	out := ""
	for i := startRow; i < len(requestLines); i++ {
		outLine := requestLines[i]
		if i < len(requestLines)-1 {
			outLine += "\n"
		}
		out += outLine
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

func (m *accessModel) renderLog() string {
	var splitLines []string
	for _, line := range m.log {
		wrapped := wordwrap.String(line, m.width-2)
		wrappedLines := strings.Split(wrapped, "\n")
		for _, wrappedLine := range wrappedLines {
			splitLine := strings.ReplaceAll(wrappedLine, "\n", "")
			if splitLine != "" {
				splitLines = append(splitLines, splitLine)
			}
		}
	}
	maxRows := accessLogStyle.GetHeight()
	startRow := 0
	if len(splitLines) > maxRows {
		startRow = len(splitLines) - maxRows
	}
	out := ""
	for i := startRow; i < len(splitLines); i++ {
		outLine := splitLines[i]
		if i < len(splitLines)-1 {
			outLine += "\n"
		}
		out += outLine
	}
	return out
}

func (m *accessModel) Write(p []byte) (n int, err error) {
	in := string(p)
	lines := strings.Split(in, "\n")
	for _, line := range lines {
		cleanLine := strings.ReplaceAll(line, "\n", "")
		if cleanLine != "" {
			m.prg.Send(accessLogLine(cleanLine))
		}
	}
	return len(p), nil
}

var accessHeaderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Align(lipgloss.Center)

var accessRequestsStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63"))

var accessLogStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63"))
