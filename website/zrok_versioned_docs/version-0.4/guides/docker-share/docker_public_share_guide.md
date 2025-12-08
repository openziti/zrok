---
title: Docker Compose Public Share
sidebar_position: 10
sidebar_label: Public Share
---

## Goal

Publicly share a Docker Compose service with a separate zrok environment and a permanent zrok share URL.

## Overview

With zrok, you can publicly share a service that's running in Docker. You need a zrok public share running somewhere that it can reach the service you're sharing. As long as that public share is running and your service is available, anyone with the address can use your service.

Here's a short article with an overview of [public sharing with zrok](/concepts/sharing-public.mdx).

## Walkthrough Video

<iframe width="100%" height="315" src="https://www.youtube.com/embed/ycov--9ZtB4" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## How it Works

The Docker Compose project uses your zrok account token to reserve a public subdomain and keep sharing the backend
target.

When the project runs it will:

1. enable a zrok environment unless `/mnt/.zrok/environment.json` exists in the `zrok_env` volume
1. reserve a public subdomain for the service unless `/mnt/.zrok/reserved.json` exists
1. start sharing the target specified in the `ZROK_TARGET` environment variable

## Create the Docker Project

1. Make a folder on your computer to use as a Docker Compose project for your zrok public share with a reserved subdomain and switch to the new directory in your terminal.
1. Download [the reserved public share `compose.yml` project file](pathname:///zrok-public-reserved/compose.yml) into the same directory.
1. Copy your zrok account's enable token from the zrok web console to your clipboard and paste it in a file named `.env` in the same folder like this:

    ```bash title=".env"
    ZROK_ENABLE_TOKEN="8UL9-48rN0ua"
    ```

1. Name the Share

    This unique name becomes part of the domain name of the share, e.g. `https://my-prod-app.in.zrok.io`. A random name is generated if you don't specify one.

    ```bash title=".env"
    ZROK_UNIQUE_NAME="my-prod-app"
    ```

1. Run the Compose project to start sharing the built-in demo web server. Be sure to `--detach` so the project runs in the background if you want it to auto-restart when your computer reboots.

    ```bash
    docker compose up --detach
    ```

1. Get the public share URL from the output of the `zrok-share` service or by peeking in the zrok console where the share will appear in the graph.

    ```bash
    docker compose logs zrok-share
    ```

    ```buttonless title="Output"
    zrok-public-share-1  |  https://w6r1vesearkj.in.zrok.io/
    ```

This concludes the minimum steps to begin sharing the demo web server. Read on to learn how to pivot to sharing any website or web service by leveraging additional zrok backend modes.

## Proxy Any Web Server

The simplest way to share your existing HTTP server is to set `ZROK_TARGET` (e.g. `https://example.com`) in the environment of the `docker compose up` command. When you restart the share will auto-configure for that URL.

```bash title=".env"
ZROK_TARGET="http://example.com:8080"
```

```bash
docker compose down && docker compose up
```

## Require Authentication

You can require a password or an OAuth login with certain email addresses.

### OAuth Email

You can allow specific email addresse patterns by setting `ZROK_OAUTH_PROVIDER` to `github` or `google` and
`ZROK_OAUTH_EMAILS`. Read more about the OAuth features in [this blog
post](https://blog.openziti.io/the-zrok-oauth-public-frontend).

```bash title=".env"
ZROK_OAUTH_PROVIDER="github"
ZROK_OAUTH_EMAILS="alice@example.com *@acme.example.com"
```

## Caddy is Powerful

The reserved public share project uses zrok's default backend mode, `proxy`. Another backend mode, `caddy`, accepts a path to [a Caddyfile](https://caddyserver.com/docs/caddyfile) as the value of `ZROK_TARGET` ([zrok Caddyfile examples](https://github.com/openziti/zrok/tree/main/etc/caddy)). 

Caddy is the most powerful and flexible backend mode in zrok. You must reserve a new public subdomain whenever you switch the backend mode, so using `caddy` reduces the risk that you'll have to share a new frontend URL with your users. 

With Caddy, you can balance the workload for websites or web services or share static sites and files or all of the above at the same time. You can update the Caddyfile and restart the Docker Compose project to start sharing the new configuration with the same reserved public subdomain.

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

1. Create a file `compose.override.yml`. This example adds two `httpbin` containers for load balancing, and mounts the Caddyfile into the container.

    ```yaml title="compose.override.yml"
    services:
      httpbin1:
        image: mccutchen/go-httpbin
        expose: 8080
      httpbin2:
        image: mccutchen/go-httpbin
        expose: 8080
      zrok-share:
        volumes:
          - ./Caddyfile:/mnt/.zrok/Caddyfile
    ```

1. Start a new Docker Compose project or delete the existing state volume. 

    ```bash
    docker compose down --volumes
    ```

  If you prefer to keep using the same zrok environment with the new share then delete `/mnt/.zrok/reserved.json` instead of the entire volume.

1. Run the project to load the new configuration.

    ```bash
    docker compose up --detach
    ```

1. Note the new reserved share URL from the log.

    ```bash
    docker compose logs zrok-share
    ```

    ```buttonless title="Output"
    INFO: zrok public URL: https://88s803f2qvao.in.zrok.io/
    ```
