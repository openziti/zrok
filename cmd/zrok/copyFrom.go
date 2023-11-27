package main

import (
	"github.com/openziti/zrok/util/sync"
	"github.com/sirupsen/logrus"
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
	target := "."
	if len(args) == 2 {
		target = args[1]
	}

	t := sync.NewFilesystemTarget(&sync.FilesystemTargetConfig{
		Root: target,
	})
	destTree, err := t.Inventory()
	if err != nil {
		panic(err)
	}
	for _, f := range destTree {
		logrus.Infof("<- %v [%v, %v, %v]", f.Path, f.Size, f.Modified, f.ETag)
	}

	s, err := sync.NewWebDAVTarget(&sync.WebDAVTargetConfig{
		URL:      args[0],
		Username: "",
		Password: "",
	})
	if err != nil {
		panic(err)
	}
	srcTree, err := s.Inventory()
	if err != nil {
		panic(err)
	}
	for _, f := range srcTree {
		logrus.Infof("-> %v [%v, %v, %v]", f.Path, f.Size, f.Modified, f.ETag)
	}
}
