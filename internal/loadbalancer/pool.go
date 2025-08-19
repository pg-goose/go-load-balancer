package loadbalancer

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

// UpstreamPool represents a pool of upstream servers used by a load balancer.
// It manages a collection of upstream servers and handles their health checking
// to ensure traffic is only directed to healthy instances. The pool keeps track
// of each server's health status by performing periodic health checks.
type UpstreamPool struct {
	upstreams         []*Upstream
	cancel            context.CancelFunc
	context           context.Context
	current           uint64
	healthCheckTries  int
	healthCheckPeriod int
}

// NewPool creates a new UpstreamPool based on the provided configuration.
// It initializes the pool with the upstream servers specified in config,
// setting up reverse proxies for each one. The pool is ready to handle
// traffic using the configured health check parameters.
//
// If any upstream URL in the configuration cannot be parsed, the function
// will log a fatal error and terminate the program.
func NewPool(config *Config) *UpstreamPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &UpstreamPool{
		upstreams:         make([]*Upstream, 0, len(config.Upstreams)),
		current:           0,
		context:           ctx,
		cancel:            cancel,
		healthCheckTries:  config.HealthCheckTries,
		healthCheckPeriod: config.HealthCheckPeriod,
	}
	for _, a := range config.Upstreams {
		u, err := url.Parse(a)
		if err != nil {
			log.Fatalf("failed to parse backend URL %s: %v", a, err)
		}
		pool.upstreams = append(pool.upstreams, &Upstream{
			URL:          u,
			ReverseProxy: httputil.NewSingleHostReverseProxy(u),
		})
	}
	return pool
}

// Balance routes the incoming HTTP request to the next available upstream server
// in the pool using a reverse proxy. If no servers are available, it returns
// an HTTP 503 Service Unavailable error.
func (pool *UpstreamPool) Balance(w http.ResponseWriter, r *http.Request) {
	target := pool.Next()
	if target != nil {
		target.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "server not available", http.StatusServiceUnavailable)
}

// Close releases resources held by the upstream pool by canceling the internal context.
// This will terminate any long-running operations that depend on the pool's context.
func (pool *UpstreamPool) Close() {
	pool.cancel()
}

// HealthCheck initiates an infinite loop that iterates the registered backends
// every HEALTH_CHECK_PERIOD seconds and dials a tcp connection to them with a
// HEALTH_CHECK_TIMEOUT seconds timeout. The loop ends when the BackendPool context
// is canceled.
func (pool *UpstreamPool) HealthCheck() {
	timeout := time.Duration(pool.healthCheckTries) * time.Second
	check := func() {
		for _, b := range pool.upstreams {
			host := b.URL.Host
			conn, err := net.DialTimeout("tcp", host, timeout)
			conn.Close()
			if err != nil {
				log.Printf("%s unreachable", host)
			}
			b.Alive = (err == nil) // TODO definir IsAlive con mutex
		}
		select {
		case <-pool.context.Done():
			return
		default:
		}
	}
	// perform a first check to set the state of all backends and avoid ticker delay
	check()
	ticker := time.NewTicker(time.Duration(pool.healthCheckPeriod) * time.Second)
	for range ticker.C {
		check()
	}
}

// Next returns the next available upstream server from the pool.
// It iterates through the list starting from the next index, looking for
// an alive upstream. If an alive upstream is found, it returns that upstream
// and updates the current index if necessary. If no alive upstream is found
// after checking all upstreams, it returns nil.
func (pool *UpstreamPool) Next() *Upstream {
	next := pool.NextIdx()
	ln := len(pool.upstreams)
	last := ln + next
	for i := next; i < last; i++ {
		idx := i % ln
		if !pool.upstreams[idx].Alive {
			continue
		}
		if i != next {
			atomic.StoreUint64(&pool.current, uint64(idx))
		}
		return pool.upstreams[idx]
	}
	return nil
}

// NextIdx returns the next index in the upstream pool in a round-robin fashion.
// It atomically increments the current counter and returns its value modulo the
// number of upstreams, ensuring thread-safe distribution of requests across
// the available upstream servers.
func (pool *UpstreamPool) NextIdx() int {
	return int(atomic.AddUint64(&pool.current, uint64(1)) % uint64(len(pool.upstreams)))
}
