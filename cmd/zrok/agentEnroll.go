package main

import (
	"bufio"
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/agent"
	"github.com/openziti/zrok/environment"
	agent2 "github.com/openziti/zrok/rest_client_zrok/agent"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	agentCmd.AddCommand(newAgentEnrollCommand().cmd)
}

type agentEnrollCommand struct {
	cmd      *cobra.Command
	headless bool
}

func newAgentEnrollCommand() *agentEnrollCommand {
	cmd := &cobra.Command{
		Use:   "enroll",
		Short: "Enroll the agent in remote control",
		Args:  cobra.NoArgs,
	}
	command := &agentEnrollCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Run the agent enrollment in headless mode")
	cmd.Run = command.run
	return command
}

func (cmd *agentEnrollCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	enrlPath, err := root.AgentEnrollment()
	if err != nil {
		tui.Error("error getting agent enrollment path", err)
	}

	_, err = agent.LoadEnrollment(enrlPath)
	if err == nil {
		tui.Error("agent already enrolled; 'zrok agent unenroll' first", nil)
	}

	if !cmd.headless {
		fmt.Println()
		fmt.Println(tui.SeriousBusiness.Render("warning! proceeding will allow remote control of your zrok agent!"))
		fmt.Println()
		fmt.Println("your agent will accept remote commands from:")
		fmt.Println()
		fmt.Println(tui.Attention.Render(root.Environment().ApiEndpoint))
		fmt.Println()
		fmt.Println("you should only proceed if you understand the implications of this action!")
		fmt.Println()
		fmt.Print("to proceed, type 'yes': ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			text := scanner.Text()
			if text != "yes" {
				tui.Error("agent enrollment aborted!", nil)
			}
		}
		fmt.Println()
	}

	zrok, err := root.Client()
	if err != nil {
		tui.Error("error creating zrok api client", err)
	}

	req := agent2.NewEnrollParams()
	req.Body.EnvZID = root.Environment().ZitiIdentity
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)
	resp, err := zrok.Agent.Enroll(req, auth)
	if err != nil {
		tui.Error("error enrolling agent", err)
	}

	enrl := agent.NewEnrollment(resp.Payload.Token)
	if err := enrl.Save(enrlPath); err != nil {
		tui.Error("error saving agent enrollment", err)
	}

	fmt.Printf("agent enrolled with token '%v'\n", enrl.Token)
	if !cmd.headless {
		fmt.Println()
		fmt.Println(tui.SeriousBusiness.Render("restart your zrok agent to enable remote control"))
		fmt.Println()
	}
}
