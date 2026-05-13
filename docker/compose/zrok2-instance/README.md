# zrok2 Docker Compose Self-Hosting

This directory contains the Docker Compose project for self-hosting a
complete zrok2 instance with a Ziti overlay network.

For the full self-hosting guide, see deployments:
**<https://netfoundry.io/docs/zrok/category/deployment>**

## Quick start

```bash
cp .env.example .env
# Edit .env — set ZROK2_DNS_ZONE, ZROK2_ADMIN_TOKEN, and ZITI_PWD
docker compose up -d
```

Or download the essential Docker Compose files without cloning the repository:

```bash
curl -sSfL https://get.openziti.io/zrok2-instance/fetch.bash | bash
cd zrok2-instance
```

## Essential files

| File | Purpose |
| ---- | ------- |
| `compose.yml` | Core services: Ziti overlay, zrok2, PostgreSQL, RabbitMQ |
| `compose.caddy.yml` | Optional TLS overlay with Caddy |
| `.env.example` | Documented environment variable template |
| `entrypoint-init.bash` | Bootstrap script for the `zrok2-init` one-shot container |
| `fetch.bash` | Download script — fetches these files into a `zrok2-instance/` directory |

## Services

The stack deploys these services:

- **ziti-controller** — Ziti control plane (PKI, identities, policies)
- **ziti-router** — Ziti data plane (SDK traffic)
- **postgresql** — zrok2 database (default; SQLite3 available)
- **rabbitmq** — AMQP message bus for the dynamic frontend's real-time
  share mapping updates (required for named public shares via
  `zrok2 share public --name-selection`)
- **zrok2-init** — one-shot bootstrap (generates config, creates
  identities, sets up the `dynamicProxyController` gRPC service)
- **zrok2-controller** — zrok2 API and admin
- **zrok2-frontend** — AMQP-backed dynamic public share frontend

Optional metrics pipeline (enable with `--profile metrics`):

- **influxdb** — metrics time-series storage
- **zrok2-metrics-bridge** — reads Ziti usage events and publishes to
  RabbitMQ for the zrok2 controller to write to InfluxDB
