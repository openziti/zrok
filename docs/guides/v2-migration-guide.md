---
title: Migrating from v1 to v2
sidebar_label: v2 Migration Guide
sidebar_position: 5
---

# Migrating from zrok v1 to v2

This guide helps you transition from zrok v1.x to v2.0, focusing on the major paradigm shift from reserved shares to the new namespaces and names system.

:::warning breaking changes
zrok v2.0 introduces breaking changes. The reserved sharing commands (`zrok reserve`, `zrok release`, `zrok share reserved`) have been removed and replaced with a more flexible namespace and name system.
:::

## What's Changing?

### The Big Picture

In v1.x, you created reserved shares with persistent tokens using `zrok reserve`. In v2.0, this concept has evolved into a more powerful system:

- **namespaces** - logical groupings that contain names (like folders)
- **names** - unique identifiers within namespaces that can be reserved or ephemeral
- **name selections** - the combination of namespace and name you use when sharing

This new model provides:
- Better organization of your shares
- More flexible routing through namespace-to-frontend mappings
- Cleaner separation between identity (names) and sharing (shares)
- Support for multiple names per share

### Configuration Changes

The `defaultFrontend` configuration option has been replaced with `defaultNamespace`. You'll need to update your configuration files accordingly (see `zrok status` and `zrok config --help`).

---

## Concept Mapping: v1 → v2

| v1.x concept | v2.0 equivalent | description |
|--------------|----------------|-------------|
| reserved share | reserved name in a namespace | A persistent identifier for your share |
| share token | name within a namespace | The unique identifier users see in URLs |
| `zrok reserve` | `zrok create name` with `-r` flag | Create a reserved name |
| `zrok share reserved <token>` | `zrok share public/private -n <namespace>/<name>` | Share using a name |
| `zrok release <token>` | `zrok delete name <name>` | Remove a reserved name |

---

## Command Reference

### Removed Commands

These commands no longer exist in v2.0:

```bash
# ❌ no longer available
zrok reserve public --backend-mode web /path/to/files
zrok reserve private http://localhost:3000
zrok share reserved <token>
zrok release <token>
zrok overview public-frontends
```

### New Commands

#### Namespace Management (End Users)

```bash
# list available namespaces
zrok list namespaces

# list all your names
zrok list names

# create a new name (ephemeral by default)
zrok create name <namespace> <name>

# create a reserved name (persistent)
zrok create name <namespace> <name> -r

# modify a name (e.g., toggle reserved status)
zrok modify name <name> -r

# delete a name
zrok delete name <name>

# explicitly unshare (previously implicit)
zrok unshare <share-token>
```

#### Sharing with Names

```bash
# public share with a name selection
zrok share public <target> -n <namespace>/<name>

# private share with a name selection
zrok share private <target> -n <namespace>/<name>

# private share with vanity token
zrok share private <target> --share-token my-custom-token
```

---

## Migration Walkthrough

### Scenario 1: Simple Reserved Public Share

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
$ zrok list namespaces
available namespaces:
- public (default)

# create a reserved name in the 'public' namespace
$ zrok create name public mysite -r
created reserved name 'mysite' in namespace 'public'

# start sharing using the name selection
$ zrok share public --backend-mode web /var/www/mysite -n public/mysite
share is running at: https://mysite.share.zrok.io

# the share persists across restarts - just run the same command again
$ zrok share public --backend-mode web /var/www/mysite -n public/mysite

# when done, delete the name
$ zrok delete name mysite
```

### Scenario 2: Private Reserved Share

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
# create a reserved name
$ zrok create name public myapi -r
created reserved name 'myapi' in namespace 'public'

# share privately using the name
$ zrok share private http://localhost:8080 -n public/myapi
access your share with: zrok access private <share-token>

# alternatively, use a custom share token (vanity token)
$ zrok share private http://localhost:8080 -n public/myapi --share-token myapi-prod
access your share with: zrok access private myapi-prod

# access from another environment
$ zrok access private myapi-prod
```

### Scenario 3: Ephemeral Shares (Unchanged)

Ephemeral shares work mostly the same, but now support optional name selections:

```bash
# v1.x - still works in v2.0
$ zrok share public 8080

# v2.0 - optionally use a name
$ zrok create name public temp-demo
$ zrok share public 8080 -n public/temp-demo
# name is automatically removed when share ends (not reserved)
```

---

## Using the zrok Agent with v2

If you're using the zrok agent, there are significant improvements in v2.0:

### Automatic Retry and Error Handling

The agent now automatically retries failed shares with exponential backoff. Errored processes receive transient `err_XXXX` tokens for management.

### Persistent Shares

Shares with reserved name selections automatically restart after abnormal exit:

```bash
# create a reserved name
$ zrok create name public myapp -r

# share via agent (persists across agent restarts)
$ zrok agent share public http://localhost:3000 -n public/myapp
```

### Improved Status Command

The `zrok agent status` command now shows:
- Detailed error states for failed processes
- Frontend endpoints for public shares
- Failure information with error messages

---

## Working with Multiple Names

One powerful v2.0 feature: a single share can use multiple name selections:

```bash
# create multiple names
$ zrok create name public myapp -r
$ zrok create name public myapp-staging -r

# share using both names
$ zrok share public http://localhost:3000 \
  -n public/myapp \
  -n public/myapp-staging

# both URLs now point to the same share:
# - https://myapp.share.zrok.io
# - https://myapp-staging.share.zrok.io
```

---

## Understanding Namespaces

### What is a Namespace?

A namespace is a logical grouping for names, similar to how folders organize files. Your zrok instance may have one or more namespaces available:

- **public** - typically the default namespace for all users
- **custom namespaces** - may be created by administrators for specific purposes

### Listing Available Namespaces

```bash
$ zrok list namespaces
NAMESPACE    DESCRIPTION
public       default public namespace
```

### Namespace Grants

Administrators can control which accounts can create names in specific namespaces using namespace grants. If a namespace is "open", any user can create names in it without a grant.

---

## Checking Your Current Shares and Names

### View All Your Names

```bash
$ zrok list names
NAMESPACE    NAME         RESERVED    CREATED
public       mysite       true        2025-10-14 10:30:00
public       temp-demo    false       2025-10-14 11:45:00
```

### View Overview (Now Includes Names)

```bash
$ zrok overview
# shows human-readable format with names and namespaces

# for json output
$ zrok overview --json
```

---

## Migration Checklist

When upgrading from v1.x to v2.0:

- [ ] Identify all reserved shares you're currently using
- [ ] Note the share tokens and frontend URLs
- [ ] Check your configuration for `defaultFrontend` and change to `defaultNamespace`
- [ ] For each reserved share:
  - [ ] Create a reserved name with `zrok create name <namespace> <name> -r`
  - [ ] Stop the old `zrok share reserved <token>` process
  - [ ] Start new share with `zrok share <mode> <target> -n <namespace>/<name>`
  - [ ] Verify the share works at the new URL
  - [ ] Update any bookmarks or external references to the new URL
- [ ] Update any scripts or automation to use new commands
- [ ] If using the agent, review new error handling and status features

---

## Common Questions

### Do I Need to Keep the Same URL?

No, with the namespace/name system, your URLs will change based on the name you choose. If you need to maintain the same identifier, you can choose a name that matches your old token, though the full URL structure may differ based on how your zrok instance's frontends are configured.

### Can I Use the Old Share Tokens as Names?

Yes, names can use the same format as old share tokens. However, this is your opportunity to choose more meaningful, memorable names for your shares.

### What Happens to Permission Modes?

Permission modes (open/closed) still work the same way with `--open` and `--closed` flags, plus the `--access-grant` flag for granting access to specific accounts.

### Do Ephemeral Shares Still Work?

Yes! Ephemeral shares work just as before. The main difference is they now support optional name selections, and by default names created without the `-r` flag are ephemeral.

### What if I Have Scripts Using the Old Commands?

You'll need to update your scripts to use the new command structure. The good news is the new system is more flexible and often requires fewer steps for common workflows.

---

## Getting Help

If you run into issues during migration:

1. Check `zrok status` to verify your environment is properly enabled
2. Use `zrok list namespaces` to see what namespaces are available to you
3. Use `zrok list names` to see your current names
4. Review the error messages - v2.0 has improved error reporting
5. Consult the [self-hosting guides](/docs/category/self-hosting/) if you manage your own instance
6. Check the [concepts documentation](/concepts/index.md) for deeper understanding

---

## Next Steps

- Explore [namespace concepts](/concepts/sharing-public.mdx) (update needed for v2)
- Learn about [the zrok agent](/guides/agent/index.mdx) and its improved error handling
- Review [permission modes](/guides/permission-modes.md) which work the same in v2
- If you're an administrator, see the admin commands for namespace management
