package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/pg-goose/go-load-balancer/internal/loadbalancer"
	lb "github.com/pg-goose/go-load-balancer/internal/loadbalancer"
)

const USAGE_STR = `golb - tiny go load balancer

Usage:
  golb [flags]

Flags:`

func main() {
	// custom help message
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, USAGE_STR)
		flag.PrintDefaults()
	}
	confpath := flag.String("confpath", "config.yml", "path to configuration file")
	flag.Parse()

	// start the load balancer
	config := loadbalancer.LoadConfig(*confpath)
	lb := lb.NewLoadBalancer(config)

	defer lb.Stop()
	log.Fatal(lb.Start())
}
