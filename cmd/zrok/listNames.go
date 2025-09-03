package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(newListNamesCommand().cmd)
}

type listNamesCommand struct {
	cmd            *cobra.Command
	namespaceToken string
}

func newListNamesCommand() *listNamesCommand {
	cmd := &cobra.Command{
		Use:   "names",
		Short: "list names within a namespace or all accessible namespaces",
		Args:  cobra.NoArgs,
	}
	command := &listNamesCommand{cmd: cmd}
	cmd.Flags().StringVarP(&command.namespaceToken, "namespace-token", "n", "", "namespace token")
	cmd.Run = command.run
	return command
}

func (cmd *listNamesCommand) run(_ *cobra.Command, args []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	if cmd.namespaceToken != "" {
		// list names for specific namespace
		req := share.NewListShareNamesParams()
		req.NamespaceToken = cmd.namespaceToken

		resp, err := zrok.Share.ListShareNames(req, auth)
		if err != nil {
			panic(err)
		}

		fmt.Println()
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"Name", "Created At"})
		for _, name := range resp.Payload {
			t.AppendRow(table.Row{
				name.Name,
				time.Unix(name.CreatedAt, 0),
			})
		}
		t.Render()
		fmt.Println()
	} else {
		// list all names across all accessible namespaces
		req := share.NewListAllShareNamesParams()

		resp, err := zrok.Share.ListAllShareNames(req, auth)
		if err != nil {
			panic(err)
		}

		fmt.Println()
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleRounded)
		t.AppendHeader(table.Row{"Name", "Reserved", "Namespace", "Created At"})
		for _, name := range resp.Payload {
			t.AppendRow(table.Row{
				name.Name,
				name.Reserved,
				name.NamespaceName,
				time.Unix(name.CreatedAt, 0),
			})
		}
		t.Render()
		fmt.Println()
	}
}
