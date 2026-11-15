package proxy

import (
	"net/http"
	"net/http/httputil"
)

// New creates a new ReverseProxy with the given transport and error handler
func New(transport http.RoundTripper, onError func(error)) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetXForwarded()
		},
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			onError(err)
			http.Error(w, "bad gateway", http.StatusBadGateway)
		},
	}
}
