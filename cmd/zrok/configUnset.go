package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	configCmd.AddCommand(newConfigUnsetCommand().cmd)
}

type configUnsetCommand struct {
	cmd *cobra.Command
}

func newConfigUnsetCommand() *configUnsetCommand {
	cmd := &cobra.Command{
		Use:   "unset <configName>",
		Short: "Unset a value from the environment config",
		Args:  cobra.ExactArgs(1),
	}
	command := &configUnsetCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *configUnsetCommand) run(_ *cobra.Command, args []string) {
	configName := args[0]

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	switch configName {
	case "apiEndpoint":
		if err := env.SetConfig(&env_core.Config{}); err != nil {
			tui.Error("unable to save config", err)
		}
		fmt.Println("zrok configuration updated")
		if env.IsEnabled() {
			fmt.Printf("\n[%v]: because you have a %v-d environment, you won't see your config change until you run %v first!\n\n", tui.WarningLabel, tui.Code.Render("zrok enable"), tui.Code.Render("zrok disable"))
		}

	default:
		fmt.Printf("unknown config name '%v'\n", configName)
		os.Exit(1)
	}
}
