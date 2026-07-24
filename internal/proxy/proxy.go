package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func New(target *url.URL, transport http.RoundTripper, onError func(error)) *httputil.ReverseProxy {
	p := httputil.NewSingleHostReverseProxy(target)
	p.Transport = transport
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		onError(err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}
	return p
}
