package main

import (
	"github.com/openziti-test-kitchen/zrok/endpoints/listen"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(proxyCmd)
}

var proxyCmd = &cobra.Command{
	Use:   "proxy <configPath>",
	Short: "Start a zrok proxy",
	Run: func(_ *cobra.Command, args []string) {
		if err := listen.Run(&listen.Config{IdentityPath: args[0], Address: "0.0.0.0:10111"}); err != nil {
			panic(err)
		}
	},
}
