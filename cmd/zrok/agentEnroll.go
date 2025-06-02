package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/agent"
	"github.com/openziti/zrok/environment"
	agent2 "github.com/openziti/zrok/rest_client_zrok/agent"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	agentCmd.AddCommand(newAgentEnrollCommand().cmd)
}

type agentEnrollCommand struct {
	cmd *cobra.Command
}

func newAgentEnrollCommand() *agentEnrollCommand {
	cmd := &cobra.Command{
		Use:   "enroll",
		Short: "Enroll the agent in remote control",
		Args:  cobra.NoArgs,
	}
	command := &agentEnrollCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *agentEnrollCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	enrlPath, err := root.AgentEnrollment()
	if err != nil {
		tui.Error("error getting agent enrollment path", err)
	}

	_, err = agent.LoadEnrollment(enrlPath)
	if err == nil {
		tui.Error("agent already enrolled; 'zrok agent unenroll' first", nil)
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
}
