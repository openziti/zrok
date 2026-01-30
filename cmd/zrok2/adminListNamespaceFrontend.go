package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminListCmd.AddCommand(newAdminListNamespaceFrontendCommand().cmd)
}

type adminListNamespaceFrontendCommand struct {
	cmd *cobra.Command
}

func newAdminListNamespaceFrontendCommand() *adminListNamespaceFrontendCommand {
	cmd := &cobra.Command{
		Use:     "namespace-frontend <namespaceToken>",
		Aliases: []string{"nf"},
		Short:   "List frontends mapped to a namespace",
		Args:    cobra.ExactArgs(1),
	}
	command := &adminListNamespaceFrontendCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminListNamespaceFrontendCommand) run(_ *cobra.Command, args []string) {
	namespaceToken := args[0]

	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := root.Client()
	if err != nil {
		panic(err)
	}

	// fetch all frontends for lookup
	frontendsReq := admin.NewListFrontendsParams()
	frontendsResp, err := zrok.Admin.ListFrontends(frontendsReq, mustGetAdminAuth())
	if err != nil {
		dl.Errorf("error listing frontends: %v", err)
		os.Exit(1)
	}

	// create frontend lookup map
	frontendMap := make(map[string]*admin.ListFrontendsOKBodyItems0)
	for _, fe := range frontendsResp.Payload {
		frontendMap[fe.FrontendToken] = fe
	}

	// fetch namespace-frontend mappings
	req := admin.NewListNamespaceFrontendMappingsParams()
	req.NamespaceToken = namespaceToken

	resp, err := zrok.Admin.ListNamespaceFrontendMappings(req, mustGetAdminAuth())
	if err != nil {
		dl.Errorf("error listing namespace-frontend mappings: %v", err)
		os.Exit(1)
	}

	if len(resp.Payload) > 0 {
		fmt.Println()
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"frontend token", "public name", "url template", "default", "created"})
		for _, mapping := range resp.Payload {
			if fe, exists := frontendMap[mapping.FrontendToken]; exists {
				created := time.UnixMilli(fe.CreatedAt).Format("2006-01-02 15:04:05")
				t.AppendRow(table.Row{mapping.FrontendToken, fe.PublicName, fe.URLTemplate, mapping.IsDefault, created})
			} else {
				t.AppendRow(table.Row{mapping.FrontendToken, "[unknown]", "", mapping.IsDefault, ""})
			}
		}
		t.Render()
		fmt.Println()
	} else {
		fmt.Printf("no frontend mappings found for namespace '%v'\n", namespaceToken)
	}
}
