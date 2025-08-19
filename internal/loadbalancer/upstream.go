package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

// Upstream is a backend server in a load balancing setup.
// It maintains information about the server's health and provides
// functionality for proxying requests to it.
type Upstream struct {
	// URL is the address of the backend server.
	URL *url.URL
	// Alive indicates whether the upstream server is currently operational.
	Alive bool
	// Mux provides synchronization for concurrent access to the Upstream.
	Mux sync.RWMutex
	// ReverseProxy handles forwarding requests to this upstream server.
	ReverseProxy *httputil.ReverseProxy
}

// SetAlive sets the upstream Alive flag in a thread-safe manner.
func (u *Upstream) SetAlive(b bool) {
	u.Mux.Lock()
	u.Alive = b
	u.Mux.Unlock()
}

// IsAlive returns the upstream Alive flag in a thread-safe manner.
func (u *Upstream) IsAlive() bool {
	u.Mux.Lock()
	a := u.Alive
	u.Mux.Unlock()
	return a
}
