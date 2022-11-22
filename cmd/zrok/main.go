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
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(shareCmd)
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

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "HTTP endpoint operations",
}

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share resources",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
