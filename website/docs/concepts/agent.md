---
sidebar_label: zrok agent overview
sidebar_position: 40
---

# zrok agent overview

The zrok agent centralizes management of your zrok shares and accesses as a single persistent background process. It
provides both a web-based console and a CLI. When the agent is running, `zrok2 share` and `zrok2 access` commands
delegate to it automatically.

## Centralized management

Without the agent running, each `zrok2 share` or `zrok2 access` command creates a separate process for each share or
access. When the agent is running:

- All shares and accesses run under a single agent process.
- The `zrok2 share` and `zrok2 access` commands delegate to the running agent automatically.
- You can stop and restart individual shares or accesses without stopping the agent.

## Restart behavior

The agent distinguishes between reserved and ephemeral shares when it restarts:

- **Reserved shares** started with `zrok2 share reserved` are automatically restarted by the agent.
- **Private accesses** started with `zrok2 access private` are automatically restarted.
- **Ephemeral shares** started with `zrok2 share public` or `zrok2 share private` are not restarted—they exist only
  for the lifetime of the agent session.

## Agent console

The agent provides a web-based console accessible with:

```bash
zrok2 agent console
```

From the console, you can:

- View the status of all active shares and accesses
- Create new shares and accesses
- Stop or restart existing shares and accesses
- Monitor traffic and connection statistics
