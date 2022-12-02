package main

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pfxlog.GlobalInit(logrus.InfoLevel, pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti-test-kitchen/"))
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&panicInstead, "panic", "p", false, "Panic instead of showing pretty errors")
	zrokdir.AddZrokApiEndpointFlag(&apiEndpoint, rootCmd.PersistentFlags())
	rootCmd.AddCommand(accessCmd)
	adminCmd.AddCommand(adminCreateCmd)
	adminCmd.AddCommand(adminDeleteCmd)
	rootCmd.AddCommand(adminCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(testCmd)
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
var apiEndpoint string

var accessCmd = &cobra.Command{
	Use:   "access",
	Short: "Create frontend access for services",
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

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Create backend access for services",
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Utilities for testing zrok deployments",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
