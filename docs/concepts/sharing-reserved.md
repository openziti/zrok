---
sidebar_position: 10
---

# Reserved Names and Namespaces

:::info v2.0 feature
This guide describes the v2.0 namespace and name system. If you're migrating from v1.x, see the [v2 migration guide](/docs/guides/v2-migration-guide.md) for details on how this replaces the old `zrok reserve` workflow.
:::

By default, when you create a `public` or `private` share using the `zrok share` command, zrok assigns it a randomly generated _share token_. When you terminate the `zrok share` command, the share is deleted and the token is no longer valid. If you run `zrok share` again, you will receive a brand new share token.

In v2.0, zrok introduces a more powerful system for creating persistent shares through **namespaces** and **names**.

## Understanding Namespaces and Names

### Namespaces

A **namespace** is a logical grouping for names, similar to how a DNS zone works. Think of it as a container that holds related names. For example:

- A `public` namespace might correspond to `share.zrok.io`
- A custom namespace might correspond to your own domain like `example.com`

Namespaces can be:
- **Open** - accessible to all users of the zrok service instance
- **Closed** - requiring explicit grants for access

You can see available namespaces with:

```bash
zrok list namespaces
```

### Names

A **name** is a unique identifier within a namespace. Names can be:

- **Reserved** - persistent across multiple runs of `zrok share`, similar to v1.x reserved shares
- **Ephemeral** - temporary, deleted when the share terminates

Think of names as similar to DNS A records within a zone. For example, if you create a name `api` in the `public` namespace (corresponding to `share.zrok.io`), your share might be accessible at `https://api.share.zrok.io`.

## Creating Reserved Names

To create a reserved name, use the `zrok create name` command:

```bash
# create a reserved name in the default namespace
zrok create name myapp

# create a reserved name in a specific namespace
zrok create name -n public myapp

# create a reserved name in a custom namespace
zrok create name -n <namespaceToken> api
```

Once created, you can use this name repeatedly across share sessions. The name persists even when your share is not running.

## Using Names with Shares

### Public Shares with Names

Use the `-n` flag to specify a name selection when creating a public share:

```bash
# share using a name in the default namespace
zrok share public localhost:8080 -n public:myapp

# share using a name in a specific namespace
zrok share public localhost:8080 -n <namespaceToken>:api
```

The name can be either reserved (created with `zrok create name`) or ephemeral (created on-the-fly).

### Private Shares with Custom Tokens

For private shares, you can use the `--share-token` flag to specify a persistent vanity token:

```bash
# create a private share with a custom token
zrok share private localhost:8080 --share-token myapi-prod

# access it from another environment
zrok access private myapi-prod
```

When using the zrok agent, shares with `--share-token` are automatically persistent and will restart after abnormal exit or agent restart.

### Multiple Names on One Share

A powerful v2.0 feature: you can specify multiple names for a single share:

```bash
# create multiple names
zrok create name -n public myapp
zrok create name -n public myapp-staging

# share using both names
zrok share public localhost:3000 \
  -n public:myapp \
  -n public:myapp-staging
```

Both URLs will point to the same backend target, allowing you to use different names for the same service.

## Managing Names

### Listing Your Names

See all your names across all namespaces:

```bash
zrok list names
```

This shows a table with:
- URL (the full public URL if applicable)
- Name
- Namespace
- Share token (if currently being shared)
- Reserved status
- Creation timestamp

### Modifying Name Status

Toggle the reserved status of a name:

```bash
# make a name reserved (persistent)
zrok modify name -n public myapp -r

# make a name ephemeral (will be deleted when share ends)
zrok modify name -n public myapp -r=false
```

### Deleting Names

Remove a reserved name when you no longer need it:

```bash
# delete a name from the default namespace
zrok delete name myapp

# delete a name from a specific namespace
zrok delete name -n <namespaceToken> api
```

## Configuring Default Namespace

You can set a default namespace to avoid specifying `-n` on every command:

```bash
# set via config command
zrok config set defaultNamespace public

# or via environment variable
export ZROK_DEFAULT_NAMESPACE=public
```

Once configured, commands will use this namespace by default:

```bash
# these are equivalent if defaultNamespace is set to 'public'
zrok create name myapp
zrok create name -n public myapp
```

## Migration from v1.x

If you're coming from zrok v1.x, here's the mapping:

| v1.x Command | v2.0 Equivalent |
|--------------|-----------------|
| `zrok reserve public <target>` | `zrok create name <name>` + `zrok share public <target> -n <namespace>:<name>` |
| `zrok share reserved <token>` | `zrok share public <target> -n <namespace>:<name>` |
| `zrok release <token>` | `zrok delete name <name>` |
| `zrok reserve private <target>` | `zrok share private <target> --share-token <name>` |

See the [v2 migration guide](/docs/guides/v2-migration-guide.md) for comprehensive migration instructions.

## Benefits of the Namespace/Name System

The v2.0 namespace and name system provides several advantages over v1.x reserved shares:

1. **Flexibility** - less coupling between environments and external names
2. **Portability** - easily move share backends between hosts without changing public names
3. **Multiple Names** - use multiple names for the same share
4. **Organization** - logical grouping through namespaces
5. **Custom Domains** - support for custom domain namespaces (when configured by administrators)
