package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/studio-b12/gowebdav"
	"path/filepath"
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
	c := gowebdav.NewClient(args[0], "", "")
	if err := c.Connect(); err != nil {
		panic(err)
	}
	if err := cmd.recurseTree(c, ""); err != nil {
		panic(err)
	}
}

func (cmd *copyFromCommand) recurseTree(c *gowebdav.Client, path string) error {
	files, err := c.ReadDir(path)
	if err != nil {
		return err
	}
	for _, f := range files {
		sub := filepath.ToSlash(filepath.Join(path, f.Name()))
		if f.IsDir() {
			logrus.Infof("-> %v", sub)
			if err := cmd.recurseTree(c, sub); err != nil {
				return err
			}
		} else {
			etag := "<etag>"
			if v, ok := f.(gowebdav.File); ok {
				etag = v.ETag()
			}
			logrus.Infof("++ %v (%v)", sub, etag)
		}
	}
	return nil
}
