package main

import (
	"context"
	"github.com/openziti/zrok/util/sync/driveClient"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
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
		Args:  cobra.ExactArgs(3),
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
	ws, err := client.Open(context.Background(), args[1])
	if err != nil {
		panic(err)
	}
	fs, err := os.Create(args[2])
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(fs, ws)
	if err != nil {
		panic(err)
	}
	ws.Close()
	fs.Close()
}
