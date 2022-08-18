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
		httpListener, err := listen.NewHTTP(&listen.Config{
			IdentityPath: args[0],
			Address:      "0.0.0.0:10111",
		})
		if err != nil {
			panic(err)
		}
		if err := httpListener.Run(); err != nil {
			panic(err)
		}
	},
}
