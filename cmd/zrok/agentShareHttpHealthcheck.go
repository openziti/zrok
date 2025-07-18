package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	agentShareCmd.AddCommand(newAgentShareHttpHealthcheckCommand().cmd)
}

type agentShareHttpHealthcheckCommand struct {
	cmd     *cobra.Command
	timeout time.Duration
}

func newAgentShareHttpHealthcheckCommand() *agentShareHttpHealthcheckCommand {
	cmd := &cobra.Command{
		Use:     "http-healthcheck <shareToken> <httpVerb> <healthcheckEndpoint> <expectedHttpStatus>",
		Aliases: []string{"health"},
		Short:   "Perform a share target healthcheck for 'proxy' shares in the agent",
		Args:    cobra.ExactArgs(4),
	}
	command := &agentShareHttpHealthcheckCommand{cmd: cmd}
	cmd.Flags().DurationVar(&command.timeout, "timeout", 6*time.Second, "Timeout for healthcheck request")
	cmd.Run = command.run
	return command
}

func (cmd *agentShareHttpHealthcheckCommand) run(_ *cobra.Command, args []string) {
	expectedHttpStatus, err := strconv.Atoi(args[3])
	if err != nil {
		tui.Error(fmt.Sprintf("'%v' is not a valid HTTP status", args[3]), err)
	}

	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("unable to load environment", err)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("unable to connect to agent", err)
	}
	defer conn.Close()

	resp, err := client.ShareHttpHealthcheck(context.Background(), &agentGrpc.ShareHttpHealthcheckRequest{
		Token:                args[0],
		HttpVerb:             args[1],
		Endpoint:             args[2],
		ExpectedHttpResponse: uint32(expectedHttpStatus),
		TimeoutMs:            uint64(cmd.timeout.Milliseconds()),
	})
	if err != nil {
		tui.Error("error performing healthcheck", err)
	}

	if resp.Healthy {
		fmt.Println("healthy")
	} else {
		fmt.Printf("unhealthy; %v\n", resp.Error)
		os.Exit(1)
	}
}
