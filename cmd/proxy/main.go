package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"sld/internal"
	"sld/internal/balancer"
	"sld/internal/config"
	"sld/internal/proxy"
	"sld/internal/retry"
	"syscall"
	"time"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalln("failed to load config: " + err.Error())
	}

	err = cfg.ValidateConfig()
	if err != nil {
		log.Fatalln("failed to validate config: " + err.Error())
	}

	upstreams := internal.BuildUpstreams(cfg.Upstreams)

	var balancerStrategy balancer.Balancer
	if cfg.Balancer.Strategy == string(config.RoundRobin) {
		balancerStrategy = balancer.NewRoundRobinBalancer(upstreams)
	}

	if balancerStrategy == nil {
		log.Fatalln("unsupported balancer strategy: " + cfg.Balancer.Strategy)
	}

	transport := proxy.NewTransport(cfg.Transport)
	// TODO: make customizable
	backoffer := retry.NewExponentialBackoffWithJitter(0.025, 0.25, 1)
	policy := retry.NewBackoffPolicy(backoffer, 5)

	retrying := proxy.NewRetryingTransport(transport, balancerStrategy, policy)
	reverseProxy := proxy.New(retrying, func(err error) {
		log.Printf("proxy error: %s", err)
	})

	server := http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: reverseProxy,
	}

	go func() {
		log.Printf("starting proxy with balancer strategy %s and addr %s", cfg.Balancer.Strategy, server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("listendAndServer: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

}
