package main

import (
	"fmt"
	"github.com/openziti/zrok/environment/env_v0_3"
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

	zrd, err := env_v0_3.Load()
	if err != nil {
		panic(err)
	}

	modified := false
	switch configName {
	case "apiEndpoint":
		if zrd.Cfg != nil && zrd.Cfg.ApiEndpoint != "" {
			zrd.Cfg.ApiEndpoint = ""
			modified = true
		}

	default:
		fmt.Printf("unknown config name '%v'\n", configName)
		os.Exit(1)
	}

	if modified {
		if err := zrd.Save(); err != nil {
			panic(err)
		}
		fmt.Println("zrok configuration updated")
		if zrd.Env != nil && configName == "apiEndpoint" {
			fmt.Printf("\n[%v]: because you have a %v-d environment, you won't see your config change until you run %v first!\n\n", tui.WarningLabel, tui.Code.Render("zrok enable"), tui.Code.Render("zrok disable"))
		}
	} else {
		fmt.Println("zrok configuration not changed")
	}
}
