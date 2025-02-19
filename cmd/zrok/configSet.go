package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"strconv"
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
		Long:  "Set a value into the environment config. Use 'zrok status' to list available configuration names and current values.",
		Args:  cobra.ExactArgs(2),
	}
	command := &configSetCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *configSetCommand) run(_ *cobra.Command, args []string) {
	configName := args[0]
	value := args[1]

	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	switch configName {
	case "apiEndpoint":
		ok, err := isFullyValidUrl(value)
		if err != nil {
			tui.Error("unable to validate api endpoint", err)
		}
		if !ok {
			tui.Error("invalid apiEndpoint; please make sure URL starts with http:// or https://", nil)
		}
		if env.Config() == nil {
			if err := env.SetConfig(&env_core.Config{ApiEndpoint: value}); err != nil {
				tui.Error("unable to save config", err)
			}
		} else {
			cfg := env.Config()
			cfg.ApiEndpoint = value
			if err := env.SetConfig(cfg); err != nil {
				tui.Error("unable to save config", err)
			}
		}
		fmt.Println("zrok configuration updated")
		if env.IsEnabled() {
			fmt.Printf("\n[%v]: because you have a %v-d environment, you won't see your config change until you run %v first!\n\n", tui.WarningLabel, tui.Code.Render("zrok enable"), tui.Code.Render("zrok disable"))
		}

	case "defaultFrontend":
		if env.Config() == nil {
			if err := env.SetConfig(&env_core.Config{DefaultFrontend: value}); err != nil {
				tui.Error("unable to save config", err)
			}
		} else {
			cfg := env.Config()
			cfg.DefaultFrontend = value
			if err := env.SetConfig(cfg); err != nil {
				tui.Error("unable to save config", err)
			}
		}
		fmt.Println("zrok configuration updated")

	case "headless":
		headless, err := strconv.ParseBool(value)
		if err != nil {
			tui.Error("unable to parse value for 'headless': %v", err)
		}
		if env.Config() == nil {
			if err := env.SetConfig(&env_core.Config{Headless: headless}); err != nil {
				tui.Error("unable to save config", err)
			}
		} else {
			cfg := env.Config()
			cfg.Headless = headless
			if err := env.SetConfig(cfg); err != nil {
				tui.Error("unable to save config", err)
			}
		}
		fmt.Println("zrok configuration updated")

	default:
		fmt.Printf("unknown config name '%v'\n", configName)
		os.Exit(1)
	}
}

func isFullyValidUrl(rawUrl string) (bool, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return false, err
	}
	if u.Scheme == "" || u.Host == "" {
		return false, nil
	}
	return true, nil
}
