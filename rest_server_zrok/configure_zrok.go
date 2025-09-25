// This file is safe to edit. Once it exists it will not be overwritten

package rest_server_zrok

import (
	"crypto/tls"
	"net/http"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/ui"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/openziti/zrok/rest_server_zrok/operations"
)

var HealthCheck func(w http.ResponseWriter, r *http.Request)

//go:generate swagger generate server --target ../../zrok --name Zrok --spec ../specs/zrok.yml --model-package rest_model_zrok --server-package rest_server_zrok --principal interface{} --exclude-main

func configureFlags(api *operations.ZrokAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ZrokAPI) http.Handler {
	api.ServeError = errors.ServeError
	api.Logger = func(m string, args ...interface{}) { dl.Infof(m, args...) }
	api.UseSwaggerUI()
	api.JSONConsumer = runtime.JSONConsumer()
	api.JSONProducer = runtime.JSONProducer()
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
	return ui.Middleware(handler, HealthCheck)
}
