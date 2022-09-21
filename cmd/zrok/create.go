package main

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti/foundation/v2/term"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.AddCommand(newCreateAccountCommand().cmd)
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create objects",
}

type createAccountCommand struct {
	cmd *cobra.Command
}

func newCreateAccountCommand() *createAccountCommand {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Create new zrok account",
		Args:  cobra.ExactArgs(0),
	}
	command := &createAccountCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *createAccountCommand) run(_ *cobra.Command, _ []string) {
	email, err := term.Prompt("New Email: ")
	if err != nil {
		panic(err)
	}
	confirm, err := term.Prompt("Confirm Email: ")
	if err != nil {
		panic(err)
	}
	if confirm != email {
		fmt.Println("entered emails do not match... aborting!")
	}

	zrok, err := newZrokClient(apiEndpoint)
	if err != nil {
		panic(err)
	}
	req := identity.NewCreateAccountParams()
	req.Body = &rest_model_zrok.AccountRequest{
		Email: email,
	}
	_, err = zrok.Identity.CreateAccount(req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("registration invitation sent to '%v'!\n", email)
}
