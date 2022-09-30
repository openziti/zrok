package main

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/foundation/v2/term"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newInviteCommand().cmd)
}

type inviteCommand struct {
	cmd *cobra.Command
}

func newInviteCommand() *inviteCommand {
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a new user to zrok",
		Args:  cobra.ExactArgs(0),
	}
	command := &inviteCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *inviteCommand) run(_ *cobra.Command, _ []string) {
	email, err := term.Prompt("New Email: ")
	if err != nil {
		panic(err)
	}
	confirm, err := term.Prompt("Confirm Email: ")
	if err != nil {
		panic(err)
	}
	if confirm != email {
		showError("entered emails do not match... aborting!", nil)
	}

	zrok, err := zrokdir.ZrokClient(apiEndpoint)
	if err != nil {
		if !panicInstead {
			showError("error creating zrok api client", err)
		}
		panic(err)
	}
	req := identity.NewCreateAccountParams()
	req.Body = &rest_model_zrok.AccountRequest{
		Email: email,
	}
	_, err = zrok.Identity.CreateAccount(req)
	if err != nil {
		if !panicInstead {
			showError("error creating account", err)
		}
		panic(err)
	}

	fmt.Printf("registration invitation sent to '%v'!\n", email)
}
