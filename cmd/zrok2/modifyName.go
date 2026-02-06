package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
)

func init() {
	modifyCmd.AddCommand(newModifyNameCommand().cmd)
}

type modifyNameCommand struct {
	cmd            *cobra.Command
	namespaceToken string
	reserved       bool
}

func newModifyNameCommand() *modifyNameCommand {
	cmd := &cobra.Command{
		Use:   "name <name>",
		Short: "modify a name within a namespace",
		Args:  cobra.ExactArgs(1),
	}
	command := &modifyNameCommand{cmd: cmd}

	// default namespace handling
	defaultNamespace := "public"
	if root, err := environment.LoadRoot(); err == nil {
		defaultNamespace, _ = root.DefaultNamespace()
	}

	cmd.Flags().StringVarP(&command.namespaceToken, "namespace-token", "n", defaultNamespace, "namespace token")
	cmd.Flags().BoolVarP(&command.reserved, "reserved", "r", false, "set reservation state (true=reserved, false=ephemeral)")
	cmd.Run = command.run
	return command
}

func (cmd *modifyNameCommand) run(_ *cobra.Command, args []string) {
	name := args[0]

	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading environment", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok2 enable'?", nil)
	}

	zrok, err := root.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("unable to create zrok client", err)
		}
		panic(err)
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)

	req := share.NewUpdateShareNameParams()
	req.Body = share.UpdateShareNameBody{
		NamespaceToken: cmd.namespaceToken,
		Name:           name,
		Reserved:       cmd.reserved,
	}

	_, err = zrok.Share.UpdateShareName(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to update name", err)
		}
		panic(err)
	}

	reservedState := "ephemeral"
	if cmd.reserved {
		reservedState = "reserved"
	}
	fmt.Printf("updated name '%v' in namespace - now %v\n", name, reservedState)
}