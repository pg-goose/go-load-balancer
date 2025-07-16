package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/pg-goose/go-load-balancer/backend"
)

var port = flag.Int("port", 8080, "port where the load balancer listens")
var backends = []string{"http://0.0.0.0:8081", "http://0.0.0.0:8082", "http://0.0.0.0:8083"}

func main() {
	backendPool := backend.NewPool(backends...)
	defer backendPool.Close()

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: http.HandlerFunc(backendPool.Balance),
	}

	go backendPool.HealthCheck()
	log.Fatal(server.ListenAndServe())
}
