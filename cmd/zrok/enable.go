package main

import (
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
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
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", token)
	req := identity.NewEnableParams()
	resp, err := zrok.Identity.Enable(req, auth)
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
