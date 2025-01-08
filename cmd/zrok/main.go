package main

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/cobra-to-md"
	"github.com/openziti/transport/v2"
	"github.com/openziti/transport/v2/tcp"
	"github.com/openziti/transport/v2/udp"
	_ "github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pfxlog.GlobalInit(logrus.InfoLevel, pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/"))
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&panicInstead, "panic", "p", false, "Panic instead of showing pretty errors")
	rootCmd.AddCommand(accessCmd)
	adminCmd.AddCommand(adminCreateCmd)
	adminCmd.AddCommand(adminDeleteCmd)
	adminCmd.AddCommand(adminListCmd)
	adminCmd.AddCommand(adminUpdateCmd)
	testCmd.AddCommand(testCanaryCmd)
	rootCmd.AddCommand(adminCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(modifyCmd)
	organizationCmd.AddCommand(organizationAdminCmd)
	rootCmd.AddCommand(organizationCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(gendoc.NewGendocCmd(rootCmd))
	transport.AddAddressParser(tcp.AddressParser{})
	transport.AddAddressParser(udp.AddressParser{})
}

var rootCmd = &cobra.Command{
	Use:   strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])),
	Short: "zrok",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if verbose {
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

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure your zrok environment",
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
