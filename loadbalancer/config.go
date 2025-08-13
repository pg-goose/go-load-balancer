package loadbalancer

import (
	"log"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Port              int      `yaml:"port"`
	Backends          []string `yaml:"backends"`
	HealthCheckTries  int      `yaml:"healthCheckTries"`
	HealthCheckPeriod int      `yaml:"healthCheckPeriod"`
}

func LoadConfig(p string) *Config {
	data, err := os.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}
	c := &Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		log.Panic(err)
	}
	return c
}
