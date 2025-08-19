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

type UpstreamPool struct {
	upstreams         []*Upstream
	cancel            context.CancelFunc
	context           context.Context
	current           uint64
	healthCheckTries  int
	healthCheckPeriod int
}

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

func (pool *UpstreamPool) Balance(w http.ResponseWriter, r *http.Request) {
	target := pool.Next()
	if target != nil {
		target.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "server not available", http.StatusServiceUnavailable)
}

func (pool *UpstreamPool) Close() {
	// Cancel the backend pool context,
	// this will return any long lived function that depends on it
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

func (pool *UpstreamPool) NextIdx() int {
	return int(atomic.AddUint64(&pool.current, uint64(1)) % uint64(len(pool.upstreams)))
}
