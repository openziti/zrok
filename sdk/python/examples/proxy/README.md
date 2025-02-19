
# zrok Python Proxy Example

This demonstrates using the ProxyShare class to forward requests from the public frontend to a target URL.

## Run the Example

```bash
LOG_LEVEL=INFO python ./proxy.py http://127.0.0.1:3000
```

Expected output:

```txt
2025-01-29 06:37:00,884 - __main__ - INFO - === Starting proxy server ===
2025-01-29 06:37:00,884 - __main__ - INFO - Target URL: http://127.0.0.1:3000
2025-01-29 06:37:01,252 - __main__ - INFO - Access proxy at: https://24x0pq7s6jr0.zrok.example.com:443
2025-01-29 06:37:07,981 - zrok.proxy - INFO - Share 24x0pq7s6jr0 released
```

## Basic Usage

```python
from zrok.proxy import ProxyShare
import zrok

# Load the user's zrok environment from ~/.zrok
zrok_env = zrok.environment.root.Load()

# Create a temporary proxy share (will be cleaned up on exit)
proxy = ProxyShare.create(root=zrok_env, target="http://127.0.0.1:3000")

print(f"Public URL: {proxy.endpoints}")
proxy.run()
```

## Creating a Reserved Proxy Share

To create a share token that persists and can be reused, run the example `proxy.py --unique-name my-persistent-proxy`. If the unique name already exists it will be reused. Here's how it works:

```python
proxy = ProxyShare.create(
    root=root,
    target="http://127.0.0.1:3000",
    unique_name="myuniquename"
)
```
