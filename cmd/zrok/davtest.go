package main

import (
	"context"
	"github.com/openziti/zrok/util/sync/driveClient"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
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
		Args:  cobra.ExactArgs(1),
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
	fis, err := client.Readdir(context.Background(), "/", true)
	if err != nil {
		panic(err)
	}
	for _, fi := range fis {
		logrus.Infof("=> %s", fi.Path)
	}
}
