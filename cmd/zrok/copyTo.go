package main

import (
	"github.com/openziti/zrok/util/sync"
	"github.com/spf13/cobra"
)

func init() {
	copyCmd.AddCommand(newCopyToCommand().cmd)
}

type copyToCommand struct {
	cmd *cobra.Command
}

func newCopyToCommand() *copyToCommand {
	cmd := &cobra.Command{
		Use:   "to <share> <source>",
		Short: "Copy files to a zrok drive from source",
		Args:  cobra.ExactArgs(2),
	}
	command := &copyToCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *copyToCommand) run(_ *cobra.Command, args []string) {
	dst, err := sync.NewWebDAVTarget(&sync.WebDAVTargetConfig{
		URL:      args[0],
		Username: "",
		Password: "",
	})
	if err != nil {
		panic(err)
	}
	src := sync.NewFilesystemTarget(&sync.FilesystemTargetConfig{
		Root: args[1],
	})

	if err := sync.Synchronize(src, dst); err != nil {
		panic(err)
	}
}
