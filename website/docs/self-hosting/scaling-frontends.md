---
sidebar_label: Scaling frontends
sidebar_position: 47
---

# Scaling zrok frontends

The [Linux](/docs/zrok/self-hosting/deployment/linux/),
[Docker](/docs/zrok/self-hosting/deployment/docker/), and
[Kubernetes](/docs/zrok/self-hosting/deployment/kubernetes/) deployment guides
describe a single-host model where one frontend process handles all public
share traffic. This page explains how to run multiple frontend instances for
higher throughput and availability.

## How the dynamic frontend works

Each `zrok2 access dynamicProxy` process:

1. Loads an OpenZiti identity from a JSON file and connects to the OpenZiti overlay
2. Subscribes to an AMQP exchange (`dynamicProxy`) using an **ephemeral queue**
   bound to its frontend token as the routing key
3. Queries the controller via gRPC (`dynamicProxyController` OpenZiti service) for
   the initial set of share mappings
4. Listens on an HTTP/HTTPS address for incoming requests
5. Routes requests by matching the `Host` header against its in-memory mapping
   table, proxying to the share's backend through the OpenZiti overlay

The AMQP queue is unique per process instance—when the controller publishes a
mapping update for a frontend token, **every instance** subscribed to that token
receives an independent copy. Instances do not compete for messages.

## Choose a scaling approach

### Option A: Multiple instances of one frontend (simplest)

Run multiple `zrok2 access dynamicProxy` processes that share the **same
frontend token and OpenZiti identity**. Place a load balancer in front of them.

```text
                    ┌─ Frontend Instance A (same token, same identity)
Load Balancer ──────┤
                    └─ Frontend Instance B (same token, same identity)
```

Each instance:

- Uses the same `frontend.yaml` (with a different `bind_address` if co-located)
- Loads the same `public.json` OpenZiti identity file (read-only—no locking)
- Receives identical AMQP mapping updates independently
- Maintains its own in-memory mapping table

This is the simplest approach. No additional zrok admin commands are needed.
The frontend token and identity file can be copied to additional hosts.

### Option B: Separate frontends per instance

Create distinct frontend records in the controller, each with its own token and
optionally its own OpenZiti identity. Map each to the same namespace(s).

Create additional frontends (each gets a unique token):

```bash
zrok2 admin create frontend --dynamic -- <public-ziti-id> frontend-2
zrok2 admin create frontend --dynamic -- <public-ziti-id> frontend-3
```

Map them to the same namespace:

```bash
zrok2 admin create namespace-frontend public <frontend-2-token>
zrok2 admin create namespace-frontend public <frontend-3-token>
```

Each frontend can share the same OpenZiti identity (`public.json`) or use separate
identities. Separate identities provide stronger isolation—if one identity is
compromised, the others are unaffected.

To use separate identities, create a new identity for each additional frontend:

```bash
zrok2 admin create identity public-2
```

Create the frontend using the new identity's OpenZiti ID:

```bash
zrok2 admin create frontend --dynamic -- <public-2-ziti-id> frontend-2
```

Then configure each frontend's `frontend.yaml` with its own `frontend_token`,
`identity`, and `controller.identity_path`.

### Compare the options

| Concern               | Option A (shared)                       | Option B (separate)                          |
| --------------------- | --------------------------------------- | -------------------------------------------- |
| Setup complexity      | Lowest—copy files                       | More admin commands                          |
| Identity isolation    | Shared                                  | Independent                                  |
| Namespace flexibility | All instances serve the same namespaces | Each frontend can serve different namespaces |
| AMQP routing          | All instances share one routing key     | Each has its own routing key                 |
| Monitoring            | Instances are indistinguishable         | Each frontend has a unique token in logs     |

For most deployments, **Option A** is sufficient. Use **Option B** when you need
per-frontend namespace isolation, distinct monitoring identifiers, or defense in
depth for the OpenZiti identity.

## Configure the load balancer

Place a Layer 4 (TCP) or Layer 7 (HTTP) load balancer in front of the frontend
instances. The load balancer must:

- Forward the `Host` header unchanged (the frontend uses it for routing)
- Support WebSocket upgrade (for `zrok2 share` connections)
- Use sticky sessions if your frontends serve stateful backends (optional)

For TLS termination, either:

- Terminate TLS at the load balancer and forward plaintext to the frontends
- Pass TLS through to the frontends (each must have the certificate)

### Example: Caddy

```text
*.share.example.com {
    reverse_proxy frontend-a:8080 frontend-b:8080
}
```

### Example: Docker Compose with Caddy

Plain Docker Compose does not load balance across replicas on a single port—you
need a reverse proxy. Remove `ports:` from the frontend service, scale it,
and let Caddy (or Nginx/Traefik) route to replicas via Docker DNS:

```yaml
services:
  zrok2-frontend:
    image: openziti/zrok2:latest
    command: ["access", "public", "/config/frontend.yaml"]
    deploy:
      replicas: 3
    volumes:
      - zrok2-config:/config:ro
    # No ports: — Caddy handles ingress

  caddy:
    image: caddy:2-alpine
    ports:
      - "0.0.0.0:443:443"
    command: caddy reverse-proxy --from :443 --to zrok2-frontend:8080
```

Docker DNS resolves `zrok2-frontend` to all replica IPs, and Caddy
round-robins across them.

### Example: Kubernetes

The [Kubernetes guide](/docs/zrok/self-hosting/deployment/kubernetes/) supports
scaling via the `frontend.replicaCount` value in the Helm chart.

## Scale other components

- **zrok2-controller**: Multiple controller instances can share the same
  PostgreSQL database. Each publishes AMQP mapping updates independently.
  Place a load balancer in front for the API endpoint.
- **zrok2-metrics-bridge**: Can read `fabric.usage` events from a file
  (single OpenZiti controller) or from an AMQP queue (multiple OpenZiti controllers).
  The AMQP source mode supports scaling across a multi-controller OpenZiti
  deployment.
