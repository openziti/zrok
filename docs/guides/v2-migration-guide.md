---
title: Migrating from v1 to v2
sidebar_label: v2 Migration Guide
sidebar_position: 5
---

# migrating from zrok v1 to v2

this guide helps you transition from zrok v1.x to v2.0, focusing on the major paradigm shift from reserved shares to the new namespaces and names system.

:::warning breaking changes
zrok v2.0 introduces breaking changes. the reserved sharing commands (`zrok reserve`, `zrok release`, `zrok share reserved`) have been removed and replaced with a more flexible namespace and name system.
:::

## what's changing

### the big picture

in v1.x, you created reserved shares with persistent tokens using `zrok reserve`. in v2.0, this concept has evolved into a more powerful system:

- **namespaces** - logical groupings that contain names (like folders)
- **names** - unique identifiers within namespaces that can be reserved or ephemeral
- **name selections** - the combination of namespace and name you use when sharing

this new model provides:
- better organization of your shares
- more flexible routing through namespace-to-frontend mappings
- cleaner separation between identity (names) and sharing (shares)
- support for multiple names per share

### configuration changes

the `defaultFrontend` configuration option has been replaced with `defaultNamespace`. you'll need to update your configuration files accordingly (see `zrok status` and `zrok config --help`).

---

## concept mapping: v1 → v2

| v1.x concept | v2.0 equivalent | description |
|--------------|----------------|-------------|
| reserved share | reserved name in a namespace | a persistent identifier for your share |
| share token | name within a namespace | the unique identifier users see in URLs |
| `zrok reserve` | `zrok create name` with `-r` flag | create a reserved name |
| `zrok share reserved <token>` | `zrok share public/private -n <namespace>/<name>` | share using a name |
| `zrok release <token>` | `zrok delete name <name>` | remove a reserved name |

---

## command reference

### removed commands

these commands no longer exist in v2.0:

```bash
# ❌ no longer available
zrok reserve public --backend-mode web /path/to/files
zrok reserve private http://localhost:3000
zrok share reserved <token>
zrok release <token>
zrok overview public-frontends
```

### new commands

#### namespace management (end users)

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

#### sharing with names

```bash
# public share with a name selection
zrok share public <target> -n <namespace>/<name>

# private share with a name selection
zrok share private <target> -n <namespace>/<name>

# private share with vanity token
zrok share private <target> --share-token my-custom-token
```

---

## migration walkthrough

### scenario 1: simple reserved public share

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

### scenario 2: private reserved share

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

### scenario 3: ephemeral shares (unchanged)

ephemeral shares work mostly the same, but now support optional name selections:

```bash
# v1.x - still works in v2.0
$ zrok share public 8080

# v2.0 - optionally use a name
$ zrok create name public temp-demo
$ zrok share public 8080 -n public/temp-demo
# name is automatically removed when share ends (not reserved)
```

---

## using the zrok agent with v2

if you're using the zrok agent, there are significant improvements in v2.0:

### automatic retry and error handling

the agent now automatically retries failed shares with exponential backoff. errored processes receive transient `err_XXXX` tokens for management.

### persistent shares

shares with reserved name selections automatically restart after abnormal exit:

```bash
# create a reserved name
$ zrok create name public myapp -r

# share via agent (persists across agent restarts)
$ zrok agent share public http://localhost:3000 -n public/myapp
```

### improved status command

the `zrok agent status` command now shows:
- detailed error states for failed processes
- frontend endpoints for public shares
- failure information with error messages

---

## working with multiple names

one powerful v2.0 feature: a single share can use multiple name selections:

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

## understanding namespaces

### what is a namespace?

a namespace is a logical grouping for names, similar to how folders organize files. your zrok instance may have one or more namespaces available:

- **public** - typically the default namespace for all users
- **custom namespaces** - may be created by administrators for specific purposes

### listing available namespaces

```bash
$ zrok list namespaces
NAMESPACE    DESCRIPTION
public       default public namespace
```

### namespace grants

administrators can control which accounts can create names in specific namespaces using namespace grants. if a namespace is "open", any user can create names in it without a grant.

---

## checking your current shares and names

### view all your names

```bash
$ zrok list names
NAMESPACE    NAME         RESERVED    CREATED
public       mysite       true        2025-10-14 10:30:00
public       temp-demo    false       2025-10-14 11:45:00
```

### view overview (now includes names)

```bash
$ zrok overview
# shows human-readable format with names and namespaces

# for json output
$ zrok overview --json
```

---

## migration checklist

when upgrading from v1.x to v2.0:

- [ ] identify all reserved shares you're currently using
- [ ] note the share tokens and frontend URLs
- [ ] check your configuration for `defaultFrontend` and change to `defaultNamespace`
- [ ] for each reserved share:
  - [ ] create a reserved name with `zrok create name <namespace> <name> -r`
  - [ ] stop the old `zrok share reserved <token>` process
  - [ ] start new share with `zrok share <mode> <target> -n <namespace>/<name>`
  - [ ] verify the share works at the new URL
  - [ ] update any bookmarks or external references to the new URL
- [ ] update any scripts or automation to use new commands
- [ ] if using the agent, review new error handling and status features

---

## common questions

### do i need to keep the same URL?

no, with the namespace/name system, your URLs will change based on the name you choose. if you need to maintain the same identifier, you can choose a name that matches your old token, though the full URL structure may differ based on how your zrok instance's frontends are configured.

### can i use the old share tokens as names?

yes, names can use the same format as old share tokens. however, this is your opportunity to choose more meaningful, memorable names for your shares.

### what happens to permission modes?

permission modes (open/closed) still work the same way with `--open` and `--closed` flags, plus the `--access-grant` flag for granting access to specific accounts.

### do ephemeral shares still work?

yes! ephemeral shares work just as before. the main difference is they now support optional name selections, and by default names created without the `-r` flag are ephemeral.

### what if i have scripts using the old commands?

you'll need to update your scripts to use the new command structure. the good news is the new system is more flexible and often requires fewer steps for common workflows.

---

## getting help

if you run into issues during migration:

1. check `zrok status` to verify your environment is properly enabled
2. use `zrok list namespaces` to see what namespaces are available to you
3. use `zrok list names` to see your current names
4. review the error messages - v2.0 has improved error reporting
5. consult the [self-hosting guides](/docs/guides/self-hosting/) if you manage your own instance
6. check the [concepts documentation](/docs/concepts/) for deeper understanding

---

## next steps

- explore [namespace concepts](/concepts/sharing-public.mdx) (update needed for v2)
- learn about [the zrok agent](/guides/agent/) and its improved error handling
- review [permission modes](/guides/permission-modes.md) which work the same in v2
- if you're an administrator, see the admin commands for namespace management
