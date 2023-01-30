package main

import (
	"fmt"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/zrokdir"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	configCmd.AddCommand(newConfigSetCommand().cmd)
}

type configSetCommand struct {
	cmd *cobra.Command
}

func newConfigSetCommand() *configSetCommand {
	cmd := &cobra.Command{
		Use:   "set <configName> <value>",
		Short: "Set a value into the environment config",
		Args:  cobra.ExactArgs(2),
	}
	command := &configSetCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *configSetCommand) run(_ *cobra.Command, args []string) {
	configName := args[0]
	value := args[1]

	zrd, err := zrokdir.Load()
	if err != nil {
		panic(err)
	}

	modified := false
	switch configName {
	case "apiEndpoint":
		if zrd.Cfg == nil {
			zrd.Cfg = &zrokdir.Config{}
		}
		zrd.Cfg.ApiEndpoint = value
		modified = true

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
	}
}
