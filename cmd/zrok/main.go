package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelquigley/df/dl"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/cobra-to-md"
	"github.com/openziti/transport/v2"
	"github.com/openziti/transport/v2/tcp"
	"github.com/openziti/transport/v2/udp"
	_ "github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const trimPrefix = "github.com/openziti/"

func init() {
	// dd/dl Logging
	dl.Init(dl.DefaultOptions().SetTrimPrefix(trimPrefix).SetLevel(slog.LevelInfo))
	dl.ConfigureChannel("mappings", dl.DefaultOptions().SetTrimPrefix(trimPrefix).SetLevel(slog.LevelInfo))

	// legacy pfxlog
	pfxlog.GlobalInit(logrus.InfoLevel, pfxlog.DefaultOptions().SetTrimPrefix(trimPrefix))

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&panicInstead, "panic", "p", false, "Panic instead of showing pretty errors")
	rootCmd.AddCommand(accessCmd)
	adminCmd.AddCommand(adminCreateCmd)
	adminCmd.AddCommand(adminDeleteCmd)
	adminCmd.AddCommand(adminListCmd)
	adminCmd.AddCommand(adminUpdateCmd)
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentReleaseCmd)
	agentCmd.AddCommand(agentShareCmd)
	rootCmd.AddCommand(adminCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(modifyCmd)
	organizationCmd.AddCommand(organizationAdminCmd)
	rootCmd.AddCommand(organizationCmd)
	rootCmd.AddCommand(rebaseCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(testCmd)
	testCmd.AddCommand(testCanaryCmd)
	rootCmd.AddCommand(gendoc.NewGendocCmd(rootCmd))
	transport.AddAddressParser(tcp.AddressParser{})
	transport.AddAddressParser(udp.AddressParser{})
}

var rootCmd = &cobra.Command{
	Use:   strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])),
	Short: "zrok",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if verbose {
			dl.Init(dl.DefaultOptions().SetTrimPrefix(trimPrefix).SetLevel(slog.LevelInfo))
			logrus.SetLevel(logrus.DebugLevel)
		}
	},
}
var verbose bool
var panicInstead bool

var accessCmd = &cobra.Command{
	Use:   "access",
	Short: "Create frontend access for shares",
}

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administration and operations functions",
}

var adminCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create global resources",
}

var adminDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete global resources",
}

var adminListCmd = &cobra.Command{
	Use:   "list",
	Short: "List global resources",
}

var adminUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update global resources",
}

var agentCmd = &cobra.Command{
	Use:     "agent",
	Short:   "zrok Agent commands",
	Aliases: []string{"daemon"},
}

var agentReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "zrok Agent release commands",
}

var agentShareCmd = &cobra.Command{
	Use:   "share",
	Short: "zrok Agent share commands",
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure your zrok environment",
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources",
}

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"rm"},
	Short:   "Delete resources",
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List resources",
}

var modifyCmd = &cobra.Command{
	Use:     "modify",
	Aliases: []string{"mod"},
	Short:   "Modify resources",
}

var organizationAdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Organization admin commands",
}

var organizationCmd = &cobra.Command{
	Use:     "organization",
	Aliases: []string{"org"},
	Short:   "Organization commands",
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "Rebase enabled zrok environment",
}

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Create backend access for shares",
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Utilities for testing deployments",
}

var testCanaryCmd = &cobra.Command{
	Use:   "canary",
	Short: "Utilities for performance management",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		if panicInstead {
			panic(err)
		}
		tui.Error("an error occurred", err)
	}
}
