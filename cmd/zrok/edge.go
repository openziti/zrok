package main

import (
	"crypto/x509"
	"github.com/openziti/edge/rest_management_api_client/identity"
	"github.com/openziti/edge/rest_util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	edgeCmd.AddCommand(edgePingCmd)
	edgeCmd.PersistentFlags().StringVarP(&edgeCmdUsername, "username", "u", "admin", "Edge API username")
	edgeCmd.PersistentFlags().StringVarP(&edgeCmdPassword, "password", "p", "admin", "Edge API password")
	rootCmd.AddCommand(edgeCmd)
}

var edgeCmd = &cobra.Command{
	Use:   "edge",
	Short: "Exercise the edge management API",
}
var edgeCmdUsername string
var edgeCmdPassword string

var edgePingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Validate edge management connectivity",
	Run: func(_ *cobra.Command, args []string) {
		ctrlAddress := "https://linux:1280"
		caCerts, err := rest_util.GetControllerWellKnownCas(ctrlAddress)
		if err != nil {
			panic(errors.Wrap(err, "error getting cas"))
		}
		caPool := x509.NewCertPool()
		for _, ca := range caCerts {
			caPool.AddCert(ca)
		}
		client, err := rest_util.NewEdgeManagementClientWithUpdb(edgeCmdUsername, edgeCmdPassword, ctrlAddress, caPool)
		if err != nil {
			panic(err)
		}

		resp, err := client.Identity.ListIdentities(identity.NewListIdentitiesParams(), nil)
		if err != nil {
			panic(err)
		}
		for _, id := range resp.Payload.Data {
			if id.Name != nil {
				logrus.Infof("identity = '%s'", *id.Name)
			}
		}
	},
}
