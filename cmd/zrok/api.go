package main

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_client"
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
	Use:   "version <endpoint>",
	Short: "Get API version",
	Run: func(_ *cobra.Command, args []string) {
		transport := httptransport.New(args[0], "", nil)
		transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
		transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()
		zrok := rest_zrok_client.New(transport, strfmt.Default)
		resp, err := zrok.Metadata.Version(metadata.NewVersionParams())
		if err != nil {
			panic(err)
		}
		logrus.Infof("found api version [%v]", resp.Payload.Version)
	},
}
