package main

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti-test-kitchen/zrok/rest_zrok_client"
)

func newZrokClient() *rest_zrok_client.Zrok {
	transport := httptransport.New(endpoint, "", nil)
	transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
	transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()
	return rest_zrok_client.New(transport, strfmt.Default)
}
