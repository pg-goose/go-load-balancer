Load balancer in go.
Made following a _make your own x_ tutorial.

https://kasvith.me/posts/lets-create-a-simple-lb-go/

- **What's a load balancer:** Distributes load among a set of backends.
- **Load balancer strategies:**
    - Round Robin - Equal distribution between backends.
    - Weighted Round Robin - Backends are assigned weights depending on it's power.
    - Least Connections - Load distributed to servers with the least active connections.

> Round Robin is the simplest one.

- **Pasive Health Check:** Parallel to the main execution, stablish a tcp connection to check to every backend (and close it if succesful).