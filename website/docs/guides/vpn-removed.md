---
sidebar_label: VPN Backend Removed
---

# VPN Backend Mode Removed

The `vpn` backend mode has been removed from zrok as of `v1.1.11`.

## Why Was VPN Removed?

The VPN backend mode was removed from the core zrok distribution due to dependency management issues. The underlying libraries required for VPN functionality (specifically the TUN device management libraries) created conflicts that prevented updates to critical dependencies in the zrok codebase.

Maintaining these dependencies while keeping the rest of zrok's dependencies current proved to be increasingly difficult. After careful consideration, we decided to remove the VPN backend mode from core zrok to ensure the stability and security of the main codebase.

## Future Plans

We are exploring the possibility of re-introducing VPN functionality as a separate "layer" product built on top of zrok. This would be delivered as a separate CLI tool (such as `zrok-vpn`) that provides VPN capabilities within a zrok environment, without the dependency conflicts affecting the core zrok distribution.

This approach would allow:
- The core zrok tool to remain lean and maintainable
- VPN functionality to be developed and released on its own schedule
- Users who need VPN features to opt-in to the additional tool

## Migrating Away from VPN

If you were using the VPN backend mode, consider these alternatives:

### For Host-to-Host Connectivity
- Use the `tcpTunnel` backend mode to tunnel specific TCP ports between hosts
- Use the `socks` backend mode to create a SOCKS5 proxy for dynamic port forwarding

### For Network-Level Access
- Consider deploying an OpenZiti network directly for full network-level zero-trust connectivity

## Questions or Feedback

If you have questions about this change or need help migrating your workflows, please start a discussion on the [OpenZiti Discourse Group](https://openziti.discourse.group/).
