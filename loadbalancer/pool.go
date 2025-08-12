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

type Pool struct {
	backends          []*Backend
	cancel            context.CancelFunc
	context           context.Context
	current           uint64
	healthCheckTries  int
	healthCheckPeriod int
}

func NewPool(config *Config) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &Pool{
		backends:          make([]*Backend, 0, len(config.Backends)),
		current:           0,
		context:           ctx,
		cancel:            cancel,
		healthCheckTries:  config.HealthCheckTries,
		healthCheckPeriod: config.HealthCheckPeriod,
	}
	for _, a := range config.Backends {
		u, err := url.Parse(a)
		if err != nil {
			log.Fatalf("failed to parse backend URL %s: %v", a, err)
		}
		pool.backends = append(pool.backends, &Backend{
			URL:          u,
			ReverseProxy: httputil.NewSingleHostReverseProxy(u),
		})
	}
	return pool
}

func (pool *Pool) Balance(w http.ResponseWriter, r *http.Request) {
	target := pool.Next()
	if target != nil {
		target.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "server not available", http.StatusServiceUnavailable)
}

func (pool *Pool) Close() {
	// Cancel the backend pool context,
	// this will return any long lived function that depends on it
	pool.cancel()
}

// HealthCheck initiates an infinite loop that iterates the registered backends
// every HEALTH_CHECK_PERIOD seconds and dials a tcp connection to them with a
// HEALTH_CHECK_TIMEOUT seconds timeout. The loop ends when the BackendPool context
// is canceled.
func (pool *Pool) HealthCheck() {
	ticker := time.NewTicker(time.Duration(pool.healthCheckPeriod) * time.Second)
	timeout := time.Duration(pool.healthCheckTries) * time.Second

	for range ticker.C {
		for _, b := range pool.backends {
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
}

func (pool *Pool) Next() *Backend {
	next := pool.NextIdx()
	ln := len(pool.backends)
	last := ln + next
	for i := next; i < last; i++ {
		idx := i % ln
		if !pool.backends[idx].Alive {
			continue
		}
		if i != next {
			atomic.StoreUint64(&pool.current, uint64(idx))
		}
		return pool.backends[idx]
	}
	return nil
}

func (pool *Pool) NextIdx() int {
	return int(atomic.AddUint64(&pool.current, uint64(1)) % uint64(len(pool.backends)))
}
