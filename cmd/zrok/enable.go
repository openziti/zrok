package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_model"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_client/identity"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(enableCmd)
}

var enableCmd = &cobra.Command{
	Use:   "enable <token>",
	Short: "Enable an environment for zrok",
	Run:   enable,
}

func enable(_ *cobra.Command, args []string) {
	if len(args) != 1 {
		panic(errors.Errorf("provide a single zrok token"))
	}

	zrok := newZrokClient()
	req := identity.NewEnableParams()
	req.Body = &rest_model.EnableRequest{
		Token: args[0],
	}
	resp, err := zrok.Identity.Enable(req)
	if err != nil {
		panic(err)
	}

	logrus.Infof("enabled, identity = '%v'", resp.Payload.Identity)
}
