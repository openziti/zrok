package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	modifyCmd.AddCommand(newModifyShareCommand().cmd)
}

type modifyShareCommand struct {
	addAccessGrants    []string
	removeAccessGrants []string
	cmd                *cobra.Command
}

func newModifyShareCommand() *modifyShareCommand {
	cmd := &cobra.Command{
		Use:   "share <shareToken>",
		Args:  cobra.ExactArgs(1),
		Short: "Modify a share",
	}
	command := &modifyShareCommand{cmd: cmd}
	cmd.Flags().StringArrayVar(&command.addAccessGrants, "add-access-grant", []string{}, "Add an access grant (email address)")
	cmd.Flags().StringArrayVar(&command.removeAccessGrants, "remove-access-grant", []string{}, "Remove an access grant (email address)")
	cmd.Run = command.run
	return command
}

func (cmd *modifyShareCommand) run(_ *cobra.Command, args []string) {
	shrToken := args[0]

	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading environment", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	zrok, err := root.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create zrok client", err)
		}
		panic(err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)

	if len(cmd.addAccessGrants) > 0 || len(cmd.removeAccessGrants) > 0 {
		req := share.NewUpdateShareParams()
		req.Body.ShareToken = shrToken
		req.Body.AddAccessGrants = cmd.addAccessGrants
		req.Body.RemoveAccessGrants = cmd.removeAccessGrants
		if _, err := zrok.Share.UpdateShare(req, auth); err != nil {
			if !panicInstead {
				tui.Error("unable to update share", err)
			}
			panic(err)
		}
		fmt.Println("updated")
	}
}
