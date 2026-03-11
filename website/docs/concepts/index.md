---
sidebar_title: Core Features
---

# Concepts

`zrok` was designed to make sharing local resources both secure and easy. In this section of the `zrok` documentation, we'll tour through all of the most important features.

Sharing with `zrok` can be either [`public`](./sharing-public.mdx) or [`private`](./sharing-private.mdx).
Naturally, regular web-based resources can be shared but `zrok` also includes support for sharing raw [TCP](./tunnels.md) and [UDP](./tunnels.md) network connections, and also includes a [website and file sharing](./files.md) feature.

Learn about `zrok` [hosting here](./hosting.md), including instructions on how to [install your own `zrok` instance](../self-hosting/deployment/linux/index.mdx).

## Instance and account

You create an account with a zrok *instance*. Your account is identified by a username and a password, which you use
to log into the web console. Your account also has a *secret token*, which you use to authenticate from the zrok
command line to interact with the instance.

You create a new account with NetFoundry's zrok instance by subscribing at [myzrok.io](https://myzrok.io), or in a
self-hosted zrok instance by running the [`zrok2 invite` command](../self-hosting/self-service-invite.mdx) or
`zrok2 admin create account`.

## Environment

Using your secret token, you use the zrok command line to create an *environment*. An environment corresponds to a
single command-line user on a specific host system.

You create a new environment with the `zrok2 enable` command.

## Shares

Once you've enabled an environment, you create one or more *shares*. Every share has a *public* or *private* sharing
mode and is identified by a *share token*. You use `zrok2 share` to create ephemeral shares. See
[public shares](./sharing-public.mdx) and [private shares](./sharing-private.mdx) for details on each mode.

## Persistent shares

By default, shares are ephemeral—when you terminate `zrok2 share`, the share and its token are gone. zrok also
supports *persistent shares* with consistent tokens that survive restarts. See
[reserved names and namespaces](./sharing-reserved.md) for the full v2.0 workflow.

## The agent

The zrok agent centralizes management of your shares and accesses as a single persistent background process. See
[zrok agent](./agent.md) for more info.
