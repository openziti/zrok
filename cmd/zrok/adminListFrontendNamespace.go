package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	adminListCmd.AddCommand(newAdminListFrontendNamespaceCommand().cmd)
}

type adminListFrontendNamespaceCommand struct {
	cmd *cobra.Command
}

func newAdminListFrontendNamespaceCommand() *adminListFrontendNamespaceCommand {
	cmd := &cobra.Command{
		Use:     "frontend-namespace <frontendToken>",
		Aliases: []string{"fn"},
		Short:   "List namespaces mapped to a frontend",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminListFrontendNamespaceCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminListFrontendNamespaceCommand) run(_ *cobra.Command, args []string) {
	frontendToken := args[0]

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := root.Client()
	if err != nil {
		panic(err)
	}

	// fetch all namespaces for lookup
	namespacesReq := admin.NewListNamespacesParams()
	namespacesResp, err := zrok.Admin.ListNamespaces(namespacesReq, mustGetAdminAuth())
	if err != nil {
		logrus.Errorf("error listing namespaces: %v", err)
		os.Exit(1)
	}

	// create namespace lookup map
	namespaceMap := make(map[string]*admin.ListNamespacesOKBodyItems0)
	for _, ns := range namespacesResp.Payload {
		namespaceMap[ns.NamespaceToken] = ns
	}

	// fetch frontend-namespace mappings
	req := admin.NewListFrontendNamespaceMappingsParams()
	req.FrontendToken = frontendToken

	resp, err := zrok.Admin.ListFrontendNamespaceMappings(req, mustGetAdminAuth())
	if err != nil {
		logrus.Errorf("error listing frontend-namespace mappings: %v", err)
		os.Exit(1)
	}

	if len(resp.Payload) > 0 {
		fmt.Println()
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"namespace token", "name", "description", "open", "default", "created"})
		for _, mapping := range resp.Payload {
			if ns, exists := namespaceMap[mapping.NamespaceToken]; exists {
				created := time.Unix(ns.CreatedAt, 0).Format("2006-01-02 15:04:05")
				t.AppendRow(table.Row{mapping.NamespaceToken, ns.Name, ns.Description, ns.Open, mapping.IsDefault, created})
			} else {
				t.AppendRow(table.Row{mapping.NamespaceToken, "[unknown]", "", "", mapping.IsDefault, ""})
			}
		}
		t.Render()
		fmt.Println()
	} else {
		fmt.Printf("no namespace mappings found for frontend '%v'\n", frontendToken)
	}
}
