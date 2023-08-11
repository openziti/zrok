package main

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	rootCmd.AddCommand(newCaddyCommand().cmd)
}

type caddyCommand struct {
	cmd *cobra.Command
}

func newCaddyCommand() *caddyCommand {
	cmd := &cobra.Command{
		Use:   "caddy",
		Short: "Run an embedded caddy backend",
		Args:  cobra.ExactArgs(0),
	}
	command := &caddyCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *caddyCommand) run(_ *cobra.Command, _ []string) {
	caddy.Run(&caddy.Config{})
	for {
		time.Sleep(30 * time.Minute)
	}
}
