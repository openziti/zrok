---
sidebar_position: 20
---

# Dynamic Proxy Frontend Migration Guide

This guide helps self-hosting system administrators migrate from the legacy `zrok access public` (`publicProxy`) to the new v2.0 `zrok access dynamicProxy` (`dynamicProxy`) architecture. The dynamicProxy system provides enhanced scalability, namespace-based naming mapping, and full support for OAuth integration.

## Overview

### What's Changing

In zrok v2.0, the public sharing model has been redesigned around **namespaces** and **names**:

- **v1.x model**: shares created with `zrok share public` or `zrok reserve public` connected directly to a public frontend
- **v2.0 model**: shares use reserved or ephemeral **names** within **namespaces**, which are mapped to dynamic frontends

The new `dynamicProxy` frontend replaces `publicProxy` with several key improvements:

- **Namespace-based organization**: logical grouping of names similar to DNS zones
- **Distributed mapping updates**: AMQP-based real-time synchronization across multiple frontend instances
- **Enhanced scalability**: multiple dynamicProxy instances can serve the same namespace
- **gRPC communication**: efficient controller-to-frontend communication for mapping queries

### Architecture Components

The dynamicProxy system consists of three main components:

1. **dynamicProxyController** (in zrok controller)
   - gRPC service for mapping queries from frontends
   - AMQP publisher for broadcasting mapping updates
   - Manages the distribution of namespace/name to share mappings across your entire fleet of `dynamicProxy` instances

2. **AMQP Message Broker** (RabbitMQ)
   - distributes mapping updates from controller to all frontend instances
   - enables real-time synchronization across distributed frontends
   - requires minimal configuration (topic exchange)

3. **dynamicProxy Frontend** (runs via `zrok access dynamicProxy`)
   - HTTP/HTTPS listener serving dynamic share mappings
   - AMQP subscriber receiving mapping updates
   - gRPC client for initial mapping queries and on-demand lookups
   - optional OAuth integration for authentication

```mermaid
graph TB
    subgraph "zrok Controller"
        DPC[dynamicProxyController<br/>gRPC + AMQP Publisher]
    end

    subgraph "Message Broker"
        AMQP[RabbitMQ<br/>Topic Exchange]
    end

    subgraph "Frontend Instances"
        DP1[dynamicProxy 1<br/>gRPC Client + AMQP Subscriber]
        DP2[dynamicProxy 2<br/>gRPC Client + AMQP Subscriber]
        DP3[dynamicProxy N<br/>gRPC Client + AMQP Subscriber]
    end

    subgraph "Users"
        U1[User Shares]
        U2[User Shares]
    end

    DPC -->|publishes mapping updates| AMQP
    AMQP -->|broadcasts updates| DP1
    AMQP -->|broadcasts updates| DP2
    AMQP -->|broadcasts updates| DP3

    DP1 -.->|queries mappings via gRPC| DPC
    DP2 -.->|queries mappings via gRPC| DPC
    DP3 -.->|queries mappings via gRPC| DPC

    U1 -->|creates shares| DPC
    U2 -->|creates shares| DPC

    Internet[Internet Traffic] --> DP1
    Internet --> DP2
    Internet --> DP3
```

## Prerequisites

Before setting up `dynamicProxy`, ensure you have:

- **RabbitMQ**: AMQP 0.9.1 compatible message broker (version 3.x or higher recommended)
- **Ziti Controller Access**: administrative access to create identities, services, and policies
- **Running zrok Controller**: v2.0 or higher with administrative credentials
- **DNS Configuration**: ability to create DNS records for your frontend domain
- **TLS Certificates**: for HTTPS frontends (Let's Encrypt recommended)

### Installing RabbitMQ

RabbitMQ is available through most package managers. Here are installation options:

#### Docker (Recommended for Testing)

```bash
docker run -d --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

The management UI will be available at `http://localhost:15672` (default credentials: `guest`/`guest`).

#### Ubuntu/Debian

```bash
sudo apt-get install rabbitmq-server
sudo systemctl enable rabbitmq-server
sudo systemctl start rabbitmq-server
```

#### RHEL/CentOS/Fedora

```bash
sudo dnf install rabbitmq-server
sudo systemctl enable rabbitmq-server
sudo systemctl start rabbitmq-server
```

For production deployments, consult the [RabbitMQ documentation](https://www.rabbitmq.com/download.html) for best practices on clustering, persistence, and monitoring.

## Fresh Installation

This section covers setting up dynamicProxy from scratch on a new zrok instance.

### Step 1: Create dynamicProxyController Identity

First, create a Ziti identity for the dynamicProxyController service:

```bash
zrok admin create identity dynamicProxyController
```

This command outputs the identity details. Save the identity's Ziti ID (starts with a letter, like `aBc123`), as you'll need it for the next steps.

The identity configuration file is saved to `~/.zrok/identities/dynamicProxyController.json` by default.

### Step 2: Create Ziti Service and Policies

Using the Ziti CLI, create a service for the dynamicProxyController and the necessary policies:

```bash
# Set variables for convenience
CONTROLLER_ZID="<ziti-id-from-step-1>"
SERVICE_NAME="dynamicProxyController"

# Create the service
ziti edge create service "$SERVICE_NAME"

# Create Service Edge Router Policy (allows edge routers to see the service)
ziti edge create serp "${SERVICE_NAME}-serp" \
  --edge-router-roles '#all' \
  --service-roles "@${SERVICE_NAME}"

# Create Service Policy for Bind (allows controller identity to bind the service)
ziti edge create sp "${SERVICE_NAME}-bind" Bind \
  --identity-roles "@${CONTROLLER_ZID}" \
  --service-roles "@${SERVICE_NAME}"

# Create Service Policy for Dial (allows frontend to dial the service)
# Note: Replace 'public' with your frontend identity name if different
ziti edge create sp "${SERVICE_NAME}-dial" Dial \
  --identity-roles "@public" \
  --service-roles "@${SERVICE_NAME}"
```

:::note
The `public` identity referenced in the dial policy should be the identity used by your dynamicProxy frontends. If you use a different identity name, adjust the command accordingly. You can create multiple dial policies for different frontend identities if needed.
:::

### Step 3: Configure zrok Controller

Add the dynamicProxyController configuration to your zrok controller configuration file (typically `etc/dev.yml` or your production config):

```yaml
# Add to your existing controller configuration
dynamic_proxy_controller:
  identity_path: /home/zrok/.zrok/identities/dynamicProxyController.json
  service_name: dynamicProxyController

  amqp_publisher:
    url: amqp://guest:guest@localhost:5672
    exchange_name: dynamicProxy
```

Configuration parameters:

- **`identity_path`**: absolute path to the dynamicProxyController identity JSON file created in Step 1
- **`service_name`**: name of the Ziti service created in Step 2 (must match)
- **`amqp_publisher.url`**: AMQP connection URL for RabbitMQ (format: `amqp://username:password@host:port`)
- **`amqp_publisher.exchange_name`**: AMQP exchange name for mapping updates (can be any name, must match frontend config)

Restart your zrok controller to apply the configuration:

```bash
# If using systemd
sudo systemctl restart zrok-controller

# If running manually
# Stop the controller (Ctrl+C) and restart it
zrok controller etc/ctrl.yml
```

### Step 4: Create a Namespace

Namespaces provide logical grouping for names. Create your first namespace:

```bash
# Create an open namespace (anyone can create names in it)
zrok admin create namespace --token public --open zrok.example.com

# Or create a closed namespace (requires grants)
zrok admin create namespace --token private private.example.com
```

The command outputs a namespace token (e.g., `abc123xyz`). Save this token as you'll need it for frontend mapping and when creating names.

Parameters:
- `--token` specifies the namespace token used to refer to this specific namespace (the identifier)
- `--open`: open mode (users can create names without grants); without `--open` is "closed mode" where users need explicit grants to create names in the namespace (use `zrok admin create namespace-grant`)
- And the name (`zrok.example.com`) is the DNS name for the namespace; names will end up being `myshare.zrok.example.com`

### Step 5: Create a Dynamic Frontend

Create a frontend with the dynamic flag enabled:

```bash
zrok admin create frontend --dynamic public "http://{token}.zrok.example.com:8080"
```

This creates a dynamic frontend and outputs a frontend token (e.g., `KMmfE0VXO7Pp`).

Command breakdown:
- **`--dynamic`**: enables dynamic mode (required for dynamicProxy)
- **`public`**: Ziti identity that will bind the frontend (must have dial permissions to dynamicProxyController service)
- **`"http://{token}.zrok.example.com:8080"`**: URL template for shares
  - This is the legacy URL template used with legacy `publicProxy` frontends; disregarded by new v2 `dynamicProxy` installations and will be removed in future versions of zrok
  - `{token}` is replaced with the share token or name
  - can use `https` if you configure TLS in the frontend config
  

### Step 6: Map Namespace to Frontend

Link the namespace to the frontend so shares in the namespace are served by this frontend:

```bash
zrok admin create namespace-frontend <namespaceToken> <frontendToken>
```

You can map multiple frontends to a single namespace for load balancing and high availability. All mapped frontends will receive the same mapping updates via AMQP.

For simple horizontal scalability (behind a load balancer) you can create one logical frontend in the zrok controller, and share the frontendToken and identity amongst multiple `zrok access dynamicProxy` instances

To view existing mappings:

```bash
# List frontends for a namespace
zrok admin list namespaceFrontendMappings <namespaceToken>

# List namespaces for a frontend
zrok admin list frontendNamespaceMappings <frontendToken>
```

### Step 7: Configure dynamicProxy Frontend

Create a configuration file for the dynamicProxy frontend (e.g., `etc/dynamicProxy.yml`):

```yaml
v: 1

frontend_token: <frontendToken-from-step-5>
identity: public
bind_address: "0.0.0.0:8080"
mapping_refresh_interval: 1m

amqp_subscriber:
  url: amqp://guest:guest@localhost:5672
  exchange_name: dynamicProxy

controller:
  identity_path: /home/zrok/.zrok/identities/public.json
  service_name: dynamicProxyController
```

Configuration parameters:

#### Required Parameters

- **`v`**: configuration version (always `1`)
- **`frontend_token`**: token from `zrok admin create frontend` (Step 5)
- **`amqp_subscriber.url`**: RabbitMQ connection URL (must match controller's `amqp_publisher.url`)
- **`amqp_subscriber.exchange_name`**: AMQP exchange name (must match controller's `amqp_publisher.exchange_name`)
- **`controller.identity_path`**: path to the Ziti identity JSON file for the frontend
- **`controller.service_name`**: Ziti service name for dynamicProxyController (must match Step 2 and controller config)

#### Optional Parameters

- **`identity`**: Ziti identity name to use (default: `"public"`)
- **`bind_address`**: IP and port where the frontend listens (default: `"0.0.0.0:8080"`)
- **`mapping_refresh_interval`**: how often to audit (and potentially refresh) mappings from controller (default: `5m`)

#### OAuth Configuration (Optional)

To enable OAuth authentication, add an `oauth` section. For detailed OAuth setup, see the [OAuth Configuration Guide](oauth/configuring-oauth.md).

```yaml
oauth:
  bind_address: "0.0.0.0:8181"
  endpoint_url: "https://oauth.zrok.example.com"
  cookie_name: "zrok-auth-session"
  cookie_domain: "zrok.example.com"
  session_lifetime: "6h"
  intermediate_lifetime: "5m"
  signing_key: "your-unique-signing-key-32-chars-min"
  encryption_key: "your-unique-encryption-key-24-chars-min"

  providers:
    - name: "google"
      type: "google"
      client_id: "<google-client-id>"
      client_secret: "<google-client-secret>"
```

#### TLS Configuration (Optional)

For HTTPS frontends, add a `tls` section:

```yaml
tls:
  cert_path: /etc/letsencrypt/live/zrok.example.com/fullchain.pem
  key_path: /etc/letsencrypt/live/zrok.example.com/privkey.pem
```

### Step 8: Start dynamicProxy Frontend

Run the dynamicProxy frontend using your configuration:

```bash
zrok access dynamicProxy etc/dynamicProxy.yml
```

The frontend will:
1. Connect to the Ziti network using the configured identity
2. Subscribe to the AMQP exchange for mapping updates
3. Query the controller via gRPC for initial mappings
4. Start serving HTTP/HTTPS traffic on the configured bind address

For production deployments, consider using a process manager like systemd:

```ini
# /etc/systemd/system/zrok-dynamicproxy.service
[Unit]
Description=zrok Dynamic Proxy Frontend
After=network.target rabbitmq-server.service

[Service]
Type=simple
User=zrok
Group=zrok
WorkingDirectory=/home/zrok
ExecStart=/usr/local/bin/zrok access dynamicProxy /home/zrok/etc/dynamicProxy.yml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable zrok-dynamicproxy
sudo systemctl start zrok-dynamicproxy
sudo systemctl status zrok-dynamicproxy
```

### Step 9: Test with a Share

Create a test share to verify the setup:

```bash
# First, enable a zrok environment if you haven't already
zrok enable <accountToken>

# Set the default namespace for convenience
zrok config set defaultNamespace <namespaceToken>

# Create a reserved name in the namespace
zrok create name -n <namespaceToken> my-test-share

# Share a resource using the reserved name
zrok share public --backend-mode web -n <namespaceToken>:my-test-share ~/public
```

Access your share at `http://my-test-share.zrok.example.com:8080` (adjust domain/port based on your configuration).

If you see your shared content, congratulations! Your dynamicProxy setup is working correctly.

## Migrating from publicProxy

If you're currently running `zrok access public` (publicProxy) and want to migrate to dynamicProxy, you have two main approaches:

### Option 1: Side-by-Side Operation (Gradual Migration)

This approach allows both publicProxy and dynamicProxy to run simultaneously, enabling gradual user migration with no downtime.

#### Setup Process

1. **Complete Steps 1-8** from the Fresh Installation section without shutting down your existing publicProxy
2. **Create separate DNS entries** for the new dynamicProxy frontend (e.g., `v2.zrok.example.com`)
3. **Communicate the migration** to your users:
   - users on v1.x clients can continue using the old publicProxy
   - users on v2.0+ clients can start using namespaces and names with dynamicProxy
   - provide a migration deadline

#### User Migration Process

Users need to:

1. Upgrade to zrok v2.0+
2. Set the default namespace: `zrok config set defaultNamespace <namespaceToken>`
3. Recreate their shares using names:
   ```bash
   # Instead of: zrok reserve public 8080
   # They use:
   zrok create name my-app
   zrok share public localhost:8080 -n <namespaceToken>:my-app
   ```

#### Deprecation Timeline

Once user adoption reaches your target threshold:

1. Announce the publicProxy shutdown date (recommend 30+ days notice)
2. Monitor publicProxy usage via controller metrics
3. Contact remaining users directly if possible
4. Shut down publicProxy and remove DNS entries
5. Update documentation to remove v1.x references

### Option 2: Clean Cutover (With Downtime)

This approach requires downtime but provides a clean migration in a single maintenance window.

#### Preparation (Before Maintenance Window)

1. **Complete Steps 1-7** from the Fresh Installation section
2. **Test dynamicProxy configuration** thoroughly in a staging environment
3. **Notify all users** of the scheduled maintenance window
4. **Prepare migration documentation** for users

#### During Maintenance Window

1. **Shut down publicProxy** (`zrok access public` processes)
2. **Start dynamicProxy** with the same DNS entries as the old publicProxy
3. **Notify users** to upgrade to v2.0+ and recreate their shares using names
4. **Provide support channels** for migration assistance

#### Post-Migration

All users must:

1. Upgrade to zrok v2.0+
2. Set default namespace: `zrok config set defaultNamespace <namespaceToken>`
3. Recreate shares as shown in Option 1

:::warning
Clean cutover means existing v1.x shares will stop working immediately. Only choose this approach if you can ensure all users upgrade and recreate their shares promptly, or if downtime is acceptable.
:::

### Comparing Migration Approaches

| Aspect | Side-by-Side | Clean Cutover |
|--------|--------------|---------------|
| **Downtime** | None | Required |
| **Complexity** | Higher (manage two systems) | Lower (single system) |
| **User Impact** | Gradual, self-paced | Immediate, forced |
| **Resource Usage** | Higher (duplicate frontends) | Lower (single frontend) |
| **Rollback** | Easy (keep publicProxy) | Difficult |
| **Timeline** | Weeks to months | Single maintenance window |

Choose side-by-side for larger deployments with diverse users, or clean cutover for smaller deployments with coordinated users.

## Configuration Reference

### Complete Configuration Example

Here's a complete `dynamicProxy` configuration with all available options:

```yaml
v: 1

# Frontend identification
frontend_token: KMmfE0VXO7Pp
identity: public
bind_address: "0.0.0.0:8080"

# Mapping refresh interval
mapping_refresh_interval: 1m

# AMQP subscriber for receiving mapping updates
amqp_subscriber:
  url: amqp://guest:guest@localhost:5672
  exchange_name: dynamicProxy
  queue_depth: 1024  # Optional: buffer size for mapping updates (default: 1024)

# Controller client for gRPC queries
controller:
  identity_path: /home/zrok/.zrok/identities/public.json
  service_name: dynamicProxyController
  timeout: 30s  # Optional: gRPC request timeout (default: 30s)

# Optional: TLS configuration for HTTPS
tls:
  cert_path: /etc/letsencrypt/live/zrok.example.com/fullchain.pem
  key_path: /etc/letsencrypt/live/zrok.example.com/privkey.pem

# Optional: Interstitial page configuration
interstitial:
  enabled: true
  html_path: /etc/zrok/interstitial.html
  user_agent_prefixes:
    - "Mozilla/"
    - "Chrome/"

# Optional: OAuth configuration
oauth:
  bind_address: "127.0.0.1:8181"
  endpoint_url: "https://oauth.zrok.example.com:8443"
  cookie_name: "zrok-auth-session"
  cookie_domain: "zrok.example.com"
  session_lifetime: "6h"
  intermediate_lifetime: "5m"
  signing_key: "your-32-plus-character-signing-key-here"
  encryption_key: "your-24-plus-character-encryption-key-here"
  max_cookie_size: 3072  # Optional: maximum cookie size in bytes (default: 3072)

  providers:
    # Google OAuth
    - name: "google"
      type: "google"
      client_id: "your-google-client-id"
      client_secret: "your-google-client-secret"

    # GitHub OAuth
    - name: "github"
      type: "github"
      client_id: "your-github-client-id"
      client_secret: "your-github-client-secret"

    # Generic OIDC
    - name: "microsoft"
      type: "oidc"
      client_id: "your-oidc-client-id"
      client_secret: "your-oidc-client-secret"
      scopes: ["openid", "email", "profile"]
      issuer: "https://login.microsoftonline.com/<tenant-id>/v2.0"
      supports_pkce: true
```

For detailed OAuth provider configuration, see the [OAuth Configuration Guide](oauth/configuring-oauth.md).

## Namespace Management

After setting up dynamicProxy, you'll need to manage namespaces and names for your users.

### Creating Namespaces

Namespaces can be **open** (anyone can create names) or **closed** (requires grants):

```bash
# Open namespace - users can freely create names
zrok admin create namespace public "Public Shares" -o

# Closed namespace - requires explicit grants
zrok admin create namespace private "Private Shares"
```

### Granting Namespace Access

For closed namespaces, grant users access by email:

```bash
zrok admin add namespaceGrant <namespaceToken> user@example.com
```

Remove grants with:

```bash
zrok admin remove namespaceGrant <namespaceToken> user@example.com
```

### Setting Default Namespace

Users can set their default namespace to avoid specifying it on every command:

```bash
zrok config set defaultNamespace <namespaceToken>
```

Or via environment variable:

```bash
export ZROK_DEFAULT_NAMESPACE=<namespaceToken>
```

### Listing Namespaces

Users can see available namespaces (based on grants and open namespaces):

```bash
zrok list namespaces
```

Administrators can list all namespaces:

```bash
zrok admin list namespaces
```

## Troubleshooting

### Frontend Not Receiving Mapping Updates

**Symptoms**: shares are created but don't resolve, frontend logs show no mapping updates

**Diagnostics**:

1. Verify AMQP connectivity:
   ```bash
   # Test RabbitMQ connection
   curl -u guest:guest http://localhost:15672/api/overview
   
   # Check if exchange exists
   curl -u guest:guest http://localhost:15672/api/exchanges/%2F/dynamicProxy
   ```

2. Check controller logs for AMQP publisher errors:
   ```bash
   # If using systemd
   sudo journalctl -u zrok-controller -f
   ```

3. Verify frontend logs for AMQP subscriber connection:
   ```bash
   # If using systemd
   sudo journalctl -u zrok-dynamicproxy -f
   ```

**Solutions**:
- Ensure `exchange_name` matches in both controller and frontend configs
- Verify RabbitMQ is running and accessible from both controller and frontend hosts
- Check firewall rules if RabbitMQ is on a separate host

### Frontend Cannot Query Controller via gRPC

**Symptoms**: frontend starts but shows errors querying mappings, shares don't resolve

**Diagnostics**:

1. Verify Ziti connectivity:
   ```bash
   # Test if frontend can reach dynamicProxyController service
   ziti edge list services --name dynamicProxyController
   ```

2. Check service policies:
   ```bash
   # Verify bind policy exists
   ziti edge list service-policies --name dynamicProxyController-bind
   
   # Verify dial policy exists for frontend identity
   ziti edge list service-policies --name dynamicProxyController-dial
   ```

3. Check controller logs for gRPC server errors:
   ```bash
   sudo journalctl -u zrok-controller | grep -i grpc
   ```

**Solutions**:
- Ensure `service_name` matches in controller config, frontend config, and Ziti service name
- Verify the frontend identity has dial permissions (Step 2)
- Confirm controller identity has bind permissions (Step 2)
- Check that both identities are enrolled and connected to Ziti

### Shares Don't Resolve (404 Errors)

**Symptoms**: accessing `http://<name>.zrok.example.com` returns 404

**Diagnostics**:

1. Verify namespace-to-frontend mapping exists:
   ```bash
   zrok admin list namespace-frontend <namespaceToken>
   ```

2. Check if the name exists:
   ```bash
   zrok list names
   ```

3. Verify the share is active:
   ```bash
   zrok list shares
   ```

4. Check frontend logs for mapping presence:
   ```bash
   sudo journalctl -u zrok-dynamicproxy -f
   ```

**Solutions**:
- Create namespace-to-frontend mapping if missing (Step 6)
- Verify the share is using the correct namespace token
- Restart the share if it was created before the mapping
- Check that the frontend token in config matches the frontend created in Step 5

### DNS Not Resolving

**Symptoms**: domain names don't resolve or resolve to wrong IP

**Diagnostics**:

```bash
# Test DNS resolution
dig my-share.zrok.example.com

# Check if DNS propagates correctly
nslookup my-share.zrok.example.com
```

**Solutions**:
- Ensure wildcard DNS record exists: `*.zrok.example.com -> <frontend-ip>`
- Wait for DNS propagation (can take minutes to hours)
- Use `/etc/hosts` for local testing: `127.0.0.1 my-share.zrok.example.com`

### TLS Certificate Errors

**Symptoms**: browser shows certificate warnings, HTTPS connections fail

**Solutions**:
- Use a wildcard certificate for `*.zrok.example.com`
- Verify certificate paths in `tls` configuration are correct and readable
- For Let's Encrypt, use DNS challenge for wildcard certificates:
  ```bash
  certbot certonly --manual --preferred-challenges dns \
    -d '*.zrok.example.com' -d 'zrok.example.com'
  ```

### High Memory Usage

**Symptoms**: dynamicProxy or controller consuming excessive memory

**Diagnostics**:

```bash
# Check memory usage
ps aux | grep zrok

# Monitor AMQP queue depth
curl -u guest:guest http://localhost:15672/api/queues/%2F/<queue-name>
```

**Solutions**:
- Reduce `mapping_refresh_interval` if it's too frequent
- Adjust `queue_depth` in AMQP subscriber config
- Consider multiple frontend instances with load balancing instead of increasing resources
- Monitor and limit the number of active shares per namespace

## Advanced Configuration

### Multiple Frontend Instances

For high availability and load balancing, run multiple dynamicProxy instances:

1. Create separate configuration files for each instance (with unique bind addresses if on the same host)
2. Map the same namespace to multiple frontends:
   ```bash
   zrok admin create namespace-frontend <namespaceToken> <frontendToken1>
   zrok admin create namesapce-frontend <namespaceToken> <frontendToken2>
   ```
3. Use a load balancer (nginx, HAProxy, etc.) in front of the instances
4. Each instance will receive mapping updates via AMQP independently

### Custom Exchange Names

You can use different AMQP exchange names for different frontend groups:

```yaml
# Controller config
dynamic_proxy_controller:
  amqp_publisher:
    exchange_name: dynamicProxy-prod
```

```yaml
# Frontend config
amqp_subscriber:
  exchange_name: dynamicProxy-prod
```

This allows you to isolate mapping updates between different deployment environments (prod, staging, etc.).

### Monitoring and Metrics

Integrate with your monitoring stack:

1. **RabbitMQ metrics**: Use the management plugin to monitor queue depth, message rates, and connections
2. **zrok controller metrics**: if you've configured InfluxDB metrics (see [Configuring Metrics](metrics-and-limits/configuring-metrics.md)), monitor:
   - Frontend registration count
   - Mapping update frequency
   - gRPC request latency
3. **Frontend logs**: parse structured logs for error rates and performance data

## Additional Resources

- [OAuth Configuration Guide](oauth/configuring-oauth.md) - detailed OAuth provider setup
- [Configuring Metrics](metrics-and-limits/configuring-metrics.md) - InfluxDB integration for monitoring
- [Interstitial Page Configuration](interstitial-page.md) - customizing the interstitial page
- [Error Pages](error-pages.md) - customizing error pages

## Getting Help

If you encounter issues not covered in this guide:

1. Check the [zrok GitHub Issues](https://github.com/openziti/zrok/issues) for similar problems
2. Enable verbose logging: set the `DL_USE_JSON=true` environment variable for structured logs
3. Join the [OpenZiti Community](https://openziti.discourse.group/) for support
4. Review your configuration against the complete example in the Configuration Reference section
