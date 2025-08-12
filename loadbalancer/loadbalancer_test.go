package loadbalancer_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	lb "github.com/pg-goose/go-load-balancer/loadbalancer"
)

const NB_BACKENDS = 5

func TestLoadBalancer(t *testing.T) {
	urls := []string{}
	servers := []*httptest.Server{}

	for i := range NB_BACKENDS {
		idx := i
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(fmt.Appendf([]byte{}, "server: %d", idx))
		}))
		defer s.Close()
		urls = append(urls, s.URL)
		servers = append(servers, s)
	}
	lb := lb.NewLoadBalancer(&lb.Config{
		Port:              8080,
		Backends:          urls,
		HealthCheckTries:  2,
		HealthCheckPeriod: 5,
	})
	go func() {
		if err := lb.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Error("Load balancer failed to start:", err)
		}
	}()

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
		t.Log(string(r))
	}
	t.FailNow()
	defer lb.Stop()
}
