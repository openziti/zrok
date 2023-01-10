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

func newShareModel(shareToken string, frontendEndpoints []string) *shareModel {
	return &shareModel{
		shareToken:        shareToken,
		frontendEndpoints: frontendEndpoints,
	}
}

func (m *shareModel) Init() tea.Cmd { return nil }

func (m *shareModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		shareHeaderStyle.Width((m.width - 4) / 2)
		m.height = msg.Height
		requestsStyle.Width(m.width - 2)
		requestsStyle.Height(20)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *shareModel) View() string {
	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		shareHeaderStyle.Render(strings.Join(m.frontendEndpoints, "\n")),
		shareHeaderStyle.Render(m.shareToken),
	)
	requests := requestsStyle.Render("hello")
	all := lipgloss.JoinVertical(lipgloss.Left, topRow, requests)
	return all
}

var shareHeaderStyle = lipgloss.NewStyle().
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
