package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])),
	Short: "zrok loopback harness",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
	},
}
var verbose bool
var apiEndpoint string

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
