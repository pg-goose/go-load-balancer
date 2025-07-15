package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	Mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type BackendPool struct {
	backends []*Backend
	current  uint64
}

func NewBackendPool(addrs ...string) *BackendPool {
	bp := &BackendPool{
		backends: make([]*Backend, 0, len(addrs)),
		current:  0,
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

func (bp *BackendPool) NextIdx() int {
	return int(atomic.AddUint64(&bp.current, uint64(1)) % uint64(len(bp.backends)))
}

func (bp *BackendPool) Next() *Backend {
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

func (bp *BackendPool) balance(w http.ResponseWriter, r *http.Request) {
	target := bp.Next()
	if target != nil {
		target.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "server not available", http.StatusServiceUnavailable)
}

var port = flag.Int("port", 8080, "port where the load balancer listens")
var backends = []string{"http://0.0.0.0:8081", "http://0.0.0.0:8082", "http://0.0.0.0:8083"}

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil) // pprof endpoint
	}()

	backendPool := NewBackendPool(backends...)
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: http.HandlerFunc(backendPool.balance),
	}
	for _, backend := range backendPool.backends {
		go func(b *Backend) {
			t := time.NewTicker(time.Second * 5)
			timeout := 2 * time.Second
			host := b.URL.Host
			for range t.C {
				conn, err := net.DialTimeout("tcp", host, timeout)
				if err != nil {
					log.Printf("%s unreachable", host)
					b.Alive = false
					continue
				}
				conn.Close()
				b.Alive = true
			}
		}(backend)
	}
	log.Fatal(server.ListenAndServe())
}
