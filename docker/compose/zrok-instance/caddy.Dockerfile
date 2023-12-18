# Use the official Caddy image as a parent image
FROM caddy:2-builder AS builder

# Build Caddy with the digitalocean DNS provider
RUN xcaddy build \
    --with github.com/caddy-dns/digitalocean

# Use the official Caddy image to create the final image
FROM caddy:2

# Copy the custom Caddy build into the final image
COPY --from=builder /usr/bin/caddy /usr/bin/caddy

