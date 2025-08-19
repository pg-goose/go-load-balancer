# golb - Go Load Balancer

> A tiny, HTTP reverse proxy written in Go. Configured via YAML. Implements a **round‑robin** strategy.

I built this to learn about load balancers and reverse proxies and keep my Go skills fresh.

---

## Features

* **YAML configuration** for listener and upstreams.
* **Stateless selection** using an atomic counter (concurrency‑safe).
* **Round‑robin** load balancing across a list of upstreams.
* **Concurrent healthchecks** using goroutines.

---

## How it works

1. The server listens on a configured address (e.g. `:8080`).
2. Each incoming request is forwarded to the next upstream (`(cnt++) % len(upstreams)`).
3. The reverse proxy streams the response back to the client.
4. Failures propagate back to the client (minimal handling).

---

## Configuration

Create a `config.yaml` like this:

```yaml
# config.yaml
port: 8080
healthCheckTimeout: 3
healthCheckPeriod: 5
backends:
    - http://10.0.0.1:8080/health
    - http://10.0.0.2:8080/health
    - http://10.0.0.3:8080/health
```

## Usage
### golb

```
golb - Go Load Balancer

Usage:
  golb [flags]

Flags:
  -confpath string
        path to configuration file (default "config.yml")
  -h, -help
        show this help message

Examples:
  # Run with config.yaml in the current directory
  golb -confpath ./config.yaml

  # Run with the default name in the working directory
  golb

Notes:
  The YAML config controls the listen port, health checks, and backend list.
  See the "Configuration" section for a sample file.
```

### toy-backend

```
toy-backend - Minimal HTTP server used to demo golb

Usage:
  toy-backend [flags]

Flags:
  -id int
        server id. value between 0-9 (default 0)
  -h, -help
        show this help message

Examples:
  # Start three distinct backends in separate terminals (adjust ports as implemented)
  toy-backend -id 1
  toy-backend -id 2
  toy-backend -id 3

Notes:
  The handler responds with a simple payload that includes the server id
  and exposes a /health endpoint for health checks.
```

> 
> ```bash
> for i in {1..6}; do curl -s http://localhost:8080 | head -n1; done
> ```

---

## Project layout

```
.
├── cmd
│   ├── golb
│   │   └── main.go             # read YAML, wire everything, start server
│   └── toy-backend
│       └── main.go             # toy backend to test the LB
├── internal
│   └── loadbalancer
│       ├── upstream.go
│       ├── config.go
│       ├── config_test.go
│       ├── context.go
│       ├── loadbalancer.go
│       ├── loadbalancer_test.go
│       ├── pool.go
│       └── testdata
│           └── config.test.yml
```

---

## Testing

> TODO

---

## Limitations & trade‑offs

* No **TLS termination**
* No **sticky sessions** or **weights**, strictly round‑robin.
* Minimal **timeouts/retries**, intentionally simple to keep the learning surface small.

---

## Roadmap

If I keep iterating, I’d like to add:

* Weighted round‑robin, least‑connections...
* Metrics, structured logging...
* Optional TLS on upstreams
* Admin endpoints

---

## What I learned

* Concurrency‑safe state with `sync/atomic`.
* Using `net/http` and `httputil.ReverseProxy` to implement a reverse proxy.
* Designing a small but coherent configuration surface.
* Writing tests around nondeterministic behavior (distribution vs. exact order when concurrent).