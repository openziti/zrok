package main

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	agentShareCmd.AddCommand(newAgentShareReservedCommand().cmd)
}

type agentShareReservedCommand struct {
	overrideEndpoint string
	insecure         bool
	cmd              *cobra.Command
}

func newAgentShareReservedCommand() *agentShareReservedCommand {
	cmd := &cobra.Command{
		Use:   "reserved <token>",
		Short: "Share an existing reserved share in the zrok Agent",
		Args:  cobra.ExactArgs(1),
	}
	command := &agentShareReservedCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.overrideEndpoint, "override-endpoint", "", "Override the stored target endpoint with a replacement")
	cmd.Flags().BoolVar(&command.insecure, "insecure", false, "Enable insecure TLS certificate validation")
	cmd.Run = command.run
	return command
}

func (cmd *agentShareReservedCommand) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to load environment", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer conn.Close()

	shr, err := client.ReservedShare(context.Background(), &agentGrpc.ReservedShareRequest{
		Token:            args[0],
		OverrideEndpoint: cmd.overrideEndpoint,
		Insecure:         cmd.insecure,
	})
	if err != nil {
		tui.Error("error sharing reserved share", err)
	}

	fmt.Println(shr)
}
