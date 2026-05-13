---
sidebar_label: Migrate away from VPN
---

# Migrate away from the VPN backend mode

:::note
The `vpn` backend mode was removed in `v1.1.11` due to dependency conflicts with core zrok libraries.
:::

If you were using the VPN backend mode, consider these alternatives depending on your use case:

- [Share TCP and UDP services](./shares/share-tcp-udp.mdx): Use `tcpTunnel` to forward a specific TCP or UDP port
- [Use SOCKS proxy mode](./shares/socks-proxy-mode.mdx): Use `socks` for dynamic forwarding to multiple services

## When to use each mode

| Use case                                          | Recommended mode               |
|---------------------------------------------------|--------------------------------|
| Access a single TCP service (SSH, database, etc.) | `tcpTunnel`                    |
| Access multiple services on a remote network      | `socks`                        |
| Web browsing through a remote network             | `socks`                        |
| Persistent service tunneling                      | `tcpTunnel` with reserved name |

## For network-level access

Consider deploying an OpenZiti network directly for full network-level zero-trust connectivity.

## Support

If you have questions or need help migrating, start a discussion on the
[OpenZiti Discourse Group](https://openziti.discourse.group/).
