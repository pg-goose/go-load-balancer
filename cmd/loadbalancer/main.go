package main

import (
	"flag"
	"log"

	"github.com/pg-goose/go-load-balancer/loadbalancer"
	lb "github.com/pg-goose/go-load-balancer/loadbalancer"
)

var confpath = flag.String("confpath", "config.yml", "path to configuration file")

func main() {
	config := loadbalancer.LoadConfig(*confpath)
	lb := lb.NewLoadBalancer(config)
	defer lb.Stop()
	log.Fatal(lb.Start())
}
