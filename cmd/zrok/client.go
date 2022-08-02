package main

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok"
)

func newZrokClient() *rest_client_zrok.Zrok {
	transport := httptransport.New(endpoint, "/api/v1", nil)
	transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
	transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()
	return rest_client_zrok.New(transport, strfmt.Default)
}
