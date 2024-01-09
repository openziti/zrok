package main

import (
	"context"
	"github.com/openziti/zrok/util/sync/driveClient"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

func init() {
	rootCmd.AddCommand(newDavtestCommand().cmd)
}

type davtestCommand struct {
	cmd *cobra.Command
}

func newDavtestCommand() *davtestCommand {
	cmd := &cobra.Command{
		Use:   "davtest",
		Short: "WebDAV testing wrapper",
		Args:  cobra.ExactArgs(2),
	}
	command := &davtestCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *davtestCommand) run(_ *cobra.Command, args []string) {
	client, err := driveClient.NewClient(http.DefaultClient, args[0])
	if err != nil {
		panic(err)
	}
	if err := client.Touch(context.Background(), args[1], time.Now().Add(-(24 * time.Hour))); err != nil {
		panic(err)
	}
}
