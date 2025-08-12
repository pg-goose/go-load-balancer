package loadbalancer

type Config struct {
	Port              int
	Backends          []string
	HealthCheckTries  int
	HealthCheckPeriod int
}
