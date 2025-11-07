---
title: Upgrading From 0.4 to 1.0
---

## Upgrading an existing 0.4 environment
If you have not already, [install the latest 1.x zrok binary](/docs/guides/install) into your environment.

:::note
As of zrok `1.0.2`, the zrok rebase is automatic and the configuration will automatically be updated to the v1 API.
No action is necessary.
:::

If you are running version `1.0.0` or `1.0.1`, you can run the following to rebase your environment to use the new versioned API:
```
  zrok rebase apiEndpoint https://api-v1.zrok.io
```

Resume zrok API interactions as normal!


## Trouble after upgrade?

If you run into any issues after upgrading your environment, first verify your zrok version and review your current zrok configuration:

```
zrok version
```
Review the `apiEndpoint` configuration, if you are running version `1.0` or later, the `apiEndpoint` should be `https://api-v1.zrok.io`
```
zrok status
```

If you're still having issues, we recommend you reach out to our community support team at our [zrok discourse](https://openziti.discourse.group/c/zrok/24) forum.

If you prefer to do a hard reset of your environment, you can also run the commands below:

:::warning
Running `zrok disable` will delete any running environments or shares, and will release any reserved shares
:::
```
zrok disable
```

Reset the config back to the default API endpoint for the binary version
```
zrok config unset apiEndpoint
```
Create a fresh environment
```
zrok enable <your account token>
```
