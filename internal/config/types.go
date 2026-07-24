package config

import "time"

// BalancerType is the type of load balancer to use
type BalancerType string

const RoundRobin BalancerType = "roundrobin"
const LeastConn BalancerType = "leastconn"

// Config is the main configuration struct for the load balancer
type Config struct {
	// Server is the main server configuration
	Server ServerConfig `yaml:"server"`
	// Balancer is the load balancer configuration
	Balancer BalancerConfig `yaml:"balancer"`
	// Upstreams is the list of upstream servers
	Upstreams []UpstreamConfig `yaml:"upstreams"`
	// HealthCheck is the health check configuration
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	// Transport is the transport configuration
	Transport TransportConfig `yaml:"transport"`
	// Metrics is the metrics configuration for prometheus
	Metrics MetricsConfig `yaml:"metrics"`
	// Logging is the logging configuration
	Logging LoggingConfig `yaml:"logging"`
}

// ServerConfig is the main server configuration
type ServerConfig struct {
	// ListenAddr is the address to listen on
	ListenAddr string `yaml:"listen_addr"`
	// ShutdownTimeout is the timeout for graceful shutdown
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// BalancerConfig is the configuration for the load balancer
type BalancerConfig struct {
	// Strategy is the load balancing strategy to use, e.g. "roundrobin", "leastconn"
	Strategy string `yaml:"strategy"`
}

// UpstreamConfig is the configuration for an upstream server
type UpstreamConfig struct {
	// Name is the name of the upstream server. Used in metrics and logging.
	Name string `yaml:"name"`
	// URL is the URL of the upstream server. e.g. "http://localhost:8080"
	URL string `yaml:"url"`
	// Weight is the weight of the upstream server. Used in load balancing.
	Weight int `yaml:"weight"`
}

// HealthCheckConfig is the configuration for health checking an upstream server
// An independent goroutine that periodically checks the health of the upstream server.
type HealthCheckConfig struct {
	// Path is the path to check for health. e.g. "/health"
	Path string `yaml:"path"`
	// Interval is the time between health checks.
	Interval time.Duration `yaml:"interval"`
	// Timeout is the time to wait for a health check to complete.
	Timeout time.Duration `yaml:"timeout"`
	// HealthyThreshold is the number of consecutive successful health checks required to mark the upstream server as healthy.
	HealthyThreshold int `yaml:"healthy_threshold"`
	// UnhealthyThreshold is the number of consecutive failed health checks required to mark the upstream server as unhealthy.
	UnhealthyThreshold int `yaml:"unhealthy_threshold"`
}

// TransportConfig is the configuration for the transport layer.
type TransportConfig struct {
	// MaxIdleConns is the maximum number of idle connections to keep open. Defaults to 200.
	MaxIdleConns int `yaml:"max_idle_conns"`
	// MaxIdleConnsPerHost is the maximum number of idle connections to keep open per host. Defaults to 50.
	MaxIdleConnsPerHost int `yaml:"max_idle_conns_per_host"`
	// IdleConnTimeout is the time to wait for an idle connection to be reused. Defaults to 90 seconds.
	IdleConnTimeout time.Duration `yaml:"idle_conn_timeout"`
	// DialTimeout is the time to wait for a connection to be established. Defaults to 10 seconds.
	DialTimeout time.Duration `yaml:"dial_timeout"`
	// TLSHandshakeTimeout is the time to wait for a TLS handshake to complete. Defaults to 5 seconds.
	TLSHandshakeTimeout time.Duration `yaml:"tls_handshake_timeout"`
	// ResponseHeaderTimeout is the time to wait for a response header to be received. Defaults to 10 seconds.
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout"`
}

// MetricsConfig is the configuration for the metrics server.
type MetricsConfig struct {
	// Enabled is whether the metrics server is enabled. Defaults to false.
	Enabled bool `yaml:"enabled"`
	// Path is the path to serve the metrics on. Defaults to "/metrics".
	Path string `yaml:"path"`
}

// LoggingConfig is the configuration for logging.
type LoggingConfig struct {
	// Level is the logging level. Defaults to "info".
	Level string `yaml:"level"`
	// Format is the logging format. Defaults to "json".
	Format string `yaml:"format"`
}
