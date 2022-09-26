package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/identity"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/spf13/cobra"
	"os"
	user2 "os/user"
	"strings"
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
		showError(fmt.Sprintf("you already have an environment '%v' for '%v'", env.ZitiIdentityId, env.ZrokToken), nil)
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
		if !panicInstead {
			showError("the zrok service returned an error", err)
		}
		panic(err)
	}
	if err := zrokdir.SaveEnvironment(&zrokdir.Environment{ZrokToken: token, ZitiIdentityId: resp.Payload.Identity, ApiEndpoint: apiEndpoint}); err != nil {
		if !panicInstead {
			showError("there was an error saving the new environment", err)
		}
		panic(err)
	}
	if err := zrokdir.WriteZitiIdentity("environment", resp.Payload.Cfg); err != nil {
		if !panicInstead {
			showError("there was an error writing the environment file", err)
		}
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

func showError(msg string, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v (%v)\n", msg, strings.TrimSpace(err.Error()))
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", msg)
	}
	os.Exit(1)
}
