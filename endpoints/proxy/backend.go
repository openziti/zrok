package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/corazawaf/coraza/v3"
	txhttp "github.com/corazawaf/coraza/v3/http"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
)

type BackendConfig struct {
	IdentityPath    string
	EndpointAddress string
	ShrToken        string
	Insecure        bool
	Requests        chan *endpoints.Request
	SuperNetwork    bool
}

type Backend struct {
	cfg      *BackendConfig
	listener edge.Listener
	handler  http.Handler
	waf      coraza.WAF
}

func NewBackend(cfg *BackendConfig) (*Backend, error) {
	options := ziti.ListenOptions{
		ConnectTimeout:               5 * time.Minute,
		WaitForNEstablishedListeners: 1,
	}
	zcfg, err := ziti.NewConfigFromFile(cfg.IdentityPath)
	if err != nil {
		return nil, errors.Wrap(err, "error loading config")
	}
	if cfg.SuperNetwork {
		util.EnableSuperNetwork(zcfg)
	}
	zctx, err := ziti.NewContext(zcfg)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ziti context")
	}
	listener, err := zctx.ListenWithOptions(cfg.ShrToken, &options)
	if err != nil {
		return nil, errors.Wrap(err, "error listening")
	}

	proxy, err := newReverseProxy(cfg)
	if err != nil {
		return nil, err
	}

	requestsHandler := util.NewRequestsWrapper(proxy)

	b := &Backend{
		cfg:      cfg,
		listener: listener,
	}

	waf, err := createWaf(b)
	if err != nil {
		return nil, err
	}
	b.waf = waf

	b.handler = txhttp.WrapHandler(waf, requestsHandler)

	return b, nil
}

func createWaf(b *Backend) (coraza.WAF, error) {
	directivesFile := "./default.conf"
	if s := os.Getenv("ZROK_WAF_DIRECTIVES_FILE"); s != "" {
		directivesFile = s
	}

	waf, err := coraza.NewWAF(coraza.NewWAFConfig().WithErrorCallback(b.logError).WithDirectivesFromFile(directivesFile))
	if err != nil {
		return nil, err
	}

	return waf, nil
}

func (b *Backend) Run() error {
	if err := http.Serve(b.listener, b.handler); err != nil {
		return err
	}
	return nil
}

func (b *Backend) Stop() error {
	return b.listener.Close()
}

func (b *Backend) logError(error types.MatchedRule) {
	dl.Errorf("WAF error: %v", error.ErrorLog())
}

func newReverseProxy(cfg *BackendConfig) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(cfg.EndpointAddress)
	if err != nil {
		return nil, err
	}

	tpt := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.Insecure {
		tpt.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = tpt
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		if cfg.Requests != nil {
			cfg.Requests <- &endpoints.Request{
				Stamp:      time.Now(),
				RemoteAddr: fmt.Sprintf("%v", req.Header["X-Real-Ip"]),
				Method:     req.Method,
				Path:       req.URL.String(),
			}
		}
		director(req)
		req.Header.Set("X-Proxy", "zrok")
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		dl.Errorf("error proxying: %v", err)
		w.WriteHeader(http.StatusBadGateway)
	}

	return proxy, nil
}
