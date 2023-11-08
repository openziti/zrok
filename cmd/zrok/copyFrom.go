package main

import (
	"github.com/io-developer/go-davsync/pkg/client/webdav"
	"github.com/spf13/cobra"
)

func init() {
	copyCmd.AddCommand(newCopyFromCommand().cmd)
}

type copyFromCommand struct {
	cmd *cobra.Command
}

func newCopyFromCommand() *copyFromCommand {
	cmd := &cobra.Command{
		Use:   "from <share> [<destination>]",
		Short: "Copy files from zrok drive to destination",
		Args:  cobra.RangeArgs(1, 2),
	}
	command := &copyFromCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *copyFromCommand) run(_ *cobra.Command, args []string) {
	_ = &webdav.Options{}
}
