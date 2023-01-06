package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/openziti-test-kitchen/zrok/build"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newVersionCommand().cmd)
}

type versionCommand struct {
	cmd *cobra.Command
}

func newVersionCommand() *versionCommand {
	cmd := &cobra.Command{
		Use:  "version",
		Args: cobra.ExactArgs(0),
	}
	command := &versionCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *versionCommand) run(_ *cobra.Command, _ []string) {
	zrokStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0080"))
	fmt.Println(zrokStyle.Render("               _    \n _____ __ ___ | | __\n|_  / '__/ _ \\| |/ /\n / /| | | (_) |   < \n/___|_|  \\___/|_|\\_\\") + "\n\n" + codeStyle.Render(build.String()) + "\n")
}
