package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/agent"
	"github.com/openziti/zrok/environment"
	agent2 "github.com/openziti/zrok/rest_client_zrok/agent"
	"github.com/openziti/zrok/tui"
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

	enrlPath, err := root.AgentEnrollment()
	if err != nil {
		tui.Error("error getting agent enrollment path", err)
	}

	_, err = agent.LoadEnrollment(enrlPath)
	if err != nil {
		tui.Warning("error loading agent enrollment; use 'zrok agent enroll' to enroll", err)
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
		tui.Error("error unenrolling agent", err)
	}

	if err := os.Remove(enrlPath); err != nil {
		tui.Error("error removing agent enrollment", err)
	}
}
