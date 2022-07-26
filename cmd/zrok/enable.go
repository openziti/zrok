package main

import (
	"fmt"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
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
	req.Body = &rest_model_zrok.EnableRequest{
		Token: args[0],
	}
	resp, err := zrok.Identity.Enable(req)
	if err != nil {
		panic(err)
	}

	cfgFile, err := os.Create(fmt.Sprintf("%v.json", resp.Payload.Identity))
	if err != nil {
		panic(err)
	}
	defer func() { _ = cfgFile.Close() }()
	_, err = cfgFile.Write([]byte(resp.Payload.Cfg))
	if err != nil {
		panic(err)
	}

	logrus.Infof("enabled, identity = '%v'", resp.Payload.Identity)
}
