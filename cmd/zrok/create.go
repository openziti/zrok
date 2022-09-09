package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti/foundation/v2/term"
	"github.com/sirupsen/logrus"
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
	password, err := term.PromptPassword("New Password: ", false)
	if err != nil {
		panic(err)
	}
	confirm, err := term.PromptPassword("Confirm Password: ", false)
	if err != nil {
		panic(err)
	}
	if confirm != password {
		panic("confirmed password mismatch")
	}

	zrok := newZrokClient()
	req := identity.NewCreateAccountParams()
	req.Body = &rest_model_zrok.AccountRequest{
		Email:    email,
		Password: password,
	}
	resp, err := zrok.Identity.CreateAccount(req)
	if err != nil {
		panic(err)
	}

	logrus.Infof("api token: %v", resp.Payload.Token)
}
