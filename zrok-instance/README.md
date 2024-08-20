
## Docker Instance

<iframe width="100%" height="315" src="https://www.youtube.com/embed/70zJ_h4uiD8" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

This Docker Compose project creates a zrok instance and includes a ziti controller and router. An optional Caddy container is included to provide HTTPS and reverse proxy services for the zrok API and public shares.

### DNS Configuration

1. A wildcard record exists for the IP address where the zrok instance will run, e.g. if your DNS zone is `share.example.com`, then your wildcard record is `*.share.example.com`.

#### Additional DNS Configuration for Caddy TLS

The included Caddy container can automatically manage a wildcard certificate for your zrok instance. You can enable Caddy in this compose project by renaming `compose.caddy.yml` as `compose.override.yml`.

1. Ensure A Caddy DNS plugin is available for your DNS provider (see [github.com/caddy-dns](https://github.com/orgs/caddy-dns/repositories?type=all&q=sort%3Aname-asc)).
1. Designate A DNS zone for zrok, e.g. `example.com` or `share.example.com` and create the zone on your DNS provider's platform.
1. Created an API token in your DNS provider that has permission to manage zrok's DNS zone.

### Create the Docker Compose Project

Create a working directory on your Docker host and save these Docker Compose project files.

#### Shortcut Option

1. Run this script to download the files in the current directory.

    ```bash
    curl https://get.openziti.io/zrok-instance/fetch.bash | bash
    ```

    Or, specify the Compose project directory.
    
    ```bash
    curl https://get.openziti.io/zrok-instance/fetch.bash | bash -s /path/to/compose/project/dir
    ```

#### Manual Option

1. Get the zrok repo ZIP file.

    ```bash
    wget https://github.com/openziti/zrok/archive/refs/heads/main.zip
    ```

1. Unzip the zrok-instance files into the project directory.

    ```bash
    unzip -j -d . main.zip '*/docker/compose/zrok-instance/*'
    ```

### Configure the Docker Compose Project Environment

Create an `.env` file in the working directory.

```bash title=".env required"
ZROK_DNS_ZONE=share.example.com

ZROK_USER_EMAIL=me@example.com
ZROK_USER_PWD=zrokuserpw

ZITI_PWD=zitiadminpw
ZROK_ADMIN_TOKEN=zroktoken
```

```bash title=".env options"
# Caddy TLS option: rename compose.caddy.yml to compose.override.yml and set these vars; allow 80,443 in firewall
# plugin name for your DNS provider
CADDY_DNS_PLUGIN=cloudflare
# API token from your DNS provider
CADDY_DNS_PLUGIN_TOKEN=abcd1234
# use the staging API until you're sure everything is working to avoid hitting the rate limit
CADDY_ACME_API=https://acme-staging-v02.api.letsencrypt.org/directory

# no TLS option: publish the insecure ports to the internet and allow them in the firewall 
ZROK_INSECURE_INTERFACE=0.0.0.0
ZROK_CTRL_PORT=18080
ZROK_FRONTEND_PORT=8080
ZROK_OAUTH_PORT=8081

# ziti ports must be published to the internet and allowed by firewall
ZITI_CTRL_ADVERTISED_PORT=1280
ZITI_ROUTER_PORT=3022

# configure oauth for public shares
ZROK_OAUTH_HASH_KEY=oauthhashkeysecret
ZROK_OAUTH_GITHUB_CLIENT_ID=abcd1234
ZROK_OAUTH_GITHUB_CLIENT_SECRET=abcd1234
ZROK_OAUTH_GOOGLE_CLIENT_ID=abcd1234
ZROK_OAUTH_GOOGLE_CLIENT_SECRET=abcd1234

# zrok version, e.g., 1.0.0
ZROK_CLI_TAG=latest
# ziti version, e.g., 1.0.0
ZITI_CLI_TAG=latest
```

### Start the Docker Compose Project

1. Start the zrok instance.

    The container images for zrok (including caddy) are built in this step. This provides a simple configuration to get started. You can modify the templates named like `*.envsubst` or mount a customized configuration file to mask the one that was built in.

    ```bash
    docker compose up --build --detach
    ```

### Set up a User Account

This step creates a user account. You will log in to the zrok web console with the account password created in this step. The ZROK_USER_EMAIL and ZROK_USER_PWD variables are set in the `.env` file. You can create more user accounts the same way by substituting a different email and password.

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

1. Configure the environment with the zrok API. Substitute the API endpoint with the one you're using, e.g. `https://zrok.${ZROK_DNS_ZONE}`.

    ```bash
    zrok config set apiEndpoint https://zrok.share.example.com
    ```

    or, if not using Caddy for TLS:

    ```bash
    zrok config set apiEndpoint http://zrok.share.example.com:18080
    ```

1. Enable an environment on this device with the account token from the previous step.

    ```bash
    zrok enable heMqncCyxZcx
    ```

### Firewall Configuration

The `ziti-quickstart` and `caddy` containers publish ports to all devices that use zrok shares. The `zrok-controller` and `zrok-frontend` containers expose ports only to the `caddy` container and the Docker host's loopback interface.

#### Required

1. `443/tcp` - reverse proxy handles HTTPS requests for zrok API, OAuth, and public shares (published by container `caddy`)
1. `1280/tcp` - ziti ctrl plane (published by container `ziti-quickstart`)
1. `3022/tcp` - ziti data plane (published by container `ziti-quickstart`)

#### Optional

1. `80/tcp` - reverse proxy redirects non-HTTPS requests to `443/tcp` (published by container `caddy`)
<!-- 1. 443/udp used by Caddy for HTTP/3 QUIC protocol (published by container `caddy`) -->

### Troubleshooting

1. Check the ziti and zrok logs.

    You can substitute the service container name of each to check their logs individually: `ziti-quickstart`, `zrok-controller`, `zrok-frontend`.

    ```bash
    docker compose logs zrok-controller
    ```

1. Check the caddy logs.

    It can take a few minutes for Caddy to obtain the wildcard certificate. You can check the logs to see if there were any errors completing the DNS challenge which involves using the Caddy DNS plugin to create a TXT record in your DNS zone. This leverages the API token you provided in the `.env` file, which must have permission to create DNS records in the zrok DNS zone.

    ```bash
    docker compose logs caddy
    ```

1. `zrok enable` fails certificate verification: ensure you are not using the staging API for Let's Encrypt.

    If you are using the staging API, you will see an error about the API certificate when you use the zrok CLI. You can switch to the production API by removing the overriding assignment of the `CADDY_ACME_API` variable.

    ```buttonless title="Example output"
    there was a problem enabling your environment!
    you are trying to use the zrok service at: https://zrok.share.example.com
    you can change your zrok service endpoint using this command:

    $ zrok config set apiEndpoint <newEndpoint>

    (where newEndpoint is something like: https://some.zrok.io)
    [ERROR]: error creating service client (error getting version from api endpoint 'https://zrok.share.example.com': Get "https://zrok.share.example.com/api/v1/version": tls: failed to verify certificate: x509: certificate signed by unknown authority: Get "https://zrok.share.example.com/api/v1/version": tls: failed to verify certificate: x509: certificate signed by unknown authority)
    ```

1. Validate the Caddyfile.

    ```bash
    docker compose exec caddy caddy validate --config /etc/caddy/Caddyfile
    ```

1. Verify the correct DNS provider module was built-in to Caddy.

    ```bash
    docker compose exec caddy caddy list-modules | grep dns.providers
    ```

    ```buttonless title="Example output"
    dns.providers.cloudflare
    ```

1. Use the Caddy admin API.

    You can use the Caddy admin API to check the status of the Caddy instance. The admin API is available on port `2019/tcp` inside the Docker Compose project. You can modify `compose.override.yml` to publish the port if you want to access the admin API from the Docker host or elsewhere.

    ```bash
    docker compose exec caddy curl http://localhost:2019/config/ | jq
    ```

1. My provider, e.g., Route53 doesn't give me a single API token.

    As long as your DNS provider is supported by Caddy then it will work. You can modify the Caddyfile to use a different set of properties than the example. Here's how the `tls` section should look for Route53.

    ```json
    tls {
      dns {$CADDY_DNS_PLUGIN} {
        access_key_id {$AWS_ACCESS_KEY_ID}
        secret_access_key {$AWS_SECRET_ACCESS_KEY}
      }
    }
    ```

    ```bash title=".env"
    CADDY_DNS_PLUGIN=route53
    AWS_ACCESS_KEY_ID=abcd1234
    AWS_SECRET_ACCESS_KEY=abcd1234
    ```
