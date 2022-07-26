package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(enableCmd)
}

var enableCmd = &cobra.Command{
	Use:   "enable <token>",
	Short: "Enable an environment for zrok",
	Args:  cobra.ExactArgs(1),
	Run:   enable,
}

func enable(_ *cobra.Command, args []string) {
	token := args[0]

	zrok := newZrokClient()
	req := identity.NewEnableParams()
	req.Body = &rest_model_zrok.EnableRequest{
		Token: token,
	}
	resp, err := zrok.Identity.Enable(req)
	if err != nil {
		panic(err)
	}
	if err := zrokdir.WriteToken(token); err != nil {
		panic(err)
	}
	if err := zrokdir.WriteIdentityId(resp.Payload.Identity); err != nil {
		panic(err)
	}
	if err := zrokdir.WriteIdentityConfig(resp.Payload.Cfg); err != nil {
		panic(err)
	}
	logrus.Infof("enabled, identity = '%v'", resp.Payload.Identity)
}
