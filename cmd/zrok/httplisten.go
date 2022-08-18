package main

import (
	"github.com/openziti-test-kitchen/zrok/endpoints/listen"
	"github.com/spf13/cobra"
)

func init() {
	httpCmd.AddCommand(newHttpListenCommand().cmd)
}

type httpListenCommand struct {
	endpoint string
	cmd      *cobra.Command
}

func newHttpListenCommand() *httpListenCommand {
	cmd := &cobra.Command{
		Use:   "listen <zitiIdentity>",
		Short: "Create an HTTP listener",
		Args:  cobra.ExactArgs(1),
	}
	command := &httpListenCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.endpoint, "endpoint", "e", "0.0.0.0:10111", "Address for HTTP listening endpoint")
	cmd.Run = command.run
	return command
}

func (self *httpListenCommand) run(_ *cobra.Command, args []string) {
	httpListener, err := listen.NewHTTP(&listen.Config{
		IdentityPath: args[0],
		Address:      self.endpoint,
	})
	if err != nil {
		panic(err)
	}
	if err := httpListener.Run(); err != nil {
		panic(err)
	}
}
