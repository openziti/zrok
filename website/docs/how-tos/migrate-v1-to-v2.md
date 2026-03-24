---
sidebar_label: Migrate from v1 to v2
sidebar_position: 5
---

# Migrate from zrok v1 to v2

This guide helps you transition from zrok v1.x to v2.0, focusing on the paradigm shift from reserved shares to the new namespaces model.

:::warning breaking changes
zrok v2.0 introduces breaking changes. The reserved sharing commands (`zrok reserve`, `zrok release`, `zrok share reserved`) have been removed and replaced with a more flexible namespace system.
:::

## What's changing?

### The big picture

In v1.x, you created reserved shares with persistent share tokens using `zrok reserve`. In v2.0, this concept has evolved into a more powerful system:

- **namespaces** - "zones" that contain names (typically corresponding with a DNS zone)
- **names** - unique identifiers within namespaces that can be reserved or ephemeral (typically corresponding with an `A` record in a DNS zone)

This new model provides:
- Less coupling between environments and external names; this means you can more easily move your share backends around between hosts and reconfigure how you're sharing, without changing the names you're sharing with
- Support for multiple names per share

### Configuration changes

The `defaultFrontend` configuration option has been replaced with `defaultNamespace`. You'll need to update your configuration files accordingly (see `zrok status` and `zrok config --help`).

---

## Concept mapping: v1 → v2

| v1.x command                      | v2.0 equivalent                                                                   | description                              |
|-----------------------------------|-----------------------------------------------------------------------------------|------------------------------------------|
| `zrok reserve public <target>`    | `zrok2 create name <name>` + `zrok2 share public <target> -n <namespace>:<name>`  | Create and use a reserved public name    |
| `zrok share reserved <token>`     | `zrok2 share public <target> -n <namespace>:<name>`                               | Share using a reserved name              |
| `zrok release <token>`            | `zrok2 delete name <name>`                                                        | Remove a reserved name                   |
| `zrok reserve private <target>`   | `zrok2 share private <target> --share-token <name>`                               | Create a private share with a fixed token |

---

## Command reference

### Removed commands

These commands no longer exist in v2.0:

```bash
# no longer available
zrok2 reserve public --backend-mode web /path/to/files
zrok2 reserve private http://localhost:3000
zrok2 share reserved <token>
zrok2 release <token>
zrok2 overview public-frontends
```

### New commands

#### Namespace management (end users)

```bash
# list available namespaces
zrok2 list namespaces

# list all your names
zrok2 list names

# create a reserved name (persistent)
zrok2 create name -n <namespaceToken> <name>

# modify a name (e.g., toggle reserved status)
zrok2 modify name -n <namespaceToken> <name> -r|-r=false

# delete a name
zrok2 delete name -n <namespaceToken> <name>
```

#### Sharing with Names

```bash
# public share with a name selection
zrok2 share public <target> -n <namespaceToken>:<name>

# private share with vanity token
zrok2 share private <target> --share-token my-custom-token
```

---

## Migration walkthrough

### Scenario 1: Simple reserved public share

**v1.x workflow:**

```bash
# create a reserved share
$ zrok reserve public --backend-mode web /var/www/mysite
your reserved share token is 'abc123xyz'
reserved frontend endpoint: https://abc123xyz.share.zrok.io

# start sharing
$ zrok share reserved abc123xyz

# later, release it
$ zrok release abc123xyz
```

**v2.0 workflow:**

```bash
# first, check available namespaces
$ zrok2 list namespaces
╭───────────────────────┬─────────────────┬─────────────╮
│ NAME                  │ NAMESPACE TOKEN │ DESCRIPTION │
├───────────────────────┼─────────────────┼─────────────┤
│ example.com           │ public          │             │
╰───────────────────────┴─────────────────┴─────────────╯

# create a reserved name in the 'public' namespace
$ zrok2 create name -n public api

# start sharing using the name selection
$ zrok2 share public localhost:8080 -n public:api

# the name persists across share restarts
$ zrok2 share public localhost:8080 -n public:api

# when done, delete the name
$ zrok2 delete name -n public api
```

### Scenario 2: Private reserved share

**v1.x workflow:**

```bash
# reserve a private share
$ zrok reserve private http://localhost:8080
your reserved share token is 'xyz789abc'

# share it
$ zrok share reserved xyz789abc

# access from another environment
$ zrok access private xyz789abc
```

**v2.0 workflow:**

```bash
# share privately using the name (-s specifies a share token name)
$ zrok2 share private http://localhost:8080 -s myapi-prod

# access from another environment
$ zrok2 access private myapi-prod
```

### Scenario 3: Ephemeral shares (unchanged)

Ephemeral shares work mostly the same, but now support optional name selections:

```bash
# v1.x - still works in v2.0
$ zrok2 share public :8080
```

---

## Use the zrok Agent with v2

If you're using the zrok agent, there are significant improvements in v2.0:

### Automatic retry and error handling

The agent now automatically retries failed shares with exponential backoff. Errored processes receive transient `err_XXXX` tokens for management.

### Persistent shares

Shares with reserved name selections automatically restart after abnormal exit or agent restart:

```bash
# create a reserved name (-n defaults to 'public')
$ zrok2 create name myapp

# when agent running, share will persist across agent restarts due to reserved name
# selection
$ zrok2 share public http://localhost:3000 -n public:myapp

# when agent running, private share with --share-token will persist across agent restarts
$ zrok2 share private http://localhost:3000 --share-token myapp
```

### Improved status command

The `zrok2 agent status` command now shows:
- Detailed error states for failed processes
- Frontend endpoints for public shares
- Failure information with error messages

---

## Work with multiple names

One powerful v2.0 feature: a single share can use multiple name selections:

```bash
# create multiple names
$ zrok2 create name myapp
$ zrok2 create name myapp-staging

# share using both names
$ zrok2 share public http://localhost:3000 \
  -n public:myapp \
  -n public:myapp-staging

# both URLs now point to the same share:
 - https://myapp.share.zrok.io
 - https://myapp-staging.share.zrok.io
```

---

## Understand namespaces

### What is a namespace?

A namespace is a logical grouping for names, similar to how a DNS zone works. Your zrok instance may have one or more namespaces available:

- **public** - typically the default namespace for all users
- **custom namespaces** - may be created by administrators for specific purposes (custom domains, for example)

### List available namespaces

```bash
$ zrok2 list namespaces

╭───────────────────────┬─────────────────┬─────────────╮
│ NAME                  │ NAMESPACE TOKEN │ DESCRIPTION │
├───────────────────────┼─────────────────┼─────────────┤
│ share.zrok.io         │ public          │             │
╰───────────────────────┴─────────────────┴─────────────╯

```

---

## Check your current shares and names

### View all your names

```bash
$ zrok2 list names

╭───────────────────────────────┬─────────┬───────────┬─────────────┬──────────┬─────────────────────╮
│ URL                           │ NAME    │ NAMESPACE │ SHARE TOKEN │ RESERVED │ CREATED             │
├───────────────────────────────┼─────────┼───────────┼─────────────┼──────────┼─────────────────────┤
│ testing.share.zrok.io         │ testing │ public    │             │ true     │ 2025-10-17 13:17:11 │
╰───────────────────────────────┴─────────┴───────────┴─────────────┴──────────┴─────────────────────╯
```

### View overview (now includes names)

```bash
$ zrok2 overview
# shows human-readable format with names and namespaces

# for json output
$ zrok2 overview --json
```

---

## Common questions

### Do I need to keep the same URL?

No, with the namespace/name system, your URLs will change based on the name you choose. If you need to maintain the same identifier, you can choose a name that matches your old token, though the full URL structure may differ based on how your zrok instance's frontends are configured.

### Can I use the old share tokens as names?

Yes, names can use the same format as old share tokens. However, this is your opportunity to choose more meaningful, memorable names for your shares.

### What happens to permission modes?

Permission modes (open/closed) still work the same way with `--open` and `--closed` flags, plus the `--access-grant` flag for granting access to specific accounts.

### Do ephemeral shares still work?

Yes! Ephemeral shares work just as before. The main difference is they now support optional name selections, and by default names created without a reserved name selection are ephemeral.

### What if I have scripts using the old commands?

You'll need to update your scripts to use the new command structure. The good news is the new system is more flexible and often requires fewer steps for common workflows.

---

## Get help

If you run into issues during migration:

1. Check `zrok2 status` to verify your environment is properly enabled
2. Use `zrok2 list namespaces` to see what namespaces are available to you
3. Use `zrok2 list names` to see your current names
4. Review the error messages - v2.0 has improved error reporting
5. Consult the [self-hosting guides](@zrokdocs/category/self-hosting/) if you manage your own instance
6. Check the [concepts documentation](/concepts/index.mdx) for deeper understanding
7. Reach out on the [OpenZiti Discourse](https://openziti.discourse.group) for help

