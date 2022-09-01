package main

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	pfxlog.GlobalInit(logrus.InfoLevel, pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti-test-kitchen/"))
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&apiEndpoint, "endpoint", "e", "api.zrok.io", "zrok API endpoint address")
	rootCmd.AddCommand(httpCmd)
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
var apiEndpoint string

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "HTTP endpoint operations",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func newZrokClient() *rest_client_zrok.Zrok {
	transport := httptransport.New(apiEndpoint, "/api/v1", []string{"https", "http"})
	transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
	transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()
	return rest_client_zrok.New(transport, strfmt.Default)
}
