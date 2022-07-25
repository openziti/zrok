package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_model"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_client/identity"
	"github.com/openziti/foundation/v2/term"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.AddCommand(createAccountCmd)
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create objects",
}

var createAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "create new zrok account",
	Run: func(_ *cobra.Command, _ []string) {
		username, err := term.Prompt("New Username: ")
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
		req.Body = &rest_model.AccountRequest{
			Username: username,
			Password: password,
		}
		resp, err := zrok.Identity.CreateAccount(req)
		if err != nil {
			panic(err)
		}

		logrus.Infof("api token: %v", resp.Payload.Token)
	},
}
