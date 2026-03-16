---
sidebar_position: 10
---

# Reserved names and namespaces

:::info v2.0 feature
This page describes the v2.0 namespace and name system. If you're migrating from v1.x, see the [v2 migration
guide](/how-tos/migrate-v1-to-v2.md) for details on how this replaces the old `zrok reserve` workflow.
:::

By default, when you create a public or private share using `zrok2 share`, zrok assigns it a randomly generated _share
token_. When you terminate the `zrok2 share` command, the share is deleted and the token is no longer valid. If you run
`zrok2 share` again, you'll receive a brand new share token.

In v2.0, zrok introduces a more powerful system for creating persistent shares through **namespaces** and **names**.

To create and manage reserved names, see [Manage reserved names](../how-tos/manage-reserved-names.md).

## Understand namespaces and names

### Namespaces

A **namespace** is a logical grouping for names, similar to how a DNS zone works. Think of it as a container that holds
related names. For example:

- A `public` namespace might correspond to `share.zrok.io`
- A custom namespace might correspond to your own domain like `example.com`

Namespaces can be:

- **Open**: accessible to all users of the zrok service instance
- **Closed**: requiring explicit grants for access

You can see available namespaces with:

```bash
zrok2 list namespaces
```

### Names

A **name** is a unique identifier within a namespace. Names can be:

- **Reserved**: persistent across multiple runs of `zrok2 share`, similar to v1.x reserved shares
- **Ephemeral**: temporary, deleted when the share terminates

Think of names as similar to DNS A records within a zone. For example, if you create a name `api` in the `public`
namespace (corresponding to `share.zrok.io`), your share is accessible at `https://api.share.zrok.io`.

## Migration from v1.x

If you're coming from zrok v1.x, here's the mapping:

| v1.x command                     | v2.0 equivalent                                                                  |
|----------------------------------|----------------------------------------------------------------------------------|
| `zrok2 reserve public <target>`  | `zrok2 create name <name>` + `zrok2 share public <target> -n <namespace>:<name>` |
| `zrok2 share reserved <token>`   | `zrok2 share public <target> -n <namespace>:<name>`                              |
| `zrok2 release <token>`          | `zrok2 delete name <name>`                                                       |
| `zrok2 reserve private <target>` | `zrok2 share private <target> --share-token <name>`                              |

See the [v2 migration guide](/how-tos/migrate-v1-to-v2.md) for comprehensive migration instructions.

## Benefits of the namespace/name system

The v2.0 namespace and name system provides several advantages over v1.x reserved shares:

- **Flexibility**: less coupling between environments and external names
- **Portability**: easily move share backends between hosts without changing public names
- **Multiple names**: use multiple names for the same share
- **Organization**: logical grouping through namespaces
- **Custom domains**: support for custom domain namespaces (when configured by administrators)
