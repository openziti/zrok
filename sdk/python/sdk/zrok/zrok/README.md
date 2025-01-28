# Zrok Python SDK

## Proxy Facility

The SDK includes a proxy facility that makes it easy to create and manage proxy shares. This is particularly useful when you need to:

1. Create an HTTP proxy with zrok
2. Optionally reserve the proxy with a unique name for persistence
3. Automatically handle cleanup of non-reserved shares

### Basic Usage

```python
from zrok.proxy import ProxyShare
import zrok

# Load the environment
root = zrok.environment.root.Load()

# Create a temporary proxy share (will be cleaned up on exit)
proxy = ProxyShare.create(root=root, target="http://my-target-service")

# Access the proxy's endpoints and token
print(f"Access proxy at: {proxy.endpoints}")
print(f"Share token: {proxy.token}")
```

### Creating a Reserved Proxy Share

To create a proxy share that persists and can be reused:

```python
# Create/retrieve a reserved proxy share with a unique name
proxy = ProxyShare.create(
    root=root,
    target="http://my-target-service",
    unique_name="my-persistent-proxy"
)
```

When a `unique_name` is provided:

1. If the zrok environment already has a share with that name, it will be reused
2. If no share exists, a new reserved share will be created
3. The share will be automatically cleaned up on exit if no `unique_name` is provided

When a `unique_name` is not provided, the randomly generated share will be cleaned up on exit.
