package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
)

func init() {
	agentCmd.AddCommand(newAgentStatusCommand().cmd)
}

type agentStatusCommand struct {
	cmd     *cobra.Command
	verbose bool
}

func newAgentStatusCommand() *agentStatusCommand {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the running zrok Agent",
		Args:  cobra.NoArgs,
	}
	command := &agentStatusCommand{cmd: cmd}
	cmd.Flags().BoolVarP(&command.verbose, "verbose", "v", false, "show verbose failure details")
	cmd.Run = command.run
	return command
}

func (cmd *agentStatusCommand) run(_ *cobra.Command, _ []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		tui.Error("error loading zrokdir", err)
	}

	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer conn.Close()

	status, err := client.Status(context.Background(), &agentGrpc.StatusRequest{})
	if err != nil {
		tui.Error("error getting status", err)
	}

	cmd.displayAccesses(status.GetAccesses())
	cmd.displayShares(status.GetShares())
}

func (cmd *agentStatusCommand) displayAccesses(accesses []*agentGrpc.AccessDetail) {
	if len(accesses) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("ACCESSES")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)

	if cmd.verbose {
		t.AppendHeader(table.Row{"Frontend Token", "Share Token", "Bind Address", "Status", "Failures", "Last Error", "Next Retry"})
	} else {
		t.AppendHeader(table.Row{"Frontend Token", "Share Token", "Bind Address", "Status"})
	}

	for _, access := range accesses {
		status := cmd.formatStatus(access.Status)

		// use failure ID in token column if token is empty (failed items)
		displayToken := access.FrontendToken
		if displayToken == "" && access.Failure != nil {
			displayToken = access.Failure.Id
		}

		if cmd.verbose {
			failureCount := "-"
			if access.Failure != nil {
				failureCount = fmt.Sprintf("%d", access.Failure.Count)
			}

			lastError := ""
			if access.Failure != nil {
				lastError = access.Failure.LastError
			}

			nextRetry := "-"
			if access.Failure != nil {
				nextRetry = fmt.Sprintf("%v", access.Failure.NextRetry.AsTime().Format(time.RFC3339Nano))
			}

			t.AppendRow(table.Row{displayToken, access.Token, access.BindAddress, status, failureCount, cmd.wrapString(lastError, 35), nextRetry})
		} else {
			t.AppendRow(table.Row{displayToken, access.Token, access.BindAddress, status})
		}
	}
	activeAccesses, retryingAccesses, failedAccesses := cmd.categorizeAccesses(accesses)
	t.SetCaption(fmt.Sprintf("%d active, %d retrying, %d failed\n", activeAccesses, retryingAccesses, failedAccesses))

	t.Render()
}

func (cmd *agentStatusCommand) displayShares(shares []*agentGrpc.ShareDetail) {
	if len(shares) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("SHARES")
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)

	if cmd.verbose {
		t.AppendHeader(table.Row{"Share Token", "Share Mode", "Backend Mode", "Target", "Status", "Failures", "Last Error", "Next Retry"})
	} else {
		t.AppendHeader(table.Row{"Share Token", "Share Mode", "Backend Mode", "Target", "Status"})
	}

	for _, share := range shares {
		status := cmd.formatStatus(share.Status)

		// use failure ID in token column if token is empty (failed items)
		displayToken := share.Token
		if displayToken == "" && share.Failure != nil {
			displayToken = share.Failure.Id
		}

		if cmd.verbose {
			failureCount := "-"
			if share.Failure != nil {
				failureCount = fmt.Sprintf("%d", share.Failure.Count)
			}

			lastError := ""
			if share.Failure != nil {
				lastError = share.Failure.LastError
			}

			nextRetry := "-"
			if share.Failure != nil {
				nextRetry = fmt.Sprintf("%v", share.Failure.NextRetry.AsTime().Format(time.RFC3339Nano))
			}

			t.AppendRow(table.Row{
				displayToken,
				share.ShareMode,
				share.BackendMode,
				share.BackendEndpoint,
				status,
				failureCount,
				cmd.wrapString(lastError, 35),
				nextRetry,
			})
		} else {
			t.AppendRow(table.Row{displayToken, share.ShareMode, share.BackendMode, share.BackendEndpoint, status})
		}
	}
	activeShares, retryingShares, failedShares := cmd.categorizeShares(shares)
	t.SetCaption(fmt.Sprintf("%d active, %d retrying, %d failed\n", activeShares, retryingShares, failedShares))
	t.Render()
}

func (cmd *agentStatusCommand) categorizeAccesses(accesses []*agentGrpc.AccessDetail) (active, retrying, failed int) {
	for _, access := range accesses {
		switch access.Status {
		case "active":
			active++
		case "retrying":
			retrying++
		case "failed":
			failed++
		}
	}
	return
}

func (cmd *agentStatusCommand) categorizeShares(shares []*agentGrpc.ShareDetail) (active, retrying, failed int) {
	for _, share := range shares {
		switch share.Status {
		case "active":
			active++
		case "retrying":
			retrying++
		case "failed":
			failed++
		}
	}
	return
}

func (cmd *agentStatusCommand) formatStatus(status string) string {
	switch status {
	case "active":
		return text.FgGreen.Sprint("active")
	case "retrying":
		return text.FgYellow.Sprint("retrying")
	case "failed":
		return text.FgRed.Sprint("failed")
	default:
		return status
	}
}

func (cmd *agentStatusCommand) formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("15:04:05")
}
