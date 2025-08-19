package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Upstream struct {
	URL          *url.URL
	Alive        bool
	Mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}
