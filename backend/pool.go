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
	ctx      context.Context
	current  uint64
}

func NewPool(addrs ...string) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	bp := &Pool{
		backends: make([]*Backend, 0, len(addrs)),
		current:  0,
		ctx:      ctx,
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

func (bp *Pool) Balance(w http.ResponseWriter, r *http.Request) {
	target := bp.Next()
	if target != nil {
		target.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "server not available", http.StatusServiceUnavailable)
}

func (bp *Pool) Close() {
	// Cancel the backend pool context,
	// this will return any long lived function that depends on it
	bp.cancel()
}

const HEALTH_CHECK_TIMEOUT = 1
const HEALTH_CHECK_PERIOD = 5

// HealthCheck initiates an infinite loop that iterates the registered backends
// every HEALTH_CHECK_PERIOD seconds and dials a tcp connection to them with a
// HEALTH_CHECK_TIMEOUT seconds timeout. The loop ends when the BackendPool context
// is canceled.
//
// - TODO: Make constants variables.
func (bp *Pool) HealthCheck() {
	ticker := time.NewTicker(HEALTH_CHECK_PERIOD * time.Second)
	timeout := HEALTH_CHECK_TIMEOUT * time.Second

	for range ticker.C {
		for _, b := range bp.backends {
			host := b.URL.Host
			conn, err := net.DialTimeout("tcp", host, timeout)
			conn.Close()
			if err != nil {
				log.Printf("%s unreachable", host)
			}
			b.Alive = (err == nil)
		}
		select {
		case <-bp.ctx.Done():
			return
		default:
		}
	}
}

func (bp *Pool) Next() *Backend {
	next := bp.NextIdx()
	ln := len(bp.backends)
	last := ln + next
	for i := next; i < last; i++ {
		idx := i % ln
		if !bp.backends[idx].Alive {
			continue
		}
		if i != next {
			atomic.StoreUint64(&bp.current, uint64(idx))
		}
		return bp.backends[idx]
	}
	return nil
}

func (bp *Pool) NextIdx() int {
	return int(atomic.AddUint64(&bp.current, uint64(1)) % uint64(len(bp.backends)))
}

// TODO if a request fails with a 5xx status code re-check if alive and retry 3 times with 10ms wait
// TODO attempt a retried 5xx request on another server 1 time
// VERY TODO: configurable error interception?
