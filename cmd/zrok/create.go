package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_model"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_client/identity"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.AddCommand(createAccountCmd)
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create objects",
}

var createAccountCmd = &cobra.Command{
	Use:   "account <username>",
	Short: "create new zrok account",
	Run: func(_ *cobra.Command, args []string) {
		zrok := newZrokClient()
		req := identity.NewCreateAccountParams()
		req.Body = &rest_model.AccountRequest{
			Username: args[0],
		}
		resp, err := zrok.Identity.CreateAccount(req)
		if err != nil {
			panic(err)
		}
		logrus.Infof("api token = '%v'", resp.Payload.APIToken)
	},
}
