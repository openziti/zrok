package main

import (
	"bufio"
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rebaseCmd.AddCommand(newRebaseApiEndpointCommand().cmd)
}

type rebaseApiEndpointCommand struct {
	cmd *cobra.Command
}

func newRebaseApiEndpointCommand() *rebaseApiEndpointCommand {
	cmd := &cobra.Command{
		Use:   "apiEndpoint <apiEndpoint>",
		Short: "Rebase an enabled environment onto a different API endpoint URL",
		Args:  cobra.ExactArgs(1),
	}
	command := &rebaseApiEndpointCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *rebaseApiEndpointCommand) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading root", err)
	}

	if !root.IsEnabled() {
		tui.Error("environment not enabled; 'zrok enable' your environment instead", nil)
	}

	currentEndpoint, _ := root.ApiEndpoint()
	if args[0] != currentEndpoint {
		fmt.Printf("this action will rebase your enabled environment to use the zrok API at: %v\n", args[0])
		fmt.Println()
		fmt.Println("you should only proceed if you understand why you're doing this!")
		fmt.Println()
		fmt.Print("to proceed, type 'yes': ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			text := scanner.Text()
			if text != "yes" {
				tui.Error("rebase aborted!", nil)
			}
		}
		fmt.Println()

		env := root.Environment()
		env.ApiEndpoint = args[0]

		if err := root.SetEnvironment(env); err != nil {
			tui.Error("error rebasing environment", err)
		}

		fmt.Printf("environment rebased to zrok API at: %v\n", env.ApiEndpoint)

	} else {
		fmt.Printf("environment already configured to use API endpoint: %v\n", currentEndpoint)
	}
}
