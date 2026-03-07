---
sidebar_label: "5. Set up the agent"
sidebar_position: 6
---

# Step 5: Set up the agent for always-on shares

In this step, you'll set up the zrok agent—the recommended approach for shares that need to survive terminal
sessions and system restarts.

## Why use the agent?

The foreground `zrok2 share` command in Step 4 works for quick testing, but it stops the moment you close the
terminal. The zrok agent runs in the background and manages all your shares and accesses as a single persistent
process. It also provides a web UI for monitoring and managing your shares.

When the agent is running, `zrok2 share` and `zrok2 access` commands delegate to it automatically—your shares keep
running even after you disconnect.

## Try the agent in the foreground

To try the agent before installing it as a service:

```bash
zrok2 agent
```

In another terminal, open the agent console:

```bash
zrok2 agent console
```

This opens the agent UI in your browser. You can create and monitor shares from there, or continue using the CLI.

## Install the agent as a background service

For reliable, always-on shares, install the agent as a system service:

- **Windows** — [Set up the Windows agent service](../how-tos/agent/windows-service/)
- **Linux** — [Install the Linux agent package `zrok2-agent`](../how-tos/agent/linux-service)

:::note
A native macOS agent package is not yet available. macOS users can run `zrok2 agent` in the foreground or use a
third-party process manager.
:::

Once the agent is running, use `zrok2 reserve` and `zrok2 share reserved` to create persistent shares that the agent
restarts automatically after a reboot.

For more detail on what the agent can do, see [Use the zrok Agent](../how-tos/agent/).

<div style={{marginBottom: '2rem'}} />
