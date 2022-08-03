// This file is safe to edit. Once it exists it will not be overwritten

package rest_server_zrok

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/identity"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/metadata"
	"github.com/openziti-test-kitchen/zrok/rest_server_zrok/operations/tunnel"
)

//go:generate swagger generate server --target ../../zrok --name Zrok --spec ../specs/zrok.yml --model-package rest_model_zrok --server-package rest_server_zrok --principal rest_model_zrok.Principal --exclude-main

func configureFlags(api *operations.ZrokAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ZrokAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "x-token" header is set
	if api.KeyAuth == nil {
		api.KeyAuth = func(token string) (*rest_model_zrok.Principal, error) {
			return nil, errors.NotImplemented("api key auth (key) x-token from header param [x-token] has not yet been implemented")
		}
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	if api.IdentityCreateAccountHandler == nil {
		api.IdentityCreateAccountHandler = identity.CreateAccountHandlerFunc(func(params identity.CreateAccountParams) middleware.Responder {
			return middleware.NotImplemented("operation identity.CreateAccount has not yet been implemented")
		})
	}
	if api.IdentityEnableHandler == nil {
		api.IdentityEnableHandler = identity.EnableHandlerFunc(func(params identity.EnableParams, principal *rest_model_zrok.Principal) middleware.Responder {
			return middleware.NotImplemented("operation identity.Enable has not yet been implemented")
		})
	}
	if api.MetadataListEnvironmentsHandler == nil {
		api.MetadataListEnvironmentsHandler = metadata.ListEnvironmentsHandlerFunc(func(params metadata.ListEnvironmentsParams, principal *rest_model_zrok.Principal) middleware.Responder {
			return middleware.NotImplemented("operation metadata.ListEnvironments has not yet been implemented")
		})
	}
	if api.IdentityLoginHandler == nil {
		api.IdentityLoginHandler = identity.LoginHandlerFunc(func(params identity.LoginParams) middleware.Responder {
			return middleware.NotImplemented("operation identity.Login has not yet been implemented")
		})
	}
	if api.TunnelTunnelHandler == nil {
		api.TunnelTunnelHandler = tunnel.TunnelHandlerFunc(func(params tunnel.TunnelParams, principal *rest_model_zrok.Principal) middleware.Responder {
			return middleware.NotImplemented("operation tunnel.Tunnel has not yet been implemented")
		})
	}
	if api.TunnelUntunnelHandler == nil {
		api.TunnelUntunnelHandler = tunnel.UntunnelHandlerFunc(func(params tunnel.UntunnelParams, principal *rest_model_zrok.Principal) middleware.Responder {
			return middleware.NotImplemented("operation tunnel.Untunnel has not yet been implemented")
		})
	}
	if api.MetadataVersionHandler == nil {
		api.MetadataVersionHandler = metadata.VersionHandlerFunc(func(params metadata.VersionParams) middleware.Responder {
			return middleware.NotImplemented("operation metadata.Version has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
