---
sidebar_label: Overview
sidebar_position: 1
---

# zrok overview

zrok (*/ziːɹɒk/ ZEE-rock*) is a secure, open-source platform for sharing local services and files over the internet
without opening firewall ports or managing TLS. Built on [OpenZiti](https://openziti.io/) zero-trust networking and
backed by [NetFoundry](https://netfoundry.io), it supports public HTTPS shares with optional authentication and private
shares accessible only to other zrok users. Run it as a hosted service at [myzrok.io](https://myzrok.io) with a generous
free tier, or [self-host your own instance](./self-hosting/deployment/linux/index.mdx) on Linux, Docker, or Kubernetes.

## Why use zrok

Use zrok to share a running service, like a web server or network socket, or a directory of static files.

When sharing publicly, you can reserve a public hostname, enable authentication options, or both. Public shares proxy
HTTPS to your service or files.

When sharing privately, only users with the share token (and the appropriate permission grants) can access your share.
In addition to what you can share publicly, private shares can include TCP and UDP services.

Ready to share something? [Get started.](./get-started/index.mdx)
