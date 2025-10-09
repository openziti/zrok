package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/zrok/rest_client_zrok/metadata"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util"
	"github.com/spf13/cobra"
)

var overviewCmd *cobra.Command

func init() {
	overviewCmd = newOverviewCommand().cmd
	rootCmd.AddCommand(overviewCmd)
}

type overviewCommand struct {
	cmd  *cobra.Command
	json bool
}

func newOverviewCommand() *overviewCommand {
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Display a formatted overview of zrok account details (namespaces, names, environments, shares)",
		Args:  cobra.ExactArgs(0),
	}
	command := &overviewCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.json, "json", false, "output raw JSON instead of formatted tables")
	cmd.Run = command.run
	return command
}

func (cmd *overviewCommand) run(_ *cobra.Command, _ []string) {
	env, auth := mustGetEnvironmentAuth()

	zrok, err := env.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("error creating zrok client", err)
		}
		panic(err)
	}

	req := metadata.NewOverviewParams()
	resp, err := zrok.Metadata.Overview(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("error getting overview", err)
		}
		panic(err)
	}

	overview := resp.Payload

	// if JSON flag is set, output raw JSON and return
	if cmd.json {
		jsonBytes, err := json.MarshalIndent(overview, "", "  ")
		if err != nil {
			if !panicInstead {
				tui.Error("error marshaling JSON", err)
			}
			panic(err)
		}
		fmt.Println(string(jsonBytes))
		return
	}

	fmt.Println()

	// display account status
	if overview.AccountLimited {
		fmt.Println("!! Account Limited")
		fmt.Println()
	}

	// display namespaces table
	if len(overview.Namespaces) > 0 {
		fmt.Println("* Namespaces")
		fmt.Println()
		namespacesTable := table.NewWriter()
		namespacesTable.SetOutputMirror(os.Stdout)
		namespacesTable.SetStyle(table.StyleRounded)
		namespacesTable.AppendHeader(table.Row{"Name", "Description", "Token"})

		for _, ns := range overview.Namespaces {
			namespacesTable.AppendRow(table.Row{
				ns.Name,
				ns.Description,
				ns.NamespaceToken,
			})
		}
		namespacesTable.Render()
		fmt.Println()
	}

	// display names table
	if len(overview.Names) > 0 {
		fmt.Println("* Names")
		fmt.Println()
		namesTable := table.NewWriter()
		namesTable.SetOutputMirror(os.Stdout)
		namesTable.SetStyle(table.StyleRounded)
		namesTable.AppendHeader(table.Row{"URL", "Namespace Token", "Share Token", "Reserved", "Created"})

		for _, name := range overview.Names {
			url := util.ExpandUrlTemplate(name.Name, name.NamespaceName)
			shareToken := name.ShareToken
			if shareToken == "" {
				shareToken = "-"
			}
			namesTable.AppendRow(table.Row{
				url,
				name.NamespaceToken,
				shareToken,
				name.Reserved,
				time.Unix(name.CreatedAt, 0).Format("2006-01-02 15:04:05"),
			})
		}
		namesTable.Render()
		fmt.Println()
	}

	// display environments and their resources
	if len(overview.Environments) > 0 {
		fmt.Println("* Environments")
		fmt.Println()

		for _, envRes := range overview.Environments {
			env := envRes.Environment
			if env != nil {
				// environment header
				fmt.Println("╔════════════════════════───────────────────────")
				fmt.Printf("> %s (envZId: %s)\n", env.Description, env.ZID)
				if env.Host != "" {
					fmt.Printf("      Host: %s\n", env.Host)
				}
				if env.Address != "" {
					fmt.Println("   Address:", env.Address)
				}
				if env.RemoteAgent {
					fmt.Println("  Remoting: Enabled")
				}
				if env.Limited {
					fmt.Println("   Limited")
				}
				fmt.Printf("   Created: %s\n", time.Unix(env.CreatedAt/1000, 0).Format("2006-01-02 15:04:05"))
				fmt.Println()

				// shares table
				if len(envRes.Shares) > 0 {
					fmt.Println("  > Shares")
					sharesTable := table.NewWriter()
					sharesTable.SetOutputMirror(os.Stdout)
					sharesTable.SetStyle(table.StyleRounded)
					sharesTable.AppendHeader(table.Row{"Share Token", "Mode", "Backend", "Target", "Limited", "Created"})

					for _, share := range envRes.Shares {
						target := share.Target
						if target == "" {
							target = "-"
						}
						limitedIcon := ""
						if share.Limited {
							limitedIcon = "!!"
						}

						sharesTable.AppendRow(table.Row{
							share.ShareToken,
							share.ShareMode,
							share.BackendMode,
							target,
							limitedIcon,
							time.Unix(share.CreatedAt/1000, 0).Format("2006-01-02 15:04:05"),
						})
					}
					sharesTable.Render()

					// display frontend endpoints for each share
					for _, share := range envRes.Shares {
						if len(share.FrontendEndpoints) > 0 {
							fmt.Printf("      > %s:\n", share.ShareToken)
							for _, endpoint := range share.FrontendEndpoints {
								fmt.Printf("         > %s\n", endpoint)
							}
						}
					}
					fmt.Println()
				}

				// frontends table
				if len(envRes.Frontends) > 0 {
					fmt.Println("  > Frontends")
					frontendsTable := table.NewWriter()
					frontendsTable.SetOutputMirror(os.Stdout)
					frontendsTable.SetStyle(table.StyleRounded)
					frontendsTable.AppendHeader(table.Row{"Frontend Token", "Bind Address", "Description", "Created"})

					for _, frontend := range envRes.Frontends {
						bindAddr := frontend.BindAddress
						if bindAddr == "" {
							bindAddr = "-"
						}
						desc := frontend.Description
						if desc == "" {
							desc = "-"
						}
						frontendsTable.AppendRow(table.Row{
							frontend.FrontendToken,
							bindAddr,
							desc,
							time.Unix(frontend.CreatedAt/1000, 0).Format("2006-01-02 15:04:05"),
						})
					}
					frontendsTable.Render()
				}
				fmt.Println("╚════════════════════════───────────────────────")
				fmt.Println()
			}
		}
	}

	if len(overview.Environments) == 0 && len(overview.Names) == 0 && len(overview.Namespaces) == 0 {
		fmt.Println("No environments, namespaces, or names found.")
		fmt.Println("Run 'zrok enable' to set up your first environment.")
		fmt.Println()
	}
}
