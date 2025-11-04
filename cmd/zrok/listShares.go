package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/rest_client_zrok/metadata"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(newListSharesCommand().cmd)
}

type listSharesCommand struct {
	cmd *cobra.Command

	// text search filters
	envZId         string
	shareMode      string
	backendMode    string
	shareToken     string
	target         string
	permissionMode string

	// boolean filters
	hasActivity *bool
	idle        *bool

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

func newListSharesCommand() *listSharesCommand {
	cmd := &cobra.Command{
		Use:   "shares",
		Short: "list shares in your account with optional filtering",
		Args:  cobra.NoArgs,
	}
	command := &listSharesCommand{cmd: cmd}

	// text search filters
	cmd.Flags().StringVar(&command.envZId, "env-zid", "", "filter by environment ziti identity")
	cmd.Flags().StringVar(&command.shareMode, "share-mode", "", "filter by share mode (public/private)")
	cmd.Flags().StringVar(&command.backendMode, "backend-mode", "", "filter by backend mode")
	cmd.Flags().StringVar(&command.shareToken, "share-token", "", "filter by share token (substring match)")
	cmd.Flags().StringVar(&command.target, "target", "", "filter by target (substring match)")
	cmd.Flags().StringVar(&command.permissionMode, "permission-mode", "", "filter by permission mode (open/closed)")

	// boolean filters
	cmd.Flags().BoolP("has-activity", "A", false, "filter shares with recent activity")
	cmd.Flags().BoolP("idle", "I", false, "filter shares without recent activity")

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

func (cmd *listSharesCommand) run(_ *cobra.Command, _ []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	// build request with filters
	req := metadata.NewListSharesParams()

	// text search filters
	if cmd.envZId != "" {
		req.EnvZID = &cmd.envZId
	}
	if cmd.shareMode != "" {
		req.ShareMode = &cmd.shareMode
	}
	if cmd.backendMode != "" {
		req.BackendMode = &cmd.backendMode
	}
	if cmd.shareToken != "" {
		req.ShareToken = &cmd.shareToken
	}
	if cmd.target != "" {
		req.Target = &cmd.target
	}
	if cmd.permissionMode != "" {
		req.PermissionMode = &cmd.permissionMode
	}

	// boolean filters - only set if flag was explicitly provided
	if cmd.cmd.Flags().Changed("has-activity") {
		val, _ := cmd.cmd.Flags().GetBool("has-activity")
		req.HasActivity = &val
		cmd.hasActivity = &val
	}
	if cmd.cmd.Flags().Changed("idle") {
		val, _ := cmd.cmd.Flags().GetBool("idle")
		req.Idle = &val
		cmd.idle = &val
	}

	// validate that hasActivity and idle are not both set
	if cmd.hasActivity != nil && *cmd.hasActivity && cmd.idle != nil && *cmd.idle {
		fmt.Println("error: cannot use both --has-activity and --idle flags")
		os.Exit(1)
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
	resp, err := zrok.Metadata.ListShares(req, auth)
	if err != nil {
		panic(err)
	}

	shares := resp.Payload.Shares

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
	t.AppendHeader(table.Row{"Share Token", "Frontend Endpoints", "Env", "Mode", "Backend", "Target", "Limited", "Created"})

	for _, share := range shares {
		// format env zid (truncate if too long)
		envZId := share.EnvZID
		if len(envZId) > 12 {
			envZId = envZId[:12] + "..."
		}

		// format target
		target := share.Target
		if target == "" {
			target = "-"
		} else if len(target) > 30 {
			target = target[:30] + "..."
		}

		// format limited status
		limitedStatus := ""
		if share.Limited {
			limitedStatus = "!!"
		}

		// format created timestamp
		created := time.Unix(share.CreatedAt/1000, 0).Format("2006-01-02 15:04:05")

		t.AppendRow(table.Row{
			share.ShareToken,
			strings.Join(share.FrontendEndpoints, "\n"),
			envZId,
			share.ShareMode,
			share.BackendMode,
			target,
			limitedStatus,
			created,
		})
	}

	t.Render()
	fmt.Println()

	// show summary
	if len(shares) == 0 {
		fmt.Println("no shares found matching the specified filters.")
		fmt.Println()
	} else {
		fmt.Printf("total: %d share(s)\n", len(shares))
		fmt.Println()
	}
}
