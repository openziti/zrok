package main

import (
	"fmt"
	"github.com/openziti/zrok/environment/env_v0_3"
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
		Args:  cobra.ExactArgs(1),
	}
	command := &configGetCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *configGetCommand) run(_ *cobra.Command, args []string) {
	configName := args[0]

	zrd, err := env_v0_3.Load()
	if err != nil {
		panic(err)
	}

	switch configName {
	case "apiEndpoint":
		if zrd.Cfg != nil && zrd.Cfg.ApiEndpoint != "" {
			fmt.Printf("apiEndpoint = %v\n", zrd.Cfg.ApiEndpoint)
		} else {
			fmt.Println("apiEndpoint = <unset>")
		}
	default:
		fmt.Printf("unknown config name '%v'\n", configName)
	}
}
