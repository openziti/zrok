package zrokdir

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok"
	"github.com/pkg/errors"
	"net/url"
)

func ZrokClient(endpoint string) (*rest_client_zrok.Zrok, error) {
	apiUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing api endpoint '%v'", endpoint)
	}
	transport := httptransport.New(apiUrl.Host, "/api/v1", []string{apiUrl.Scheme})
	transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
	transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()
	return rest_client_zrok.New(transport, strfmt.Default), nil
}
