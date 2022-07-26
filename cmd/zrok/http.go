package main

import (
	"github.com/openziti-test-kitchen/zrok/http"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/tunnel"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/spf13/cobra"
)

var httpCmd = &cobra.Command{
	Use:   "http <endpoint>",
	Short: "Start an http terminator",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		idCfg, err := zrokdir.IdentityFile()
		if err != nil {
			panic(err)
		}
		cfg := &http.Config{
			IdentityPath:    idCfg,
			EndpointAddress: args[0],
		}
		token, err := zrokdir.ReadToken()
		if err != nil {
			panic(err)
		}

		zrok := newZrokClient()
		req := tunnel.NewTunnelParams()
		req.Body = &rest_model_zrok.TunnelRequest{
			Endpoint: cfg.EndpointAddress,
			Token:    token,
		}
		resp, err := zrok.Tunnel.Tunnel(req)
		if err != nil {
			panic(err)
		}
		cfg.Service = resp.Payload.Service

		if err := http.Run(cfg); err != nil {
			panic(err)
		}
	},
}
