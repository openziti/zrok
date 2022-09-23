package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/spf13/cobra"
	user2 "os/user"
)

func init() {
	rootCmd.AddCommand(newEnableCommand().cmd)
}

type enableCommand struct {
	description string
	cmd         *cobra.Command
}

func newEnableCommand() *enableCommand {
	cmd := &cobra.Command{
		Use:   "enable <token>",
		Short: "Enable an environment for zrok",
		Args:  cobra.ExactArgs(1),
	}
	command := &enableCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.description, "description", "d", "<user>@<hostname>", "Description of this environment")
	cmd.Run = command.run
	return command
}

func (cmd *enableCommand) run(_ *cobra.Command, args []string) {
	env, err := zrokdir.LoadEnvironment()
	if err == nil {
		panic(errors.Errorf("environment '%v' already enabled!", env.ZitiIdentityId))
	}

	token := args[0]

	hostName, hostDetail, err := getHost()
	if err != nil {
		panic(err)
	}
	user, err := user2.Current()
	if err != nil {
		panic(err)
	}
	hostDetail = fmt.Sprintf("%v; %v", user.Username, hostDetail)
	if cmd.description == "<user>@<hostname>" {
		cmd.description = fmt.Sprintf("%v@%v", user.Username, hostName)
	}

	zrok, err := newZrokClient(apiEndpoint)
	if err != nil {
		panic(err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", token)
	req := identity.NewEnableParams()
	req.Body = &rest_model_zrok.EnableRequest{
		Description: cmd.description,
		Host:        hostDetail,
	}
	resp, err := zrok.Identity.Enable(req, auth)
	if err != nil {
		panic(err)
	}
	if err := zrokdir.SaveEnvironment(&zrokdir.Environment{ZrokToken: token, ZitiIdentityId: resp.Payload.Identity, ApiEndpoint: apiEndpoint}); err != nil {
		panic(err)
	}
	if err := zrokdir.WriteZitiIdentity("environment", resp.Payload.Cfg); err != nil {
		panic(err)
	}

	fmt.Printf("zrok environment '%v' enabled for '%v'\n", resp.Payload.Identity, token)
}

func getHost() (string, string, error) {
	info, err := host.Info()
	if err != nil {
		return "", "", err
	}
	thisHost := fmt.Sprintf("%v; %v; %v; %v; %v; %v; %v",
		info.Hostname, info.OS, info.Platform, info.PlatformFamily, info.PlatformVersion, info.KernelVersion, info.KernelArch)
	return info.Hostname, thisHost, nil
}
