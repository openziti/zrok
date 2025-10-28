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
	listCmd.AddCommand(newListEnvironmentsCommand().cmd)
}

type listEnvironmentsCommand struct {
	cmd *cobra.Command

	// text search filters
	description string
	host        string
	address     string

	// boolean filters
	remoteAgent *bool
	hasShares   *bool
	hasAccesses *bool
	hasActivity *bool

	// numeric filters
	shareCount  string
	accessCount string

	// date range filters
	createdAfter  string
	createdBefore string
	updatedAfter  string
	updatedBefore string

	// activity filter
	activityDuration string

	// output control
	jsonOutput bool
}

func newListEnvironmentsCommand() *listEnvironmentsCommand {
	cmd := &cobra.Command{
		Use:   "environments",
		Short: "list environments in your account with optional filtering",
		Args:  cobra.NoArgs,
	}
	command := &listEnvironmentsCommand{cmd: cmd}

	// text search filters
	cmd.Flags().StringVar(&command.description, "description", "", "filter by description (substring match)")
	cmd.Flags().StringVar(&command.host, "host", "", "filter by host (substring match)")
	cmd.Flags().StringVar(&command.address, "address", "", "filter by address (exact match)")

	// boolean filters
	cmd.Flags().BoolP("remote-agent", "r", false, "filter by remote agent enrollment")
	cmd.Flags().BoolP("has-shares", "s", false, "filter environments with shares")
	cmd.Flags().BoolP("has-accesses", "a", false, "filter environments with accesses")
	cmd.Flags().BoolP("has-activity", "A", false, "filter environments with recent activity")

	// numeric filters
	cmd.Flags().StringVar(&command.shareCount, "share-count", "", "filter by share count with operator (e.g., '>0', '>=5', '=3')")
	cmd.Flags().StringVar(&command.accessCount, "access-count", "", "filter by access count with operator (e.g., '>0', '>=2')")

	// date range filters
	cmd.Flags().StringVar(&command.createdAfter, "created-after", "", "filter by created date (RFC3339 format)")
	cmd.Flags().StringVar(&command.createdBefore, "created-before", "", "filter by created date (RFC3339 format)")
	cmd.Flags().StringVar(&command.updatedAfter, "updated-after", "", "filter by updated date (RFC3339 format)")
	cmd.Flags().StringVar(&command.updatedBefore, "updated-before", "", "filter by updated date (RFC3339 format)")

	// activity filter
	cmd.Flags().StringVar(&command.activityDuration, "activity-duration", "", "duration for hasActivity filter (e.g., '24h', '7d', '30d')")

	// output control
	cmd.Flags().BoolVar(&command.jsonOutput, "json", false, "output raw JSON instead of table")

	cmd.Run = command.run
	return command
}

func (cmd *listEnvironmentsCommand) run(_ *cobra.Command, _ []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	// build request with filters
	req := metadata.NewListEnvironmentsParams()

	// text search filters
	if cmd.description != "" {
		req.Description = &cmd.description
	}
	if cmd.host != "" {
		req.Host = &cmd.host
	}
	if cmd.address != "" {
		req.Address = &cmd.address
	}

	// boolean filters - only set if flag was explicitly provided
	if cmd.cmd.Flags().Changed("remote-agent") {
		val, _ := cmd.cmd.Flags().GetBool("remote-agent")
		req.RemoteAgent = &val
		cmd.remoteAgent = &val
	}
	if cmd.cmd.Flags().Changed("has-shares") {
		val, _ := cmd.cmd.Flags().GetBool("has-shares")
		req.HasShares = &val
		cmd.hasShares = &val
	}
	if cmd.cmd.Flags().Changed("has-accesses") {
		val, _ := cmd.cmd.Flags().GetBool("has-accesses")
		req.HasAccesses = &val
		cmd.hasAccesses = &val
	}
	if cmd.cmd.Flags().Changed("has-activity") {
		val, _ := cmd.cmd.Flags().GetBool("has-activity")
		req.HasActivity = &val
		cmd.hasActivity = &val
	}

	// numeric filters
	if cmd.shareCount != "" {
		req.ShareCount = &cmd.shareCount
	}
	if cmd.accessCount != "" {
		req.AccessCount = &cmd.accessCount
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

	// activity filter
	if cmd.activityDuration != "" {
		req.ActivityDuration = &cmd.activityDuration
	}

	// call API
	resp, err := zrok.Metadata.ListEnvironments(req, auth)
	if err != nil {
		panic(err)
	}

	environments := resp.Payload.Environments

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
	t.AppendHeader(table.Row{"ZID", "Description", "Host", "Address", "Agent", "Shares", "Accesses", "Limited", "Created"})

	for _, env := range environments {
		// format description
		description := env.Description
		if description == "" {
			description = "-"
		}

		// format host
		host := env.Host
		if host == "" {
			host = "-"
		}

		// format address
		address := env.Address
		if address == "" {
			address = "-"
		}

		// format agent status
		agentStatus := ""
		if env.RemoteAgent {
			agentStatus = "✓"
		}

		// format limited status
		limitedStatus := ""
		if env.Limited {
			limitedStatus = "!!"
		}

		// format created timestamp
		created := time.Unix(env.CreatedAt/1000, 0).Format("2006-01-02 15:04:05")

		t.AppendRow(table.Row{
			env.EnvZID,
			description,
			host,
			address,
			agentStatus,
			env.ShareCount,
			env.AccessCount,

			limitedStatus,
			created,
		})
	}

	t.Render()
	fmt.Println()

	// show summary
	if len(environments) == 0 {
		fmt.Println("no environments found matching the specified filters.")
		fmt.Println()
	} else {
		fmt.Printf("total: %d environment(s)\n", len(environments))
		fmt.Println()
	}
}
