---
title: Linux Service
sidebar_position: 40
---

## Goal

Connect a reserved subdomain to a backend target with a Linux systemd service.

## Requirements

The Linux distribution must have a package manager that understands the `.deb` or `.rpm` format and be running systemd v232 or newer.

## How it Works

The `zrok-share` package creates a `zrok-share.service` unit in systemd. The administrator edits the service's configuration file to specify the:

1. zrok environment enable token
1. target URL or files to be shared and backend mode, e.g. `proxy`
1. authentication options, if wanted

When the service starts it will:

1. enable the zrok environment unless an `environment.json` file already exists
1. reserve a public subdomain for the service unless `reserved.json` exists
1. start sharing the target that was specified in the configuration file when the reservation was made

With the service running, the administrator can see the reserved subdomain for their share in the zrok console or the service log, i.e. `journalctl -lfu zrok-share.service`.

## Installation

`install.bash` scripts [the manual procedure](https://openziti.io/docs/downloads/?os=Linux).

1. Download the OpenZiti install script.

    ```bash
    curl -sSo ./openziti-install.bash https://get.openziti.io/install.bash
    ```

1. Inspect the script to ensure it is suitable to run as root on your system.

    ```bash
    less ./openziti-install.bash
    ```

1. Run the script as root to install the `zrok-share` package.

    ```bash
    sudo bash ./openziti-install.bash zrok-share
    ```

## Enable

After installing the `zrok-share` package above, save the zrok environment enable token from the zrok console in the configuration file, `/opt/openziti/etc/zrok/zrok-share.env`.

```bash
ZROK_ENABLE_TOKEN="14cbfca9772f"
```

## Use Cases

You can change the share target by modifying the configuration file and restarting the service. This allows you to change the share target without having to reserve a new subdomain.

You may switch between backend modes or change authentication options by deleting the `/var/lib/zrok-share/.zrok/reserved.json` file and restarting the service. A new subdomain will be reserved and the target will be shared using the new backend mode and authentication options.

### Proxy a Web Server

Sharing a web server means that zrok will provide a public subdomain for an existing web server. The web server could be on a private network or on the same host as zrok.

This uses zrok in the default `proxy` backend mode. Specify the target URL in the configuration file, `/opt/openziti/etc/zrok/zrok-share.env`.

```bash
ZROK_TARGET="http://127.0.0.1:3000"
ZROK_BACKEND_MODE="proxy"
```

### Serve Static Files

This uses zrok's `web` backend mode, meaning that zrok will run an embedded web server that's configured to serve your files. If there's an `index.html` file in the directory then visitors will see the rendered page in their browser, otherwise they'll see the list of available files. The directory must be readable by 'other' users, i.e. `chmod o+rX /var/www/html`.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_TARGET="/var/www/html"
ZROK_BACKEND_MODE="web"
```

Start the service, and check the zrok console or the service log for the reserved subdomain.

```bash
sudo systemctl enable --now zrok-share.service
```

### WebDAV Server

This uses zrok's `drive` backend mode to serve a directory of static files as a WebDAV resource. The directory must be readable by 'other' users, i.e. `chmod o+rX /usr/share/doc`. Add the following to the configuration file.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_TARGET="/usr/share/doc"
ZROK_BACKEND_MODE="drive"
```

### Caddy Server

Use zrok's built-in Caddy server to serve static files or as a reverse proxy to multiple web servers with various HTTP routes or as a load-balanced set, or both. A sample Caddyfile is provided. Set these in the configuration file.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_TARGET="/opt/openziti/etc/zrok/multiple_upstream.Caddyfile"
ZROK_BACKEND_MODE="caddy"
```

## Authentication

You can require a password or OAuth email address suffix.

### OAuth

You can require that visitors authenticate with an email address that matches at least one of the suffixes you specify. Add the following to the configuration file.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_OAUTH_PROVIDER="github"  # or google
ZROK_OAUTH_EMAILS="bob@example.com @acme.example.com"
```

### Password

Enable HTTP basic authentication by adding the following to the configuration file.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_BASIC_AUTH="user:passwd"
```

## Start the Service

Start the service, and check the zrok console or the service log for the reserved subdomain.

```bash
sudo systemctl enable --now zrok-share.service
```

```bash
journalctl -u zrok-share.service
```
