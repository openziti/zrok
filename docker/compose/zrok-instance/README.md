## Docker Instance

<iframe width="100%" height="315" src="https://www.youtube.com/embed/70zJ_h4uiD8" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

This Docker Compose project creates a zrok instance supported by a OpenZiti controller and router. It supports flexible deployment configurations:

1. **Basic Configuration**: Services exposed on localhost only (no TLS)
2. **With Caddy**: Services published using Caddy (TLS)
3. **With Traefik**: Services published using Traefik (TLS)

### Create the Docker Compose Project

Create a working directory on your Docker host and save these Docker Compose project files.

#### YOLO

1. Run this script to download the files in the current directory.

    ```bash
    curl https://get.openziti.io/zrok-instance/fetch.bash | bash
    ```

    Or, specify the Compose project directory.
    
    ```bash
    curl https://get.openziti.io/zrok-instance/fetch.bash | bash -s /path/to/compose/project/dir
    ```

#### I'll Do it Myself

1. Get the zrok repo ZIP file.

    ```bash
    wget https://github.com/openziti/zrok/archive/refs/heads/main.zip
    ```

1. Unzip the zrok-instance files into the project directory.

    ```bash
    unzip -j -d . main.zip '*/docker/compose/zrok-instance/*'
    ```

### Basic Configuration (No TLS, Localhost Only)

This is the simplest way to get started with zrok, exposing services on localhost only, without TLS.

#### DNS Configuration (Optional for localhost-only setup)

1. If you plan to use this beyond localhost, set up a wildcard record for the IP address where the zrok instance will run 
   (e.g., if your DNS zone is `share.example.com`, then your wildcard record is `*.share.example.com`).

#### Configure the Docker Compose Project Environment

Create an `.env` file in the working directory with the minimal required configuration:

```bash title=".env minimal configuration"
# Required settings
ZROK_DNS_ZONE=share.example.com
ZROK_USER_EMAIL=me@example.com
ZROK_USER_PWD=zrokuserpw
ZITI_PWD=zitiadminpw
ZROK_ADMIN_TOKEN=zroktoken

# Expose services only on localhost (default)
ZROK_INSECURE_INTERFACE=127.0.0.1

# Service ports
ZROK_CTRL_PORT=18080
ZROK_FRONTEND_PORT=8080
ZROK_OAUTH_PORT=8081
ZITI_CTRL_ADVERTISED_PORT=80
ZITI_ROUTER_PORT=3022
```

#### Start the Docker Compose Project

Start the zrok instance:

```bash
docker compose up --build --detach
```

### Expanded Configuration with TLS (Caddy or Traefik)

For production deployments, you should use TLS. You can choose between Caddy or Traefik for TLS termination and reverse proxy to the zrok services. The ziti services are always published directly, not proxied, and they bring their own TLS.

#### DNS Configuration for TLS

1. Ensure a wildcard record exists for the IP address where the zrok instance will run
   (e.g., if your DNS zone is `share.example.com`, then your wildcard record is `*.share.example.com`).

2. Choose a DNS provider that supports automatic DNS challenge for obtaining wildcard certificates and for which a plugin is available in Caddy or Traefik.

#### Configure the Docker Compose File

Add this setting to your `.env` file to select which TLS provider to use:

```bash
# Use one of the following:
COMPOSE_FILE=compose.yml:compose.caddy.yml  # For Caddy
# OR
COMPOSE_FILE=compose.yml:compose.traefik.yml  # For Traefik
```

#### Caddy Configuration

If using Caddy, add these settings to your `.env` file:

```bash title=".env for Caddy"
# Caddy TLS configuration
CADDY_DNS_PLUGIN=cloudflare  # Plugin name for your DNS provider (see github.com/caddy-dns)
CADDY_DNS_PLUGIN_TOKEN=abcd1234  # API token from your DNS provider
CADDY_ACME_API=https://acme-v02.api.letsencrypt.org/directory  # ACME API endpoint
CADDY_HTTPS_PORT=443  # HTTPS port (optional, defaults to 443)
CADDY_INTERFACE=0.0.0.0  # Interface to bind to (optional, defaults to all interfaces)

# For AWS Route53, uncomment and set these instead of CADDY_DNS_PLUGIN_TOKEN:
# AWS_ACCESS_KEY_ID=your-access-key
# AWS_SECRET_ACCESS_KEY=your-secret-key
# AWS_REGION=your-region
# AWS_SESSION_TOKEN=your-session-token  # Only if using temporary credentials
```

#### Traefik Configuration

If using Traefik, add these settings to your `.env` file:

```bash title=".env for Traefik"
# Traefik TLS configuration
TRAEFIK_DNS_PROVIDER=digitalocean  # DNS provider for Traefik
TRAEFIK_DNS_PROVIDER_TOKEN=abcd1234  # API token from your DNS provider
TRAEFIK_ACME_API=https://acme-v02.api.letsencrypt.org/directory  # ACME API endpoint
TRAEFIK_HTTPS_PORT=443  # HTTPS port (optional, defaults to 443)
TRAEFIK_INTERFACE=0.0.0.0  # Interface to bind to (optional, defaults to all interfaces)

# For AWS Route53, uncomment and set these instead of TRAEFIK_DNS_PROVIDER_TOKEN:
# AWS_ACCESS_KEY_ID=your-access-key
# AWS_SECRET_ACCESS_KEY=your-secret-key
# AWS_REGION=your-region
# AWS_SESSION_TOKEN=your-session-token  # Only if using temporary credentials
```

#### Start the Docker Compose Project

Start the zrok instance with TLS support:

```bash
docker compose up --build --detach
```

### Set up a User Account

This step creates a user account. You will log in to the zrok web console with the account password created in this step. The ZROK_USER_EMAIL and ZROK_USER_PWD variables are set in the `.env` file.

```bash title="Create the first user account"
docker compose exec zrok-controller bash -xc 'zrok admin create account ${ZROK_USER_EMAIL} ${ZROK_USER_PWD}'
```

```buttonless title="Example output"
+ zrok admin create account me@example.com zrokuserpw
[   0.000]    INFO zrok/controller/store.Open: database connected
[   0.002]    INFO zrok/controller/store.(*Store).migrate: applied 0 migrations
heMqncCyxZcx
```

Create additional users by running the command again with a different email and password.

```bash title="Create another user"
docker compose exec zrok-controller zrok admin create account <email> <password>
```

### Enable the User Environment

You must enable each device environment with the account token obtained when the account was created. This is separate from the account password that's used to log in to the web console.

Follow [the getting started guide](/docs/getting-started#installing-the-zrok-command) to install the zrok CLI on some device and enable a zrok environment.

1. Configure the environment with the zrok API endpoint:

   ```bash
   # If using TLS (Caddy or Traefik)
   zrok config set apiEndpoint https://zrok.share.example.com
   
   # If using basic configuration (localhost, no TLS)
   zrok config set apiEndpoint http://localhost:18080
   ```

2. Enable an environment on this device with the account token from the previous step.

   ```bash
   zrok enable heMqncCyxZcx
   ```

### Firewall Configuration

- `443/tcp` - HTTPS for all services (Caddy or Traefik)
- `80/tcp` - ziti ctrl plane
- `3022/tcp` - ziti data plane

### Additional Configuration Options

You can add these additional settings to your `.env` file for more customization:

```bash
# OAuth configuration for public shares
ZROK_OAUTH_HASH_KEY=oauthhashkeysecret
ZROK_OAUTH_GITHUB_CLIENT_ID=abcd1234
ZROK_OAUTH_GITHUB_CLIENT_SECRET=abcd1234
ZROK_OAUTH_GOOGLE_CLIENT_ID=abcd1234
ZROK_OAUTH_GOOGLE_CLIENT_SECRET=abcd1234
```

#### ⚠️ OAuth Configuration Note

**If you are NOT using OAuth for public shares**, remove the entire OAuth section from `zrok-frontend-config.yml.envsubst` to avoid configuration errors:

```yaml
# Remove this entire section if not using OAuth
oauth:
  bind_address: 0.0.0.0:${ZROK_OAUTH_PORT}
  endpoint_url: https://oauth.${ZROK_DNS_ZONE}
  cookie_name: zrok-auth-session
  cookie_domain: ${ZROK_DNS_ZONE}
  session_lifetime: 6h
  intermediate_lifetime: 5m
  signing_key: ${ZROK_OAUTH_HASH_KEY}
  encryption_key: ${ZROK_OAUTH_HASH_KEY}
  providers:
    - name: github
      type: github
      client_id: ${ZROK_OAUTH_GITHUB_CLIENT_ID}
      client_secret: ${ZROK_OAUTH_GITHUB_CLIENT_SECRET}

    - name: google
      type: google
      client_id: ${ZROK_OAUTH_GOOGLE_CLIENT_ID}
      client_secret: ${ZROK_OAUTH_GOOGLE_CLIENT_SECRET}
```

Then rebuild: `docker compose up -d --build zrok-frontend`

### Troubleshooting

1. Check the service logs:

   ```bash
   # View logs for a specific service
   docker compose logs zrok-controller
   docker compose logs zrok-frontend
   docker compose logs ziti-quickstart
   
   # View logs for Caddy (if using)
   docker compose logs caddy
   
   # View logs for Traefik (if using)
   docker compose logs traefik
   ```

2. Validate TLS configuration:

   ```bash
   # For Caddy
   docker compose exec caddy caddy validate --config /etc/caddy/Caddyfile
   
   # For Traefik
   docker compose exec traefik traefik healthcheck
   ```

3. Check certificate status:

   ```bash
   # For Caddy
   docker compose exec caddy curl -s "http://localhost:2019/certificates"
   
   # For Traefik - view the ACME certificate file directly
   docker compose exec traefik cat /etc/traefik/acme/acme.json | grep -A 5 "Certificates"
   ```
