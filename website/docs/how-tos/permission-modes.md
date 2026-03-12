---
sidebar_position: 22
sidebar_label: Configure permission modes
---

# Configure permission modes

zrok shares support two permission modes that control who can access them:

- **Closed** (default): only the account that created the share, and any explicitly granted accounts, can access it.
- **Open**: any user of the zrok service instance can access the share if they know its share token.

## Closed permission mode (default)

All shares are created in the closed permission mode by default. No additional flags are needed:

```bash
zrok2 share private --headless -b web .
```

```buttonless title="Output"
[   0.066]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok2 access private 0vzwzodf0c7g
```

By default, any environment owned by the account that created the share can access it. A user from a different
account who tries to access the share will encounter:

```buttonless title="Output"
[ERROR]: unable to access ([POST /access][401] accessUnauthorized)
```

### Grant access to other accounts

The `zrok2 share` command includes an `--access-grant` flag to specify additional zrok accounts that are allowed to
access your share:

```bash
zrok2 share private --headless --access-grant anotheruser@test.com -b web .
```

```buttonless title="Output"
[   0.062]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok2 access private y6h4at5xvn6o
```

`anotheruser@test.com` can now access the share:

```bash
zrok2 access private --headless y6h4at5xvn6o
```

```buttonless title="Output"
[   0.049]    INFO main.(*accessPrivateCommand).run: allocated frontend 'VyvrJihAOEHD'
[   0.051]    INFO main.(*accessPrivateCommand).run: access the zrok share at the following endpoint: http://127.0.0.1:9191
```

## Open permission mode

If you want any user of the zrok service instance to be able to access your share, use the `--open` flag:

```bash
zrok2 share private --headless --open -b web .
```

```buttonless title="Output"
[   0.064]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok2 access private s4czjylwk7wa
```

## Add and remove access grants for existing shares

If you forgot to include an access grant when creating a share, or want to remove one, use `zrok2 modify share`.

1. Create a share:

    ```bash
    zrok2 share private --headless -b web .
    ```

    ```buttonless title="Output"
    [   0.064]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
    zrok2 access private s4czjylwk7wa
    ```

2. Add an access grant:

    ```bash
    zrok2 modify share s4czjylwk7wa --add-access-grant anotheruser@test.com
    ```

3. To remove the grant:

    ```bash
    zrok2 modify share s4czjylwk7wa --remove-access-grant anotheruser@test.com
    ```

## Use permission modes with reserved names

You can use permission modes with reserved names for persistent public shares. Create a reserved name, then share
with an access grant:

```bash
zrok2 create name -n public myapp
zrok2 share public localhost:8080 -n public:myapp --access-grant friend@example.com
```

For persistent private shares, use the `--share-token` flag:

```bash
zrok2 share private localhost:8080 --share-token myapi --access-grant colleague@example.com
```

To modify access grants after the fact, use the share token or custom share token:

```bash
zrok2 modify share <currentShareToken> --add-access-grant user@example.com
zrok2 modify share myapi --add-access-grant user@example.com
```

:::note
There is no way to list the access grants for a share.
:::
