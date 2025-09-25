package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	restEnvironment "github.com/openziti/zrok/rest_client_zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newEnableCommand().cmd)
}

type enableCommand struct {
	description string
	headless    bool
	cmd         *cobra.Command
}

func newEnableCommand() *enableCommand {
	cmd := &cobra.Command{
		Use:   "enable <token>",
		Short: "Enable an environment for zrok",
		Args:  cobra.ExactArgs(1),
	}
	command := &enableCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
	cmd.Flags().StringVarP(&command.description, "description", "d", "<user>@<hostname>", "Description of this environment")
	cmd.Run = command.run
	return command
}

func (cmd *enableCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}
	token := args[0]

	if env.IsEnabled() {
		tui.Error(fmt.Sprintf("you already have an enabled environment, %v first before you %v", tui.Code.Render("zrok disable"), tui.Code.Render("zrok enable")), nil)
	}

	hostName, hostDetail, username, err := util.GetHostDetails()
	if err != nil {
		panic(err)
	}
	hostDetail, cmd.description = util.FormatHostDetailsWithUser(username, hostName, hostDetail, cmd.description)

	zrok, err := env.Client()
	if err != nil {
		cmd.endpointError(env.ApiEndpoint())
		tui.Error("error creating service client", err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", token)
	req := restEnvironment.NewEnableParams()
	req.Body.Description = cmd.description
	req.Body.Host = hostDetail

	var prg *tea.Program
	var done = make(chan struct{})
	if !cmd.headless {
		var mdl enableTuiModel
		go func() {
			mdl = newEnableTuiModel()
			mdl.msg = "contacting the zrok service..."
			prg = tea.NewProgram(mdl)
			if _, err := prg.Run(); err != nil {
				fmt.Println(err)
			}
			close(done)
			if mdl.quitting {
				os.Exit(1)
			}
		}()
	} else {
		dl.Infof("contacting the zrok service...")
	}

	resp, err := zrok.Environment.Enable(req, auth)
	//Switch on err type (401, 400, 500, etc...)
	if err != nil {
		time.Sleep(250 * time.Millisecond)
		if !cmd.headless && prg != nil {
			prg.Send(fmt.Sprintf("the zrok service returned an error: %v\n", err))
			prg.Quit()
		} else {
			dl.Errorf("the zrok service returned an error: %v", err)
		}
		select {
		case <-done:
		case <-time.After(1 * time.Second):
		}
		cmd.endpointError(env.ApiEndpoint())
		os.Exit(1)
	}
	if err != nil {
		prg.Send("writing the environment details...")
	}
	apiEndpoint, _ := env.ApiEndpoint()
	if err := env.SetEnvironment(&env_core.Environment{AccountToken: token, ZitiIdentity: resp.Payload.Identity, ApiEndpoint: apiEndpoint}); err != nil {
		if !cmd.headless && prg != nil {
			prg.Send(fmt.Sprintf("there was an error saving the new environment: %v", err))
			prg.Quit()
		} else {
			dl.Errorf("there was an error saving the new environment: %v", err)
		}
		select {
		case <-done:
		case <-time.After(1 * time.Second):
		}
		os.Exit(1)
	}
	if err := env.SaveZitiIdentityNamed(env.EnvironmentIdentityName(), resp.Payload.Cfg); err != nil {
		if !cmd.headless && prg != nil {
			prg.Send(fmt.Sprintf("there was an error writing the environment: %v", err))
			prg.Quit()
		} else {
			dl.Errorf("there was an error writing the environment: %v", err)
		}
		select {
		case <-done:
		case <-time.After(1 * time.Second):
		}
		os.Exit(1)
	}

	if !cmd.headless && prg != nil {
		prg.Send(fmt.Sprintf("the zrok environment was successfully enabled..."))
		prg.Quit()
	} else {
		dl.Infof("the zrok environment was successfully enabled...")
	}
	select {
	case <-done:
	case <-time.After(1 * time.Second):
	}
}

func (cmd *enableCommand) endpointError(apiEndpoint, _ string) {
	fmt.Printf("%v\n\n", tui.SeriousBusiness.Render("there was a problem enabling your environment!"))
	fmt.Printf("you are trying to use the zrok service at: %v\n\n", tui.Code.Render(apiEndpoint))
	fmt.Printf("you can change your zrok service endpoint using this command:\n\n")
	fmt.Printf("%v\n\n", tui.Code.Render("$ zrok config set apiEndpoint <newEndpoint>"))
	fmt.Printf("(where newEndpoint is something like: %v)\n\n", tui.Code.Render("https://some.zrok.io"))
}

type enableTuiModel struct {
	spinner  spinner.Model
	msg      string
	quitting bool
}

func newEnableTuiModel() enableTuiModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = tui.Attention
	return enableTuiModel{spinner: s}
}

func (m enableTuiModel) Init() tea.Cmd { return m.spinner.Tick }

func (m enableTuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case string:
		m.msg = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		default:
			return m, nil
		}

	case struct{}:
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m enableTuiModel) View() string {
	str := fmt.Sprintf("%s %s\n", m.spinner.View(), m.msg)
	if m.quitting {
		return str
	}
	return str
}
