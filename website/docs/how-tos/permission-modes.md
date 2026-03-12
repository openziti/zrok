---
sidebar_position: 22
sidebar_label: Configure permission modes
---

# Configure permission modes

zrok shares support two _permission modes_ that control who can access them:

- **Closed** (default) -- only the account that created the share (and any explicitly granted accounts) can access it. This is the most secure option and the default for all shares.
- **Open** -- any user of the zrok service instance can access the share if they know its share token.

## Closed permission mode (default)

All shares are created with the _closed permission mode_ by default. No additional flags are needed:

```
$ zrok2 share private --headless -b web .
[   0.066]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok2 access private 0vzwzodf0c7g
```

By default any environment owned by the account that created the share is _allowed_ to access the new share. But a user trying to access the share from an environment owned by a different account will encounter the following error message:

```
$ zrok2 access private 0vzwzodf0c7g
[ERROR]: unable to access ([POST /access][401] accessUnauthorized)
```

### Grant access to other accounts

The `zrok2 share` command includes an `--access-grant` flag, which allows you to specify additional zrok accounts that are allowed to access your shares:

```
$ zrok2 share private --headless --access-grant anotheruser@test.com -b web .
[   0.062]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok2 access private y6h4at5xvn6o
```

And now `anotheruser@test.com` will be allowed to access the share:

```
$ zrok2 access private --headless y6h4at5xvn6o
[   0.049]    INFO main.(*accessPrivateCommand).run: allocated frontend 'VyvrJihAOEHD'
[   0.051]    INFO main.(*accessPrivateCommand).run: access the zrok share at the following endpoint: http://127.0.0.1:9191
```

## Open permission mode

If you want any user of the zrok service instance to be able to access your share, use the `--open` flag:

```
$ zrok2 share private --headless --open -b web .
[   0.064]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok2 access private s4czjylwk7wa
```

## Add and remove access grants for existing shares

If you've created a share and you forgot to include an access grant, or want to remove an access grant that was mistakenly added, you can use the `zrok2 modify share` command to make the adjustments:

Create a share:

```
$ zrok2 share private --headless -b web .
[   0.064]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok2 access private s4czjylwk7wa
```

In another shell in the same environment you can execute:

```
$ zrok2 modify share s4czjylwk7wa --add-access-grant anotheruser@test.com
updated
```

And to remove the grant:

```
$ zrok2 modify share s4czjylwk7wa --remove-access-grant anotheruser@test.com
updated
```

## Use permission modes with reserved names

You can use permission modes with reserved names for persistent public shares:

```bash
# create a reserved name
$ zrok2 create name -n public myapp

# share with closed permission mode (the default) and grant access
$ zrok2 share public localhost:8080 -n public:myapp --access-grant friend@example.com
```

For persistent private shares, use the `--share-token` flag:

```bash
# create a persistent private share with custom token and grant access
$ zrok2 share private localhost:8080 --share-token myapi --access-grant colleague@example.com
```

You can modify access grants for shares using reserved names or custom share tokens:

```bash
# modify a share using a reserved name's current share token
$ zrok2 modify share <currentShareToken> --add-access-grant user@example.com

# or modify using the custom share token
$ zrok2 modify share myapi --add-access-grant user@example.com
```
