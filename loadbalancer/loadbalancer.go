package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
)

type LoadBalancer struct {
	server      *http.Server
	backendPool *Pool
}

func NewLoadBalancer(config *Config) *LoadBalancer {
	backendPool := NewPool(config)
	return &LoadBalancer{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: http.HandlerFunc(backendPool.Balance),
		},
		backendPool: backendPool,
	}
}

func (lb *LoadBalancer) Url() string {
	return lb.server.Addr // TODO fix `lb.server.Addr` only returning :8080
}

func (lb *LoadBalancer) Start() error {
	go lb.backendPool.HealthCheck()
	return lb.server.ListenAndServe()
}

func (lb *LoadBalancer) Stop() {
	if err := lb.server.Close(); err != nil {
		log.Fatal("failed to close server:", err.Error())
	}
	lb.backendPool.Close()
}
