package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(newListNamespacesCommand().cmd)
}

type listNamespacesCommand struct {
	cmd  *cobra.Command
	json bool
}

func newListNamespacesCommand() *listNamespacesCommand {
	cmd := &cobra.Command{
		Use:   "namespaces",
		Short: "List available namespaces",
		Args:  cobra.NoArgs,
	}
	command := &listNamespacesCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.json, "json", false, "output raw JSON instead of formatted tables")
	cmd.Run = command.run
	return command
}

func (cmd *listNamespacesCommand) run(_ *cobra.Command, _ []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		dl.Fatal(err)
	}

	req := share.NewListShareNamespacesParams()
	resp, err := zrok.Share.ListShareNamespaces(req, auth)
	if err != nil {
		dl.Fatal(err)
	}

	if cmd.json {
		jsonBytes, err := json.MarshalIndent(resp.Payload, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonBytes))
		return
	}

	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"Name", "Namespace Token", "Description"})

	for _, namespace := range resp.Payload {
		t.AppendRow(table.Row{namespace.Name, namespace.NamespaceToken, namespace.Description})
	}

	t.Render()
	fmt.Println()
}
