---
title: Upgrading From 0.4 to 1.0
---

## Upgrading an existing 0.4 environment
If you have not already, [install the latest 1.x zrok binary](getting-started.mdx#installing-the-zrok-command) into your environment.

Run the following to rebase your environment to use the new versioned API:
```
  zrok rebase apiEndpoint https://api-v1.zrok.io
```

Resume zrok API interactions as normal!


## Trouble after upgrade?

If you run into any issues after upgrading your environment, try running the following:

```
zrok disable
zrok config set apiEndpoint https://api-v1.zrok.io
zrok enable <your account token>

```
