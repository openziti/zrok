package main

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"os"
)

func mustGetAdminAuth() runtime.ClientAuthInfoWriter {
	adminToken := os.Getenv("ZROK_ADMIN_TOKEN")
	if adminToken == "" {
		panic("please set ZROK_ADMIN_TOKEN to a valid admin token for your zrok instance")
	}
	return httptransport.APIKeyAuth("X-TOKEN", "header", adminToken)
}
