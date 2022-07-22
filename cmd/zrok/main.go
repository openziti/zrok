package main

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti-test-kitchen/zrok/controller"
	"github.com/openziti-test-kitchen/zrok/http"
	"github.com/openziti-test-kitchen/zrok/proxy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pfxlog.GlobalInit(logrus.InfoLevel, pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti-test-kitchen/"))
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.AddCommand(controllerCmd)
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(proxyCmd)
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

var controllerCmd = &cobra.Command{
	Use:     "controller <configPath>",
	Short:   "Start a zrok controller",
	Aliases: []string{"ctrl"},
	Run: func(_ *cobra.Command, args []string) {
		if err := controller.Run(&controller.Config{Host: "0.0.0.0", Port: 10888}); err != nil {
			panic(err)
		}
	},
}

var httpCmd = &cobra.Command{
	Use:   "http <identity>",
	Short: "Start an http terminator",
	Run: func(_ *cobra.Command, args []string) {
		if err := http.Run(&http.Config{IdentityPath: args[0]}); err != nil {
			panic(err)
		}
	},
}

var proxyCmd = &cobra.Command{
	Use:   "proxy <configPath>",
	Short: "Start a zrok proxy",
	Run: func(_ *cobra.Command, args []string) {
		if err := proxy.Run(&proxy.Config{IdentityPath: args[0], Address: "0.0.0.0:10111"}); err != nil {
			panic(err)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
