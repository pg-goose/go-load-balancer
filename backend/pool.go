package backend

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
	backends []*Backend
	cancel   context.CancelFunc
	context  context.Context
	current  uint64
}

func NewPool(addrs ...string) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	bp := &Pool{
		backends: make([]*Backend, 0, len(addrs)),
		current:  0,
		context:  ctx,
		cancel:   cancel,
	}
	for _, a := range addrs {
		u, err := url.Parse(a)
		if err != nil {
			log.Fatalf("failed to parse backend URL %s: %v", a, err)
		}
		bp.backends = append(bp.backends, &Backend{
			URL:          u,
			ReverseProxy: httputil.NewSingleHostReverseProxy(u),
		})
	}
	return bp
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

const HEALTH_CHECK_TIMEOUT = 1
const HEALTH_CHECK_PERIOD = 5

// HealthCheck initiates an infinite loop that iterates the registered backends
// every HEALTH_CHECK_PERIOD seconds and dials a tcp connection to them with a
// HEALTH_CHECK_TIMEOUT seconds timeout. The loop ends when the BackendPool context
// is canceled.
//
// - TODO: Make constants variables.
func (pool *Pool) HealthCheck() {
	ticker := time.NewTicker(HEALTH_CHECK_PERIOD * time.Second)
	timeout := HEALTH_CHECK_TIMEOUT * time.Second

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

// TODO if a request fails with a 5xx status code re-check if alive and retry 3 times with 10ms wait
// TODO attempt a retried 5xx request on another server 1 time
// VERY TODO: read configuration for: registering backends, error intervention strategies, set constants, etc
