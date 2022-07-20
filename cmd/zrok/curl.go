package main

import (
	"context"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/spf13/cobra"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

func init() {
	rootCmd.AddCommand(curlCmd)
}

var curlCmd = &cobra.Command{
	Use:   "curl <identity>",
	Short: "curl a zrok service",
	Run:   curl,
}

func curl(_ *cobra.Command, args []string) {
	zCfg, err := config.NewFromFile(args[0])
	if err != nil {
		panic(err)
	}
	zCtx := ziti.NewContextWithConfig(zCfg)
	zDialContext := zitiDialContext{context: zCtx}
	zTransport := http.DefaultTransport.(*http.Transport).Clone()
	zTransport.DialContext = zDialContext.Dial
	client := &http.Client{Transport: zTransport}
	resp, err := client.Get("http://zrok/")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		panic(err)
	}
}

type zitiDialContext struct {
	context ziti.Context
}

func (dc *zitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	service := strings.Split(addr, ":")[0] // will always get passed host:port
	return dc.context.Dial(service)
}
