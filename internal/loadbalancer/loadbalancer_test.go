package loadbalancer_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	lb "github.com/pg-goose/go-load-balancer/internal/loadbalancer"
	"github.com/stretchr/testify/assert"
)

const NB_UPSTREAMS = 5

func TestLoadBalancer(t *testing.T) {
	urls := []string{}
	servers := []*httptest.Server{}
	expected := []string{}

	for i := range NB_UPSTREAMS {
		idx := i
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf([]byte{}, "server: %d", idx))
		}))
		defer s.Close()
		urls = append(urls, s.URL)
		servers = append(servers, s)
		expected = append(expected, fmt.Sprintf("server: %d", idx))
	}
	lb := lb.NewLoadBalancer(&lb.Config{
		Port:               8080,
		Upstreams:          urls,
		HealthCheckTimeout: 2,
		HealthCheckPeriod:  5,
	})
	go func() {
		if err := lb.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Error("Load balancer failed to start:", err)
		}
	}()
	// let the lb health check all backends
	time.Sleep(time.Millisecond * time.Duration(len(servers)))

	results := []string{}
	for range len(urls) {
		resp, err := http.Get("http://localhost" + lb.Url()) // TODO fix lb.Url() only returning :8080
		if err != nil {
			t.Fatal(err.Error())
		}
		r, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err.Error())
		}
		results = append(results, string(r))
	}
	slices.Sort(results) // sort them since the order is not important, easier to compare
	assert.Equal(t, expected, results)
	lb.Stop()
}
