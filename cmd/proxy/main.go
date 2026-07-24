package main

import (
	"log"
	"net/http"
	"sld/internal"
	"sld/internal/balancer"
	"sld/internal/config"
	"sld/internal/proxy"
)

// func proxyHandler(w http.ResponseWriter, r *http.Request) {
// 	upstream := "http://localhost:9090"
// 	url, err := url.Parse(upstream)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	proxy := httputil.ReverseProxy{
// 		Rewrite: func(pr *httputil.ProxyRequest) {
// 			pr.SetURL(url)
// 			pr.Out.Host = pr.In.Host
// 			pr.SetXForwarded()
// 		},
// 	}
// 	proxy.ServeHTTP(w, r)
// }

// func main() {
// 	http.HandleFunc("/", proxyHandler)
// 	http.ListenAndServe(":8080", nil)
// }

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalln("failed to load config: " + err.Error())
	}

	err = cfg.ValidateConfig()
	if err != nil {
		log.Fatalln("failed to validate config: " + err.Error())
	}

	transport := proxy.NewTransport(cfg.Transport)
	upstreams := internal.BuildUpstreams(cfg.Upstreams, transport)

	var balancerStrategy balancer.Balancer
	if cfg.Balancer.Strategy == string(config.RoundRobin) {
		balancerStrategy = balancer.NewRoundRobinBalancer(upstreams, transport)
	}

	if balancerStrategy == nil {
		log.Fatalln("unsupported balancer strategy: " + cfg.Balancer.Strategy)
	}

	log.Println("starting proxy with balancer strategy: " + cfg.Balancer.Strategy)
	http.ListenAndServe(":8080", balancerStrategy)
}
