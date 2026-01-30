package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/v2/agent"
	"github.com/openziti/zrok/v2/environment"
	agent2 "github.com/openziti/zrok/v2/rest_client_zrok/agent"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	agentCmd.AddCommand(newAgentUnenrollCommand().cmd)
}

type agentUnenrollCommand struct {
	cmd *cobra.Command
}

func newAgentUnenrollCommand() *agentUnenrollCommand {
	cmd := &cobra.Command{
		Use:   "unenroll",
		Short: "Unenroll the agent from remote management",
		Args:  cobra.NoArgs,
	}
	command := &agentUnenrollCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentUnenrollCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok2 enable'?", nil)
	}

	enrlPath, err := root.AgentEnrollment()
	if err != nil {
		tui.Error("error getting agent enrollment path", err)
	}

	if _, err := os.Stat(enrlPath); os.IsNotExist(err) {
		tui.Error("agent not enrolled; use 'zrok2 agent enroll' to enroll", nil)
	}

	_, err = agent.LoadEnrollment(enrlPath)
	if err != nil {
		tui.Warning("error loading agent enrollment; use 'zrok2 agent enroll' to enroll", err)
	}

	zrok, err := root.Client()
	if err != nil {
		tui.Error("error creating zrok api client", err)
	}

	req := agent2.NewUnenrollParams()
	req.Body.EnvZID = root.Environment().ZitiIdentity
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)
	_, err = zrok.Agent.Unenroll(req, auth)
	if err != nil {
		fmt.Printf("%v: error unenrolling agent remote from '%v'; ignoring\n", tui.Attention.Render("WARNING"), root.Environment().ApiEndpoint)
	} else {
		fmt.Printf("%v: unenrolled agent from '%v'\n", tui.SeriousBusiness.Render("SUCCESS"), root.Environment().ApiEndpoint)
	}

	if err := os.Remove(enrlPath); err != nil {
		tui.Error("error removing agent enrollment", err)
	} else {
		fmt.Printf("%v: removed agent-enrollment.json\n", tui.SeriousBusiness.Render("SUCCESS"))
	}
}
