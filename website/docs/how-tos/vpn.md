---
title: Migrate away from the VPN backend mode
sidebar_label: Migrate away from VPN
---

# Migrate away from the VPN backend mode

:::note
The `vpn` backend mode was removed in `v1.1.11` due to dependency conflicts with core zrok libraries.
:::

If you were using the VPN backend mode, consider these alternatives depending on your use case.

## For host-to-host connectivity

### TCP tunnel mode

The `tcpTunnel` backend mode tunnels specific TCP ports between hosts. Use this when you need to access a specific
service on a remote machine.

#### Example: SSH access to a remote machine

1. On the machine you want to access, create a private share of the SSH port:

    ```bash
    zrok2 share private --backend-mode tcpTunnel localhost:22
    ```

1. On your local machine, bind the share to a local port:

    ```bash
    zrok2 access private --bind 127.0.0.1:2222 <share-token>
    ```

1. Connect via SSH through the tunnel:

    ```bash
    ssh -p 2222 user@127.0.0.1
    ```

#### Example: Database on a remote server

1. On the remote machine, create a private share of the database port:

    ```bash
    zrok2 share private --backend-mode tcpTunnel localhost:5432
    ```

1. On your local machine, bind the share to a local port:

    ```bash
    zrok2 access private --bind 127.0.0.1:5432 <share-token>
    ```

1. Connect with your database client:

    ```bash
    psql -h 127.0.0.1 -p 5432 -U myuser mydatabase
    ```

### SOCKS proxy mode

The `socks` backend mode creates a SOCKS5 proxy for dynamic port forwarding to multiple destinations through a single
share. Use this when you need to access multiple services on a remote network.

1. On the remote machine, create a private share in SOCKS mode:

    ```bash
    zrok2 share private --backend-mode socks
    ```

1. On your local machine, bind the share to a local SOCKS5 port:

    ```bash
    zrok2 access private --bind 127.0.0.1:1080 <share-token>
    ```

1. Configure your applications to use the SOCKS5 proxy at `127.0.0.1:1080`. For example:

    **curl:**

    ```bash
    curl --socks5-hostname 127.0.0.1:1080 http://internal-server:8080/api
    ```

    **SSH:**

    ```bash
    ssh -o ProxyCommand='nc -x 127.0.0.1:1080 %h %p' user@internal-host
    ```

    **Browser:** Configure your browser's proxy settings to use SOCKS5 proxy `127.0.0.1:1080`.

### When to use each mode

| Use case | Recommended mode |
|---|---|
| Access a single TCP service (SSH, database, etc.) | `tcpTunnel` |
| Access multiple services on a remote network | `socks` |
| Web browsing through a remote network | `socks` |
| Persistent service tunneling | `tcpTunnel` with reserved name |

## For network-level access

Consider deploying an OpenZiti network directly for full network-level zero-trust connectivity.

## Support

If you have questions or need help migrating, start a discussion on the
[OpenZiti Discourse Group](https://openziti.discourse.group/).
