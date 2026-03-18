---
sidebar_position: 22
---

# Share an HTTP service

The `proxy` backend mode forwards incoming requests to an HTTP or HTTPS server running on your machine. It's the default
backend mode, so you don't need to specify `--backend-mode` unless you're switching to a different mode.

1. Run `zrok2 share public` with your port or URL:

    ```bash
    zrok2 share public 8080
    ```

    zrok assigns a public URL and opens a full-screen terminal display showing the URL, share type, and a live feed of
    incoming requests:

    ![zrok2 share terminal output](../../images/zrok2-serve.png)

    The share is active as long as the command is running. Press `Ctrl+C` or `q` to exit and tear down the share.

2. Optionally, to disable the terminal UI and send output to `stdout`, pass the `--headless` flag:

    ```bash
    zrok2 share public --headless 8080
    ```

To create public shares with the agent, see [Manage shares with the agent](../../how-tos/agent/manage-shares.mdx).
