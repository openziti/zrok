---
title: Linux Service
sidebar_position: 40
---

## Goal

Proxy a reserved public subdomain to a backend target with a Linux service.

## Requirements

The Linux distribution must have a package manager that understands the `.deb` or `.rpm` format and be running systemd v232 or newer. The service was tested with Ubuntu 20-22, Debian 11-12, Rocky 8-9, and Fedora 37-38.

## How it Works

The `zrok-share` package creates a `zrok-share.service` unit in systemd. The administrator edits the service's configuration file to specify the:

1. zrok environment enable token
1. target URL or files to be shared and backend mode, e.g. `proxy`
1. authentication options, if wanted

When the service starts it will:

1. enable the zrok environment unless `/var/lib/zrok-share/.zrok/environment.json` exists
1. reserve a public subdomain for the service unless `/var/lib/zrok-share/.zrok/reserved.json` exists
1. start sharing the target specified in the configuration file

## Installation

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

Save the enable token from the zrok console in the configuration file.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_ENABLE_TOKEN="14cbfca9772f"
```

## Use Cases

You may change the target for the current backend mode, e.g. `proxy`, by editing the configuration file and restarting the service. The reserved subdomain will remain the same.

You may switch between backend modes or change authentication options by deleting `/var/lib/zrok-share/.zrok/reserved.json` and restarting the service. A new subdomain will be reserved.

### Proxy a Web Server

Proxy a reserved subdomain to an existing web server. The web server could be on a private network or on the same host as zrok.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_TARGET="http://127.0.0.1:3000"
ZROK_BACKEND_MODE="proxy"
```

### Serve Static Files

Run zrok's embedded web server to serve the files in a directory. If there's an `index.html` file in the directory then visitors will see that web page in their browser, otherwise they'll see a generated index of the files. The directory must be readable by 'other', e.g. `chmod -R o+rX /var/www/html`.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_TARGET="/var/www/html"
ZROK_BACKEND_MODE="web"
```

### WebDAV Server

This uses zrok's `drive` backend mode to serve a directory of static files as a WebDAV resource. The directory must be readable by 'other', e.g. `chmod -R o+rX /usr/share/doc`.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_TARGET="/usr/share/doc"
ZROK_BACKEND_MODE="drive"
```

### Caddy Server

Use zrok's built-in Caddy server to serve static files or as a reverse proxy to multiple web servers with various HTTP routes or as a load-balanced set. A sample Caddyfile is available in the path shown.

```bash title="/opt/openziti/etc/zrok/zrok-share.env"
ZROK_TARGET="/opt/openziti/etc/zrok/multiple_upstream.Caddyfile"
ZROK_BACKEND_MODE="caddy"
```

## Authentication

You can limit access to certain email addresses with OAuth or require a password.

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

```bash title="run now and at startup"
sudo systemctl enable --now zrok-share.service
```

```bash title="run now"
sudo systemctl restart zrok-share.service
```

```bash
journalctl -u zrok-share.service
```

## Package Contents

The files included in the `zrok-share` package are sourced [here in GitHub](https://github.com/openziti/zrok/tree/main/nfpm).
