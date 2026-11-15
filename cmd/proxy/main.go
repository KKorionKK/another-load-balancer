package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sld/internal"
	"sld/internal/balancer"
	"sld/internal/config"
	"sld/internal/proxy"
	"syscall"
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

	transport := proxy.NewTransport(cfg.Transport)
	upstreams := internal.BuildUpstreams(cfg.Upstreams, transport)

	var balancerStrategy balancer.Balancer
	if cfg.Balancer.Strategy == string(config.RoundRobin) {
		balancerStrategy = balancer.NewRoundRobinBalancer(upstreams, transport)
	}

	if balancerStrategy == nil {
		log.Fatalln("unsupported balancer strategy: " + cfg.Balancer.Strategy)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.ListenAddr),
		Handler: balancerStrategy,
	}

	go func() {
		<-ctx.Done()
		log.Println("graceful shutdown")
		server.Shutdown(ctx)
	}()

	log.Printf("starting proxy with balancer strategy %s and addr %s", cfg.Balancer.Strategy, server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("listendAndServer: %v", err)
	}
}
