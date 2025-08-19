package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
)

// A LoadBalancer contains a HTTP server and a pool of upstreams to send requests to.
type LoadBalancer struct {
	server       *http.Server
	upstreamPool *UpstreamPool
}

// NewLoadBalancer creates and configures a new LoadBalancer instance with the provided configuration.
// It sets up an HTTP server on the port specified in the config and initializes the upstream pool.
// The server is configured to use the backendPool's Balance handler for incoming requests.
func NewLoadBalancer(config *Config) *LoadBalancer {
	pool := NewPool(config)
	return &LoadBalancer{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: http.HandlerFunc(pool.Balance),
		},
		upstreamPool: pool,
	}
}

// Url returns the LoadBalancer HTTP address.
func (lb *LoadBalancer) Url() string {
	return lb.server.Addr // TODO fix `lb.server.Addr` only returning :8080
}

// Start begins the load balancer operation by initiating health checks
// on the upstream pool and starting the HTTP server.
// It returns an error if the server fails to start.
func (lb *LoadBalancer) Start() error {
	go lb.upstreamPool.HealthCheck()
	return lb.server.ListenAndServe()
}

// Stop shuts down the load balancer by closing its server and upstream pool.
// It will terminate the program with a fatal error if it fails to close the server.
func (lb *LoadBalancer) Stop() {
	if err := lb.server.Close(); err != nil {
		log.Fatal("failed to close server:", err.Error())
	}
	lb.upstreamPool.Close()
}
