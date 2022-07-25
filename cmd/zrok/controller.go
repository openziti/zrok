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
	rootCmd.AddCommand(controllerCmd)
}

var controllerCmd = &cobra.Command{
	Use:     "controller <configPath>",
	Short:   "Start a zrok controller",
	Aliases: []string{"ctrl"},
	Run: func(_ *cobra.Command, args []string) {
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
			Host: host,
			Port: port,
			Store: &store.Config{
				Path: "zrok.db",
			},
		}); err != nil {
			panic(err)
		}
	},
}
