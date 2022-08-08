package main

import (
	"github.com/openziti-test-kitchen/zrok/controller"
	"github.com/openziti-test-kitchen/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

func init() {
	rootCmd.AddCommand(newControllerCommand().cmd)
}

type controllerCommand struct {
	dbPath string
	cmd    *cobra.Command
}

func newControllerCommand() *controllerCommand {
	cmd := &cobra.Command{
		Use:     "controller",
		Short:   "Start a zrok controller",
		Aliases: []string{"ctrl"},
	}
	ccmd := &controllerCommand{
		cmd: cmd,
	}
	cmd.Run = ccmd.run
	cmd.Flags().StringVarP(&ccmd.dbPath, "database", "d", "zrok.db", "Path to zrok controller database")
	return ccmd
}

func (cmd *controllerCommand) run(_ *cobra.Command, _ []string) {
	tokens := strings.Split(endpoint, ":")
	if len(tokens) != 2 {
		panic(errors.Errorf("malformed endpoint '%v'", endpoint))
	}
	host := tokens[0]
	port, err := strconv.Atoi(tokens[1])
	if err != nil {
		panic(err)
	}
	if err := controller.Run(&controller.Config{
		Host:  host,
		Port:  port,
		Store: &store.Config{Path: cmd.dbPath},
	}); err != nil {
		panic(err)
	}
}
