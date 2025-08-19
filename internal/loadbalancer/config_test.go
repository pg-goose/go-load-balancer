package loadbalancer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	lb "github.com/pg-goose/go-load-balancer/internal/loadbalancer"
)

func TestLoadConfig(t *testing.T) {
	expected := &lb.Config{
		Port: 8080,
		Upstreams: []string{
			"http://10.0.0.1:8080/health",
			"http://10.0.0.2:8080/health",
			"http://10.0.0.3:8080/health",
		},
		HealthCheckTries:  3,
		HealthCheckPeriod: 5,
	}
	c := lb.LoadConfig("testdata/config.test.yml")
	assert.Equal(t, c, expected)
}
