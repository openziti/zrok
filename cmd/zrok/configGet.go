package main

import (
	"fmt"

	"github.com/openziti/zrok/v2/environment"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.AddCommand(newConfigGetCommand().cmd)
}

type configGetCommand struct {
	cmd *cobra.Command
}

func newConfigGetCommand() *configGetCommand {
	cmd := &cobra.Command{
		Use:   "get <configName>",
		Short: "Get a value from the environment config",
		Long:  "Get a value from the environment config. Use 'zrok status' to list available configuration names and current values.",
		Args:  cobra.ExactArgs(1),
	}
	command := &configGetCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *configGetCommand) run(_ *cobra.Command, args []string) {
	configName := args[0]

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	switch configName {
	case "apiEndpoint":
		if env.Config() != nil && env.Config().ApiEndpoint != "" {
			fmt.Printf("apiEndpoint = %v\n", env.Config().ApiEndpoint)
		} else {
			fmt.Println("apiEndpoint = <unset>")
		}
	case "defaultNamespace":
		if env.Config() != nil && env.Config().DefaultNamespace != "" {
			fmt.Printf("defaultNamespace = %v\n", env.Config().DefaultNamespace)
		} else {
			fmt.Println("defaultNamespace = <unset>")
		}
	case "headless":
		if env.Config() != nil {
			fmt.Printf("headless = %v\n", env.Config().Headless)
		} else {
			fmt.Println("headless = <unset>")
		}
	case "superNetwork":
		if env.Config() != nil {
			fmt.Printf("superNetwork = %v\n", env.Config().SuperNetwork)
		} else {
			fmt.Println("superNetwork = <unset>")
		}
	default:
		fmt.Printf("unknown config name '%v'\n", configName)
	}
}
