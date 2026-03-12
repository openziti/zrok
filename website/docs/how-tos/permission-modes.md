---
sidebar_position: 22
sidebar_label: Configure permission modes
---

# Configure permission modes

:::note
As of `v1.0.5`, zrok sharing defaults to the *closed* permission mode. Use the `--open` flag if you want the open
permission model instead.
:::

Permission modes control who can access your zrok shares. In the *open* permission mode, any user of the zrok service
instance can access your share if they know the share token. In the *closed* permission mode, only accounts you
explicitly grant access to can use `zrok2 access private` to reach your share.

## Create a share with closed permission mode

1. Create a private share with the `--closed` flag:

    ```bash
    zrok2 share private --headless --closed -b web .
    ```

    ```buttonless title="Output"
    [   0.066]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
    zrok2 access private 0vzwzodf0c7g
    ```

2. Share the share token with authorized users. A user from a different account who tries to access the share without
   a grant will see:

    ```buttonless title="Output"
    [ERROR]: unable to access ([POST /access][401] accessUnauthorized)
    ```

3. To grant access to another account, add the `--access-grant` flag when creating the share:

    ```bash
    zrok2 share private --headless --closed --access-grant anotheruser@test.com -b web .
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

## Add and remove access grants for existing shares

If you forgot to include an access grant when creating a share, or want to remove one, use `zrok2 modify share`.

1. Create a closed share:

    ```bash
    zrok2 share private --headless --closed -b web .
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

## Use permission modes with reserved names (v2.0)

In zrok v2.0, you can use permission modes with reserved names for persistent public shares.

1. Create a reserved name:

    ```bash
    zrok2 create name -n public myapp
    ```

2. Share with closed permission mode using the name:

    ```bash
    zrok2 share public localhost:8080 -n public:myapp --closed --access-grant friend@example.com
    ```

For persistent private shares, use the `--share-token` flag:

```bash
zrok2 share private localhost:8080 --share-token myapi --closed --access-grant colleague@example.com
```

To modify access grants after the fact, use the share token or custom share token:

```bash
zrok2 modify share <currentShareToken> --add-access-grant user@example.com
zrok2 modify share myapi --add-access-grant user@example.com
```

:::note
There is no way to list the access grants for a share.
:::
