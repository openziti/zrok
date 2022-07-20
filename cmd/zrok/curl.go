package main

import (
	"github.com/openziti-test-kitchen/zrok/util"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
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
	zDialContext := util.ZitiDialContext{Context: zCtx}
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
