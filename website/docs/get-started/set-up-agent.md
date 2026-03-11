---
sidebar_label: 5. Set up the agent
sidebar_position: 6
---

# Step 5: Set up the agent

In this step, you'll set up the zrok agent—the recommended approach for always-on shares that need to survive
terminal sessions and system restarts.

## Why use the agent?

The foreground `zrok2 share` command in Step 4 works for quick testing, but it stops the moment you close the
terminal. The zrok agent runs in the background and manages all your shares and accesses as a single persistent
process. It also provides a web UI for monitoring and managing your shares.

When the agent is running, `zrok2 share` and `zrok2 access` commands delegate to it automatically—your shares keep
running even after you disconnect.

## Try the agent in the foreground

The foreground share from step 4 won't carry over to the agent—it started as its own process and the agent begins
fresh. You can leave it running or cancel it; either way, you'll create shares through the agent separately.

To try the agent before installing it as a service:

1. Start the agent:

    ```bash
    zrok2 agent start
    ```

1. In another terminal, open the agent console:

    ```bash
    zrok2 agent console
    ```

    This opens the agent UI in your browser. You can create and monitor shares from there, or continue using the
    CLI.

## Install the agent as a background service

For reliable, always-on shares, install the agent as a system service:

- **Windows**: [Set up the Windows agent service](../how-tos/agent/windows-service/index.mdx)
- **Linux**: [Install the Linux agent package `zrok2-agent`](../how-tos/agent/linux-service.mdx)
- **Docker**: [Run the zrok agent in Docker](../how-tos/docker-agent/index.mdx)

:::note
A native macOS agent package is not yet available. macOS users can run `zrok2 agent start` in the foreground or use a
third-party process manager.
:::

For more detail on what the agent can do, see [Use the zrok Agent](../how-tos/agent/).
