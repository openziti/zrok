# zrok2 Docker Compose Self-Hosting

This directory contains the Docker Compose project for self-hosting a
complete zrok2 instance with a Ziti overlay network.

For the full self-hosting guide, see:
**<https://docs.zrok.io/docs/self-hosting/docker/>**

## Quick Start

```bash
cp .env.example .env
# Edit .env — set ZROK2_DNS_ZONE, ZROK2_ADMIN_TOKEN, and ZITI_PWD
docker compose up -d
```

## Files

| File | Purpose |
|------|---------|
| `compose.yml` | Core services (Ziti, zrok2, PostgreSQL) |
| `compose.caddy.yml` | Optional TLS overlay with Caddy |
| `Caddyfile` | Caddy reverse proxy configuration |
| `.env.example` | Documented environment variable template |

## Architecture

The stack deploys these services:

- **ziti-controller** — Ziti control plane (PKI, identities, policies)
- **ziti-router** — Ziti data plane (SDK traffic)
- **postgresql** — zrok2 database (default; SQLite3 available)
- **zrok2-init** — one-shot bootstrap (generates config, creates identities)
- **zrok2-controller** — zrok2 API and admin
- **zrok2-frontend** — public share frontend

Optional metrics pipeline (enable with `--profile metrics`):

- **rabbitmq** — metrics message queue
- **influxdb** — metrics time-series storage
