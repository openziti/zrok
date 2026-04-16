---
title: Upgrade from 0.4 to 1.0
---

This guide is for users migrating from 0.4 to an older 1.x release.

:::note
- For the latest zrok version, see [Install zrok](../how-tos/install/index.mdx).
- As of zrok `1.0.2`, the rebase is automatic and the configuration updates to the v1 API with no action required.
:::

## Rebase your environment

If you're running version `1.0.0` or `1.0.1`, rebase your environment to use the new versioned API:

```shell
zrok rebase apiEndpoint https://api-v1.zrok.io
```

You can now resume normal zrok API interactions.

## Troubleshoot after upgrade

If you run into issues, first verify your zrok version:

```shell
zrok version
```

Then review your configuration. If you're running version `1.0` or later, `apiEndpoint` should be `https://api-v1.zrok.io`:

```shell
zrok status
```

If you're still having issues, visit the [zrok discourse forum](https://openziti.discourse.group/c/zrok/24).

### Reset your environment

If you prefer to reset your environment from scratch:

:::warning
Running `zrok disable` deletes any running environments and shares, and releases any reserved shares.
:::

1. Disable your current environment:

    ```shell
    zrok disable
    ```

2. Reset the API endpoint to the default:

    ```shell
    zrok config unset apiEndpoint
    ```

3. Create a fresh environment:

    ```shell
    zrok enable <your account token>
    ```
