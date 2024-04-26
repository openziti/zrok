# Use the official Caddy image as a parent image
FROM caddy:2-builder AS builder

# e.g., "github.com/caddy-dns/digitalocean"
ARG CADDY_DNS_PLUGIN

# Build Caddy with the specified DNS provider plugin
RUN xcaddy build \
    --with github.com/caddy-dns/${CADDY_DNS_PLUGIN}

# Use the official Caddy image to create the final image
FROM caddy:2

# install curl to support using the Caddy API
RUN apk add --no-cache curl

# Copy the custom Caddy build into the final image
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY ./Caddyfile /etc/caddy/Caddyfile
