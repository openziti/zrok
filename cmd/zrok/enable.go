package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/environment"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/tui"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/spf13/cobra"
	"os"
	user2 "os/user"
	"time"
)

func init() {
	rootCmd.AddCommand(newEnableCommand().cmd)
}

type enableCommand struct {
	description string
	cmd         *cobra.Command
}

func newEnableCommand() *enableCommand {
	cmd := &cobra.Command{
		Use:   "enable <token>",
		Short: "Enable an environment for zrok",
		Args:  cobra.ExactArgs(1),
	}
	command := &enableCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.description, "description", "d", "<user>@<hostname>", "Description of this environment")
	cmd.Run = command.run
	return command
}

func (cmd *enableCommand) run(_ *cobra.Command, args []string) {
	zrd, err := zrokdir.Load()
	if err != nil {
		panic(err)
	}
	token := args[0]

	hostName, hostDetail, err := getHost()
	if err != nil {
		panic(err)
	}
	user, err := user2.Current()
	if err != nil {
		panic(err)
	}
	hostDetail = fmt.Sprintf("%v; %v", user.Username, hostDetail)
	if cmd.description == "<user>@<hostname>" {
		cmd.description = fmt.Sprintf("%v@%v", user.Username, hostName)
	}
	zrok, err := zrd.Client()
	if err != nil {
		cmd.endpointError(zrd.ApiEndpoint())
		tui.Error("error creating service client", err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", token)
	req := environment.NewEnableParams()
	req.Body = &rest_model_zrok.EnableRequest{
		Description: cmd.description,
		Host:        hostDetail,
	}

	var prg *tea.Program
	var mdl enableTuiModel
	var done = make(chan struct{})
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

	resp, err := zrok.Environment.Enable(req, auth)
	if err != nil {
		time.Sleep(250 * time.Millisecond)
		prg.Send(fmt.Sprintf("the zrok service returned an error: %v\n", err))
		prg.Quit()
		<-done
		cmd.endpointError(zrd.ApiEndpoint())
		os.Exit(1)
	}
	prg.Send("writing the environment details...")
	apiEndpoint, _ := zrd.ApiEndpoint()
	zrd.Env = &zrokdir.Environment{Token: token, ZId: resp.Payload.Identity, ApiEndpoint: apiEndpoint}
	if err := zrd.Save(); err != nil {
		prg.Send(fmt.Sprintf("there was an error saving the new environment: %v", err))
		prg.Quit()
		<-done
		os.Exit(1)
	}
	if err := zrokdir.SaveZitiIdentity("backend", resp.Payload.Cfg); err != nil {
		prg.Send(fmt.Sprintf("there was an error writing the environment: %v", err))
		prg.Quit()
		<-done
		os.Exit(1)
	}

	prg.Send(fmt.Sprintf("the zrok environment was successfully enabled..."))
	prg.Quit()
	<-done
}

func (cmd *enableCommand) endpointError(apiEndpoint, _ string) {
	fmt.Printf("%v\n\n", tui.SeriousBusiness.Render("there was a problem enabling your environment!"))
	fmt.Printf("you are trying to use the zrok service at: %v\n\n", tui.Code.Render(apiEndpoint))
	fmt.Printf("you can change your zrok service endpoint using this command:\n\n")
	fmt.Printf("%v\n\n", tui.Code.Render("$ zrok config set apiEndpoint <newEndpoint>"))
	fmt.Printf("(where newEndpoint is something like: %v)\n\n", tui.Code.Render("https://some.zrok.io"))
}

func getHost() (string, string, error) {
	info, err := host.Info()
	if err != nil {
		return "", "", err
	}
	thisHost := fmt.Sprintf("%v; %v; %v; %v; %v; %v; %v",
		info.Hostname, info.OS, info.Platform, info.PlatformFamily, info.PlatformVersion, info.KernelVersion, info.KernelArch)
	return info.Hostname, thisHost, nil
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
