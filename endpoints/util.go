package endpoints

import (
	"net/url"
	"strings"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/sdk-golang/ziti"
)

func GetRefreshedService(svcName string, ctx ziti.Context) (*rest_model.ServiceDetail, bool) {
	svc, found := ctx.GetService(svcName)
	if !found {
		svc, err := ctx.RefreshService(svcName)
		if err != nil {
			dl.Errorf("error refreshing service '%v': %v", svcName, err)
			return nil, false
		}
		if svc == nil {
			dl.Errorf("service '%v' not found", svcName)
			return nil, false
		}
		return svc, true
	}
	return svc, found
}

func JoinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
