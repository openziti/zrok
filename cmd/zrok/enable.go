package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	enableCmd.Flags().StringVarP(&enableDescription, "description", "d", "", "Description of this environment")
	rootCmd.AddCommand(enableCmd)
}

var enableCmd = &cobra.Command{
	Use:   "enable <token>",
	Short: "Enable an environment for zrok",
	Args:  cobra.ExactArgs(1),
	Run:   enable,
}
var enableDescription string

func enable(_ *cobra.Command, args []string) {
	token := args[0]

	thisHost, err := getHost()
	if err != nil {
		panic(err)
	}

	zrok := newZrokClient()
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", token)
	req := identity.NewEnableParams()
	req.Body = &rest_model_zrok.EnableRequest{
		Description: enableDescription,
		Host:        thisHost,
	}
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

func getHost() (string, error) {
	info, err := host.Info()
	if err != nil {
		return "", err
	}
	thisHost := fmt.Sprintf("%v; %v; %v; %v; %v; %v; %v",
		info.Hostname, info.OS, info.Platform, info.PlatformFamily, info.PlatformVersion, info.KernelVersion, info.KernelArch)
	return thisHost, nil
}
