package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

const (
	MIN_ID = 0
	MAX_ID = 9
)

var idFlag = flag.Int("id", 0, "server id. value between 0-9")

func main() {
	flag.Parse()
	id := *idFlag
	if id < MIN_ID || id > MAX_ID {
		log.Fatalf("provided server id %d out of bounds", id)
	}
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("server %d ", id))

	http.HandleFunc("GET /", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(fmt.Sprintf("server %d: ok", id)))
		},
	))

	p := fmt.Sprintf(":808%d", id)
	log.Printf("listening at: %s", p)
	log.Fatal(
		http.ListenAndServe(p, nil),
	)
}

// go run main.go -id 1 &
// go run main.go -id 2 &
// go run main.go -id 3 &
