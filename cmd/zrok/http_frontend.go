package main

import (
	"github.com/openziti-test-kitchen/zrok/endpoints/frontend"
	"github.com/spf13/cobra"
)

func init() {
	httpCmd.AddCommand(newHttpFrontendCommand().cmd)
}

type httpFrontendCommand struct {
	endpoint string
	cmd      *cobra.Command
}

func newHttpFrontendCommand() *httpFrontendCommand {
	cmd := &cobra.Command{
		Use:     "frontend <zitiIdentity>",
		Aliases: []string{"fe"},
		Short:   "Create an HTTP frontend",
		Args:    cobra.ExactArgs(1),
	}
	command := &httpFrontendCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.endpoint, "endpoint", "e", "0.0.0.0:10180", "Bind address for HTTP frontend")
	cmd.Run = command.run
	return command
}

func (self *httpFrontendCommand) run(_ *cobra.Command, args []string) {
	httpListener, err := frontend.NewHTTP(&frontend.Config{
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
