---
sidebar_label: VPN (Deprecated)
---

# The VPN Backend Mode Has Been Deprecated

The `vpn` backend mode has been removed from zrok as of `v1.1.11`.

## Why Was the VPN Mode Removed?

The VPN backend mode was removed from the core zrok distribution due to dependency management issues. The underlying libraries required for VPN functionality (specifically the TUN device management libraries) created conflicts that prevented updates to critical dependencies in the zrok codebase.

Maintaining these dependencies while keeping the rest of zrok's dependencies current proved to be increasingly difficult. After careful consideration, we decided to remove the VPN backend mode from core zrok to ensure the stability and security of the main codebase.

## Future Plans

We are exploring the possibility of re-introducing VPN functionality as a separate "layer" product built on top of zrok. This would be delivered as a separate CLI tool (such as `zrok-vpn`) that provides VPN capabilities within a zrok environment, without the dependency conflicts affecting the core zrok distribution.

This approach would allow:

- The core zrok tool to remain lean and maintainable
- VPN functionality to be developed and released on its own schedule
- Users who need VPN features to opt-in to the additional tool
- The VPN implementation could support a different subset of platforms than core zrok

## Migrating Away from VPN

If you were using the VPN backend mode, consider these alternatives:

### For Host-to-Host Connectivity

#### TCP Tunnel Mode

The `tcpTunnel` backend mode allows you to tunnel specific TCP ports between hosts. This is ideal when you need to access a specific service on a remote machine.

**Example: Sharing SSH access to a remote machine**

On the machine you want to access (the "sharing" side):

```bash
zrok share private --backend-mode tcpTunnel localhost:22
```

This creates a private share and outputs a share token (e.g., `abc123`).

On your local machine (the "accessing" side):

```bash
zrok access private --bind 127.0.0.1:2222 abc123
```

Now you can SSH to the remote machine through the tunnel:

```bash
ssh -p 2222 user@127.0.0.1
```

**Example: Accessing a database on a remote server**

Share a PostgreSQL database:

```bash
zrok share private --backend-mode tcpTunnel localhost:5432
```

Access it locally:

```bash
zrok access private --bind 127.0.0.1:5432 <share-token>
```

Connect with your database client:

```bash
psql -h 127.0.0.1 -p 5432 -U myuser mydatabase
```

#### SOCKS Proxy Mode

The `socks` backend mode creates a SOCKS5 proxy, enabling dynamic port forwarding to multiple destinations through a single share. This is useful when you need to access multiple services on a remote network.

**Example: Creating a SOCKS proxy to a remote network**

On the remote machine (the "sharing" side):

```bash
zrok share private --backend-mode socks
```

On your local machine (the "accessing" side):

```bash
zrok access private --bind 127.0.0.1:1080 <share-token>
```

Now configure your applications to use the SOCKS5 proxy at `127.0.0.1:1080`. For example:

**curl:**
```bash
curl --socks5-hostname 127.0.0.1:1080 http://internal-server:8080/api
```

**SSH (to access any host reachable from the remote machine):**
```bash
ssh -o ProxyCommand='nc -x 127.0.0.1:1080 %h %p' user@internal-host
```

**Browser:** Configure your browser's proxy settings to use SOCKS5 proxy `127.0.0.1:1080` to browse internal web applications.

#### When to Use Each Mode

| Use Case | Recommended Mode |
|----------|------------------|
| Access a single TCP service (SSH, database, etc.) | `tcpTunnel` |
| Access multiple services on a remote network | `socks` |
| Web browsing through a remote network | `socks` |
| Persistent service tunneling | `tcpTunnel` with reserved name |

### For Network-Level Access

- Consider deploying an OpenZiti network directly for full network-level zero-trust connectivity

## Questions or Feedback

If you have questions about this change or need help migrating your workflows, please start a discussion on the [OpenZiti Discourse Group](https://openziti.discourse.group/).
