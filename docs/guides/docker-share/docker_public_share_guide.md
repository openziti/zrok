---
sidebar_position: 10
sidebar_label: Public Share
---

# Docker Public Share

With zrok and Docker, you can publicly share a web server that's running in a local container or anywhere that's reachable by the zrok container. The share can be reached through a public URL thats temporary or reserved (reusable).

## Walkthrough Video

<iframe width="100%" height="315" src="https://www.youtube.com/embed/ycov--9ZtB4" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Before You Begin

To follow this guide you will need [Docker](https://docs.docker.com/get-docker/) and [the Docker Compose plugin](https://docs.docker.com/compose/install/) for running `docker compose` commands in your terminal.

## Temporary or Reserved Public Share

A temporary public share is a great way to share a web server running in a container with someone else for a short time. A reserved public share is a great way to share a reliable web server running in a container with someone else for a long time.

1. Make a folder on your computer to use as a Docker Compose project for your zrok public share.
1. In your terminal, change directory to the newly-created project folder.
1. Download either [the temporary public share project file](pathname:///zrok-public-share/compose.yml) or [the reserved public share project file](pathname:///zrok-public-reserved/compose.yml) into the project folder.
1. Copy your zrok environment token from the zrok web console to your clipboard and paste it in a file named `.env` in the same folder like this:

  ```bash title=".env"
  ZROK_ENABLE_TOKEN="8UL9-48rN0ua"
  ```

1. Set the zrok API endpoint if self-hosting zrok. Skip this if using zrok.io.

  ```bash title=".env"
  ZROK_API_ENDPOINT="https://zrok.example.com"
  ```

1. Run the Compose project to start sharing the built-in demo web server.

  ```bash
  docker compose up --detach
  ```

1. Get the public share URL from the output of the `zrok-share` service or by peeking in the zrok console where the share will be graphed.

  ```bash
  docker compose logs zrok-share
  ```

  ```buttonless title="Output"
  zrok-public-share-1  |  https://w6r1vesearkj.in.zrok.io/
  ```

This concludes sharing the demo web server. Read on to learn how to pivot to sharing any web server leveraging additional zrok backend modes.

## Proxy Any Web Server

The simplest way to share your web server is to set `ZROK_TARGET` (e.g. `https://example.com`) in the environment of the `docker compose up` command. When you restart the share will auto-configure for that upstream server URL. This applies to both temporary and reserved public shares.

```bash title=".env"
ZROK_TARGET="http://example.com:8080"
```

## Require Authentication

You can require authentication for your public share by setting `ZROK_OAUTH_PROVIDER` to `github` or `google` if you're using our hosted zrok.io, and any OIDC provider you've configured if self-hosting. You can parse the authenticated email address from the request cookie. Read more about the OAuth features in [this blog post](https://blog.openziti.io/the-zrok-oauth-public-frontend). This applies to both temporary and reserved public shares.

```bash title=".env"
ZROK_OAUTH_PROVIDER="github"
```

## Customize Temporary Public Share

1. Create a file `compose.override.yml`. This example demonstrates sharing a static HTML directory `/tmp/html` from the Docker host's filesystem.

  ```yaml title="compose.override.yml"
  services:
    zrok-share:
      command: share public --headless --backend-mode web /tmp/html
      volumes:
        - /tmp/html:/tmp/html
  ```

1. Re-run the project to load the new configuration.

  ```bash
  docker compose up --force-recreate --detach
  ```

1. Get the new tempoary public share URL for the `zrok-share` container.

  ```bash
  docker compose logs zrok-share
  ```

  ```buttonless title="Output"
  zrok-public-share-1  |  https://w6r1vesearkj.in.zrok.io/
    ```

## Customize Reserved Public Share

The reserved public share project uses zrok's `caddy` mode. Caddy accepts configuration as a Caddyfile that is mounted into the container ([zrok Caddyfile examples](https://github.com/openziti/zrok/tree/main/etc/caddy)).

1. Create a Caddyfile. This example demonstrates proxying two HTTP servers with a weighted round-robin load balancer.

  ```console title="Caddyfile"
  http:// {
    # zrok requires this bind address template
    bind {{ .ZrokBindAddress }}
    reverse_proxy /* {
      to http://httpbin1:8080 http://httpbin2:8080
      lb_policy weighted_round_robin 3 2
    }
  }
  ```

1. Create a file `compose.override.yml`. This example adds two `httpbin` containers for Caddy load balance, and masks the default Caddyfile with our custom one.

  ```yaml title="compose.override.yml"
  services:
    httpbin1:
      image: mccutchen/go-httpbin  # 8080/tcp
    httpbin2:
      image: mccutchen/go-httpbin  # 8080/tcp
    zrok-share:
      volumes:
        - ./Caddyfile:/mnt/.zrok/Caddyfile
  ```

1. Re-run the project to load the new configuration.

  ```bash
  docker compose up --force-recreate --detach
  ```

1. Recall the reserved share URL from the log.

  ```bash
  docker compose logs zrok-share
  ```

  ```buttonless title="Output"
  INFO: zrok public URL: https://88s803f2qvao.in.zrok.io/
  ```

## Destroy the zrok Environment

This destroys the Docker volumes containing the zrok environment secrets. The zrok environment can also be destroyed in the web console.

```bash
docker compose down --volumes
```
