package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/rest_client_zrok/metadata"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(newListAccessesCommand().cmd)
}

type listAccessesCommand struct {
	cmd *cobra.Command

	// text search filters
	envZId      string
	shareToken  string
	bindAddress string
	description string

	// date range filters
	createdAfter  string
	createdBefore string
	updatedAfter  string
	updatedBefore string

	// output control
	jsonOutput bool
}

func newListAccessesCommand() *listAccessesCommand {
	cmd := &cobra.Command{
		Use:   "accesses",
		Short: "list accesses in your account with optional filtering",
		Args:  cobra.NoArgs,
	}
	command := &listAccessesCommand{cmd: cmd}

	// text search filters
	cmd.Flags().StringVar(&command.envZId, "env-zid", "", "filter by environment ziti identity")
	cmd.Flags().StringVar(&command.shareToken, "share-token", "", "filter by associated share token")
	cmd.Flags().StringVar(&command.bindAddress, "bind-address", "", "filter by bind address (substring match)")
	cmd.Flags().StringVar(&command.description, "description", "", "filter by description (substring match)")

	// date range filters
	cmd.Flags().StringVar(&command.createdAfter, "created-after", "", "filter by created date (RFC3339 format)")
	cmd.Flags().StringVar(&command.createdBefore, "created-before", "", "filter by created date (RFC3339 format)")
	cmd.Flags().StringVar(&command.updatedAfter, "updated-after", "", "filter by updated date (RFC3339 format)")
	cmd.Flags().StringVar(&command.updatedBefore, "updated-before", "", "filter by updated date (RFC3339 format)")

	// output control
	cmd.Flags().BoolVar(&command.jsonOutput, "json", false, "output raw JSON instead of table")

	cmd.Run = command.run
	return command
}

func (cmd *listAccessesCommand) run(_ *cobra.Command, _ []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	// build request with filters
	req := metadata.NewListAccessesParams()

	// text search filters
	if cmd.envZId != "" {
		req.EnvZID = &cmd.envZId
	}
	if cmd.shareToken != "" {
		req.ShareToken = &cmd.shareToken
	}
	if cmd.bindAddress != "" {
		req.BindAddress = &cmd.bindAddress
	}
	if cmd.description != "" {
		req.Description = &cmd.description
	}

	// date range filters
	if cmd.createdAfter != "" {
		req.CreatedAfter = &cmd.createdAfter
	}
	if cmd.createdBefore != "" {
		req.CreatedBefore = &cmd.createdBefore
	}
	if cmd.updatedAfter != "" {
		req.UpdatedAfter = &cmd.updatedAfter
	}
	if cmd.updatedBefore != "" {
		req.UpdatedBefore = &cmd.updatedBefore
	}

	// call API
	resp, err := zrok.Metadata.ListAccesses(req, auth)
	if err != nil {
		panic(err)
	}

	accesses := resp.Payload.Accesses

	// if JSON flag is set, output raw JSON and return
	if cmd.jsonOutput {
		jsonBytes, err := json.MarshalIndent(resp.Payload, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonBytes))
		return
	}

	// tabular output
	fmt.Println()
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.AppendHeader(table.Row{"ID", "Frontend Token", "Env", "Bind Address", "Share", "Backend", "Description", "Limited", "Created"})

	for _, access := range accesses {
		// format env zid (truncate if too long)
		envZId := access.EnvZID
		if envZId == "" {
			envZId = "-"
		} else if len(envZId) > 10 {
			envZId = envZId[:10] + "..."
		}

		// format bind address
		bindAddr := access.BindAddress
		if bindAddr == "" {
			bindAddr = "-"
		} else if len(bindAddr) > 20 {
			bindAddr = bindAddr[:20] + "..."
		}

		// format share token
		shareToken := access.ShareToken
		if shareToken == "" {
			shareToken = "-"
		} else if len(shareToken) > 15 {
			shareToken = shareToken[:15] + "..."
		}

		// format backend mode
		backendMode := access.BackendMode
		if backendMode == "" {
			backendMode = "-"
		}

		// format description
		description := access.Description
		if description == "" {
			description = "-"
		} else if len(description) > 20 {
			description = description[:20] + "..."
		}

		// format limited status
		limitedStatus := ""
		if access.Limited {
			limitedStatus = "!!"
		}

		// format created timestamp
		created := time.Unix(access.CreatedAt/1000, 0).Format("2006-01-02 15:04:05")

		t.AppendRow(table.Row{
			access.ID,
			access.FrontendToken,
			envZId,
			bindAddr,
			shareToken,
			backendMode,
			description,
			limitedStatus,
			created,
		})
	}

	t.Render()
	fmt.Println()

	// show summary
	if len(accesses) == 0 {
		fmt.Println("no accesses found matching the specified filters.")
		fmt.Println()
	} else {
		fmt.Printf("total: %d access(es)\n", len(accesses))
		fmt.Println()
	}
}
