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
	rebaseCmd.AddCommand(newRebaseAccountTokenCommand().cmd)
}

type rebaseAccountTokenCommand struct {
	cmd *cobra.Command
}

func newRebaseAccountTokenCommand() *rebaseAccountTokenCommand {
	cmd := &cobra.Command{
		Use:   "accountToken <accountToken>",
		Short: "Rebase an enabled environment onto a different account token",
		Args:  cobra.ExactArgs(1),
	}
	command := &rebaseAccountTokenCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *rebaseAccountTokenCommand) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading root", err)
	}

	if !root.IsEnabled() {
		tui.Error("environment not enabled; 'zrok enable' your environment instead", nil)
	}

	env := root.Environment()
	if args[0] != env.AccountToken {
		fmt.Printf("this action will rebase your enabled environment to use the account token '%v'\n", args[0])
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

		env.AccountToken = args[0]
		if err := root.SetEnvironment(env); err != nil {
			tui.Error("error rebasing environment", err)
		}

		fmt.Printf("environment rebased to account token '%v'\n", env.AccountToken)
	} else {
		fmt.Printf("environment already configured to use the account token '%v'\n", env.AccountToken)
	}
}
