---
title: Upgrading From 0.4 to 1.0
---

## Upgrading an existing 0.4 environment
If you have not already, [install the latest 1.x zrok binary](/docs/guides/install) into your environment.

Run the following to rebase your environment to use the new versioned API:
```
  zrok rebase apiEndpoint https://api-v1.zrok.io
```

:::note
As of zrok `1.0.2`, the zrok rebase is automatic and the configuration will automatically be updated to the v1 API.
:::


Resume zrok API interactions as normal!


## Trouble after upgrade?

If you run into any issues after upgrading your environment, try running the following:

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
