package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/openziti/zrok/endpoints"
	"strings"
	"time"
)

const shareTuiBacklog = 256

type shareModel struct {
	shrToken             string
	frontendDescriptions []string
	shareMode            string
	backendMode          string
	requests             []*endpoints.Request
	log                  []string
	showLog              bool
	width                int
	height               int
	headerHeight         int
	prg                  *tea.Program
}

type shareLogLine string

func newShareModel(shrToken string, frontendEndpoints []string, shareMode, backendMode string) *shareModel {
	return &shareModel{
		shrToken:             shrToken,
		frontendDescriptions: frontendEndpoints,
		shareMode:            shareMode,
		backendMode:          backendMode,
	}
}

func (m *shareModel) Init() tea.Cmd { return nil }

func (m *shareModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *endpoints.Request:
		m.requests = append(m.requests, msg)
		if len(m.requests) > shareTuiBacklog {
			m.requests = m.requests[1:]
		}

	case shareLogLine:
		m.showLog = true
		m.adjustPaneHeights()

		m.log = append(m.log, string(msg))
		if len(m.log) > shareTuiBacklog {
			m.log = m.log[1:]
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		shareHeaderStyle.Width(m.width - 30)
		shareConfigStyle.Width(26)
		shareRequestsStyle.Width(m.width - 2)
		shareLogStyle.Width(m.width - 2)

		m.height = msg.Height
		m.headerHeight = len(m.frontendDescriptions) + 4
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

func (m *shareModel) View() string {
	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		shareHeaderStyle.Render(strings.Join(m.frontendDescriptions, "\n")),
		shareConfigStyle.Render(m.renderConfig()),
	)
	var panes string
	if m.showLog {
		panes = lipgloss.JoinVertical(lipgloss.Left,
			shareRequestsStyle.Render(m.renderRequests()),
			shareLogStyle.Render(m.renderLog()),
		)
	} else {
		panes = shareRequestsStyle.Render(m.renderRequests())
	}
	return lipgloss.JoinVertical(lipgloss.Left, topRow, panes)
}

func (m *shareModel) adjustPaneHeights() {
	if !m.showLog {
		shareRequestsStyle.Height(m.height - m.headerHeight)
	} else {
		splitHeight := m.height - m.headerHeight
		shareRequestsStyle.Height(splitHeight/2 - 1)
		shareLogStyle.Height(splitHeight/2 - 1)
	}
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

func (m *shareModel) renderRequests() string {
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
	maxRows := shareRequestsStyle.GetHeight()
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

func (m *shareModel) renderLog() string {
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
	maxRows := shareLogStyle.GetHeight()
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

func (m *shareModel) Write(p []byte) (n int, err error) {
	in := string(p)
	lines := strings.Split(in, "\n")
	for _, line := range lines {
		cleanLine := strings.ReplaceAll(line, "\n", "")
		if cleanLine != "" && m.prg != nil {
			m.prg.Send(shareLogLine(cleanLine))
		}
	}
	return len(p), nil
}

var shareHeaderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Align(lipgloss.Center)

var shareConfigStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Align(lipgloss.Center)

var shareRequestsStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63"))

var shareLogStyle = lipgloss.NewStyle().
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
