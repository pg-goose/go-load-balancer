package backend

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	Mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}
