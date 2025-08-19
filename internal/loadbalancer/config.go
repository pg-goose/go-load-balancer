package loadbalancer

import (
	"log"
	"os"

	"go.yaml.in/yaml/v4"
)

// Config is the load balancer configuration, can be parsed from YAML.
type Config struct {
	// Port is the TCP port the load balancer listens on.
	Port int `yaml:"port"`

	// Upstreams is the list of backend targets the load balancer will proxy to.
	// Each entry is typically a host:port or URL.
	Upstreams []string `yaml:"upstreams"`

	// HealthCheckTimeout is the timeout time before considering a backend unhealthy.
	HealthCheckTimeout int `yaml:"healthCheckTimeout"`

	// HealthCheckPeriod is the interval between health checks, in seconds.
	HealthCheckPeriod int `yaml:"healthCheckPeriod"`
}

// LoadConfig reads a YAML configuration file from the given path p,
// unmarshals it into a Config, and returns the populated struct.
// On read errors it terminates the process via log.Fatal;
// on unmarshal errors it panics via log.Panic.
func LoadConfig(p string) *Config {
	data, err := os.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}
	c := &Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		log.Fatal(err)
	}
	return c
}
