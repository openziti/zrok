package main

import (
	"github.com/openziti-test-kitchen/zrok/rest_zrok_client/metadata"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	apiCmd.AddCommand(apiVersionCmd)
	rootCmd.AddCommand(apiCmd)
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Exercise API calls",
}

var apiVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get API version",
	Run: func(_ *cobra.Command, args []string) {
		zrok := newZrokClient()
		resp, err := zrok.Metadata.Version(metadata.NewVersionParams())
		if err != nil {
			panic(err)
		}
		logrus.Infof("found api version [%v]", resp.Payload.Version)
	},
}
