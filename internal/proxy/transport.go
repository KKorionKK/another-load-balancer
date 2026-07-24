package proxy

import (
	"net"
	"net/http"
	"sld/internal/config"
)

func NewTransport(cfg config.TransportConfig) *http.Transport {
	return &http.Transport{
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		DialContext: (&net.Dialer{
			Timeout: cfg.DialTimeout,
		}).DialContext,
	}
}
