package main

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/spf13/cobra"
	"os"
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
		Use:   "caddy <configPath>",
		Short: "Run an embedded caddy backend",
		Args:  cobra.ExactArgs(1),
	}
	command := &caddyCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *caddyCommand) run(_ *cobra.Command, args []string) {
	if err := caddy.Run(&caddy.Config{}); err != nil {
		panic(err)
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		panic(err)
	}
	var adapter caddyfile.Adapter
	adapter.ServerType = httpcaddyfile.ServerType{}
	cfg, warn, err := adapter.Adapt(data, map[string]interface{}{"filename": args[0]})
	if err != nil {
		panic(err)
	}
	for _, w := range warn {
		fmt.Println(w.Message)
	}
	fmt.Printf("cfg: %v\n", string(cfg))
	if err := caddy.Load(cfg, true); err != nil {
		panic(err)
	}
	for {
		time.Sleep(30 * time.Minute)
	}
}
