package main

import (
	"log"

	lb "github.com/pg-goose/go-load-balancer/loadbalancer"
)

func main() {
	lb := lb.NewLoadBalancer(&lb.Config{
		Port:              8080,
		Backends:          []string{"http://0.0.0.0:8081"},
		HealthCheckTries:  2,
		HealthCheckPeriod: 5,
	})
	defer lb.Stop()

	log.Fatal(lb.Start())
}
