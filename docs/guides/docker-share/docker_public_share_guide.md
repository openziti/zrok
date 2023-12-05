---
title: Docker Public Share
sidebar_position: 10
sidebar_label: Public Share
---

With zrok and Docker, you can publicly share a web server that's running in a local container or anywhere that's reachable by the zrok container. The share can be reached through a temporary public URL that expires when the container is stopped. If you're looking for a reserved subdomain for the share, check out [zrok frontdoor](/guides/frontdoor.mdx).

Here's a short article with an overview of [public sharing with zrok](/concepts/sharing-public.md).

## Walkthrough Video

<iframe width="100%" height="315" src="https://www.youtube.com/embed/ycov--9ZtB4" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Before You Begin

To follow this guide you will need [Docker](https://docs.docker.com/get-docker/) and [the Docker Compose plugin](https://docs.docker.com/compose/install/) for running `docker compose` commands in your terminal.

## Begin Sharing with Docker Compose

A temporary public share is a great way to share a web server running in a container with someone else for a short time.

1. Make a folder on your computer to use as a Docker Compose project for your zrok public share.
1. In your terminal, change directory to the newly-created project folder.
1. Download [the temporary public share project file](pathname:///zrok-public-share/compose.yml).
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

The simplest way to share your web server is to set `ZROK_TARGET` (e.g. `https://example.com`) in the environment file.

```bash title=".env"
ZROK_TARGET="http://example.com:8080"
```

## Require Authentication

You can require authentication for your public share by setting `ZROK_OAUTH_PROVIDER` to `github` or `google` with zrok.io. You could parse the authenticated email address from the request cookie if you're building a custom server app. Read more about the OAuth features in [this blog post](https://blog.openziti.io/the-zrok-oauth-public-frontend).

```bash title=".env"
ZROK_OAUTH_PROVIDER="github"
```

## Customize Temporary Public Share

This technique is useful for adding a containerized service to the project, or mounting a filesystem directory into the container to share as a static website or file server.

Any additional services specified in the override file will be merged with `compose.yml` when you `up` the project.

You may override individual values from in `compose.yml` by specifying them in the override file.

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

## Destroy the zrok Environment

This destroys the Docker volumes containing the zrok environment secrets. The zrok environment can also be destroyed in the web console.

```bash
docker compose down --volumes
```
