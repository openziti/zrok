package main

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/cmd/zrok/endpoint_ui"
	"github.com/spf13/cobra"
	"net/http"
)

func init() {
	testCmd.AddCommand(newTestEndpointCommand().cmd)
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Utilities used for testing zrok",
}

type testEndpointCommand struct {
	address string
	port    uint16
	cmd     *cobra.Command
}

func newTestEndpointCommand() *testEndpointCommand {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Start a simple HTTP endpoint",
		Args:  cobra.ExactArgs(0),
	}
	command := &testEndpointCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.address, "address", "a", "0.0.0.0", "The address for the HTTP listener")
	cmd.Flags().Uint16VarP(&command.port, "port", "p", 9090, "The port for the HTTP listener")
	cmd.Run = command.run
	return command
}

func (cmd *testEndpointCommand) run(_ *cobra.Command, _ []string) {
	fs := http.FileServer(http.FS(endpoint_ui.FS))
	if err := http.ListenAndServe(fmt.Sprintf("%v:%d", cmd.address, cmd.port), fs); err != nil {
		panic(err)
	}
}
