package proxy

import (
	"fmt"
	"net/http"
)

func Run(cfg *Config) error {
	return http.ListenAndServe(cfg.Address, &handler{})
}

type handler struct{}

func (self *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "zrok")
}
