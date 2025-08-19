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

### Setup

Follow these steps to get a local demo running.

1) **Get the binaries**

_Build from source (recommended)_

```bash
# from the repo root
go build -o bin/golb ./cmd/golb
go build -o bin/toy-backend ./cmd/toy-backend
```

_…or use a prebuilt binary from releases._

* Download `golb` and (optionally) `toy-backend` for your OS/arch.

  ```bash
  chmod +x golb toy-backend
  ```

2) **Start a few toy backends**

Open **three terminals** and run:

```bash
./toy-backend -id 1
./toy-backend -id 2
./toy-backend -id 3
```

Each instance prints it's server ID and listens to port `:808[id]`

#### 3) Create the config

Create a `config.yaml` in your working directory:

```yaml
# config.yaml
port: 8080
healthCheckTimeout: 3
healthCheckPeriod: 5
backends:
  - http://127.0.0.1:8081/health
  - http://127.0.0.1:8082/health
  - http://127.0.0.1:8083/health
```

4) **Run golb**

```bash
# with an explicit config file
./golb -confpath ./config.yaml 

# or rely on the default name (config.yml) in the current directory
./golb
```

5) **Send a few requests**

In another terminal:

```bash
for i in {1..9}; do curl -s http://127.0.0.1:8080 | head -n1; done
```

You should see the responses alternate across backends.

6) **See health checks in action (optional)**

* Stop one `toy-backend` process.
* Keep curling `http://127.0.0.1:8080` — you’ll notice `golb` skips the unhealthy backend after its next check.
* Restart it and watch it rejoin the rotation.

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