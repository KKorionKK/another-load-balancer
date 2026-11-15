package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"go.yaml.in/yaml/v4"
)

// ParseConfig parses the configuration file from the CONFIG_PATH environment variable. Applies default values if not set.
func ParseConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		return nil, errors.New("CONFIG_PATH is not set")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.New("failed to read config file: " + err.Error())
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.New("failed to parse YAML: " + err.Error())
	}

	config.ApplyDefaults()

	return &config, nil
}

// ApplyDefaults applies default values to the configuration.
func (cfg *Config) ApplyDefaults() {
	// health check defaults
	if cfg.HealthCheck.Interval == 0 {
		cfg.HealthCheck.Interval = 3 * time.Second
	}
	if cfg.HealthCheck.Timeout == 0 {
		cfg.HealthCheck.Timeout = 2 * time.Second
	}

	if cfg.HealthCheck.HealthyThreshold == 0 {
		cfg.HealthCheck.HealthyThreshold = 2
	}
	if cfg.HealthCheck.UnhealthyThreshold == 0 {
		cfg.HealthCheck.UnhealthyThreshold = 3
	}

	// transport defaults
	if cfg.Transport.MaxIdleConns == 0 {
		cfg.Transport.MaxIdleConns = 200
	}
	if cfg.Transport.MaxIdleConnsPerHost == 0 {
		cfg.Transport.MaxIdleConnsPerHost = 50
	}
	if cfg.Transport.IdleConnTimeout == 0 {
		cfg.Transport.IdleConnTimeout = 90 * time.Second
	}
	if cfg.Transport.DialTimeout == 0 {
		cfg.Transport.DialTimeout = 10 * time.Second
	}
	if cfg.Transport.TLSHandshakeTimeout == 0 {
		cfg.Transport.TLSHandshakeTimeout = 5 * time.Second
	}
	if cfg.Transport.ResponseHeaderTimeout == 0 {
		cfg.Transport.ResponseHeaderTimeout = 10 * time.Second
	}

	// logging defaults
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}

	// metrics defaults
	if cfg.Metrics.Enabled {
		if cfg.Metrics.Path == "" {
			cfg.Metrics.Path = "/metrics"
		}
	}

	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = "8080"
	}

}

// ValidateConfig validates the configuration.
func (cfg *Config) ValidateConfig() error {
	if len(cfg.Upstreams) == 0 {
		return errors.New("no upstreams configured")
	}

	seen := make(map[string]bool)
	for _, u := range cfg.Upstreams {
		if _, err := url.Parse(u.URL); err != nil {
			return fmt.Errorf("upstream %s: invalid url: %w", u.Name, err)
		}
		if seen[u.Name] {
			return fmt.Errorf("duplicate upstream name: %s", u.Name)
		}
		seen[u.Name] = true
	}
	switch cfg.Balancer.Strategy {
	case string(LeastConn), string(RoundRobin):
	default:
		return fmt.Errorf("unknown balancer strategy: %s", cfg.Balancer.Strategy)
	}
	return nil
}
