package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/openziti/zrok/rest_client_zrok/account"
	"github.com/openziti/zrok/rest_client_zrok/metadata"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util"
	"github.com/openziti/zrok/zrokdir"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newInviteCommand().cmd)
}

type inviteCommand struct {
	cmd *cobra.Command
	tui inviteTui
}

func newInviteCommand() *inviteCommand {
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a new user to zrok",
		Args:  cobra.ExactArgs(0),
	}
	command := &inviteCommand{
		cmd: cmd,
		tui: newInviteTui(),
	}
	cmd.Run = command.run

	return command
}

func (cmd *inviteCommand) run(_ *cobra.Command, _ []string) {
	zrd, err := zrokdir.Load()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	zrok, err := zrd.Client()
	if err != nil {
		if !panicInstead {
			cmd.endpointError(zrd.ApiEndpoint())
			tui.Error("error creating zrok api client", err)
		}
		panic(err)
	}

	md, err := zrok.Metadata.Configuration(metadata.NewConfigurationParams())
	if err != nil {
		tui.Error("unable to get server metadata", err)
	}

	if md != nil {
		cmd.tui.RequireToken(md.GetPayload().RegistrationRequiresToken)
	}

	if _, err := tea.NewProgram(&cmd.tui).Run(); err != nil {
		tui.Error("unable to run interface", err)
		os.Exit(1)
	}
	if cmd.tui.done {
		email := cmd.tui.emailInputs[0].Value()
		token := cmd.tui.tokenInput.Value()

		req := account.NewInviteParams()
		req.Body = &rest_model_zrok.InviteRequest{
			Email: email,
			Token: token,
		}
		_, err = zrok.Account.Invite(req)
		if err != nil {
			cmd.endpointError(zrd.ApiEndpoint())
			tui.Error("error creating invitation", err)
		}

		fmt.Printf("invitation sent to '%v'!\n", email)
	}
}

func (cmd *inviteCommand) endpointError(apiEndpoint, _ string) {
	fmt.Printf("%v\n\n", tui.SeriousBusiness.Render("there was a problem creating an invitation!"))
	fmt.Printf("you are trying to use the zrok service at: %v\n\n", tui.Code.Render(apiEndpoint))
	fmt.Printf("you can change your zrok service endpoint using this command:\n\n")
	fmt.Printf("%v\n\n", tui.Code.Render("$ zrok config set apiEndpoint <newEndpoint>"))
	fmt.Printf("(where newEndpoint is something like: %v)\n\n", tui.Code.Render("https://some.zrok.io"))
}

type inviteTui struct {
	focusIndex   int
	msg          string
	emailInputs  []textinput.Model
	tokenInput   textinput.Model
	cursorMode   textinput.CursorMode
	done         bool
	requireToken bool
	maxIndex     int

	msgOk         string
	msgMismatch   string
	focusedStyle  lipgloss.Style
	blurredStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	cursorStyle   lipgloss.Style
	noStyle       lipgloss.Style
	helpStyle     lipgloss.Style
	focusedButton string
	blurredButton string
}

func newInviteTui() inviteTui {
	m := inviteTui{
		emailInputs: make([]textinput.Model, 2),
		maxIndex:    2,
	}
	m.focusedStyle = tui.Attention.Copy()
	m.blurredStyle = tui.Code.Copy()
	m.errorStyle = tui.SeriousBusiness.Copy()
	m.cursorStyle = m.focusedStyle.Copy()
	m.noStyle = lipgloss.NewStyle()
	m.helpStyle = m.blurredStyle.Copy()
	m.focusedButton = m.focusedStyle.Copy().Render("[_Submit_]")
	m.blurredButton = fmt.Sprintf("[ %v ]", m.blurredStyle.Render("Submit"))
	m.msgOk = m.noStyle.Render("enter and confirm your email address...")
	m.msg = m.msgOk
	m.msgMismatch = m.errorStyle.Render("email is invalid or does not match confirmation...")

	var t textinput.Model
	for i := range m.emailInputs {
		t = textinput.New()
		t.CursorStyle = m.cursorStyle
		t.CharLimit = 96

		switch i {
		case 0:
			t.Placeholder = "Email Address"
			t.Focus()
			t.PromptStyle = m.focusedStyle
			t.TextStyle = m.focusedStyle
		case 1:
			t.Placeholder = "Confirm Email"
		}

		m.emailInputs[i] = t
	}

	m.tokenInput = textinput.New()
	m.tokenInput.CursorStyle = m.cursorStyle
	m.tokenInput.Placeholder = "Token"

	return m
}

func (m inviteTui) Init() tea.Cmd { return textinput.Blink }

func (m *inviteTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == m.maxIndex {
				if util.IsValidEmail(m.emailInputs[0].Value()) && m.emailInputs[0].Value() == m.emailInputs[1].Value() {
					m.done = true
					return m, tea.Quit
				}
				m.msg = m.msgMismatch
				return m, nil
			}

			if s == "up" || s == "shift+tab" {
				m.msg = m.msgOk
				m.focusIndex--
			} else {
				m.msg = m.msgOk
				m.focusIndex++
			}

			if m.focusIndex > m.maxIndex {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = m.maxIndex
			}

			cmds := make([]tea.Cmd, m.maxIndex)
			for i := 0; i <= len(m.emailInputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.emailInputs[i].Focus()
					m.emailInputs[i].PromptStyle = m.focusedStyle
					m.emailInputs[i].TextStyle = m.focusedStyle
					continue
				}
				m.emailInputs[i].Blur()
				m.emailInputs[i].PromptStyle = m.noStyle
				m.emailInputs[i].TextStyle = m.noStyle
			}
			if m.requireToken {
				if m.focusIndex == 2 {
					cmds[2] = m.tokenInput.Focus()
					m.tokenInput.PromptStyle = m.focusedStyle
					m.tokenInput.TextStyle = m.focusedStyle
				} else {
					m.tokenInput.Blur()
					m.tokenInput.PromptStyle = m.noStyle
					m.tokenInput.TextStyle = m.noStyle
				}
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *inviteTui) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, m.maxIndex)
	for i := range m.emailInputs {
		m.emailInputs[i], cmds[i] = m.emailInputs[i].Update(msg)
	}
	if m.requireToken {
		m.tokenInput, cmds[2] = m.tokenInput.Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m inviteTui) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%v\n\n", m.msg))

	for i := 0; i < len(m.emailInputs); i++ {
		b.WriteString(m.emailInputs[i].View())
		b.WriteRune('\n')
	}

	if m.requireToken {
		b.WriteString(m.tokenInput.View())
		b.WriteRune('\n')
	}

	button := &m.blurredButton
	if m.focusIndex == m.maxIndex {
		button = &m.focusedButton
	}
	_, _ = fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

func (m *inviteTui) RequireToken(require bool) {
	m.requireToken = require
	if require {
		m.maxIndex = 3
	} else {
		m.maxIndex = 2
	}
}
