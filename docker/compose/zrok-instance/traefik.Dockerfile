# Use the official Traefik image
FROM traefik:v2.10

# Install curl for healthcheck
RUN apk add --no-cache curl

# Create necessary directories
RUN mkdir -p /etc/traefik/dynamic /etc/traefik/acme

# Add configuration file
COPY ./traefik.dynamic.toml /etc/traefik/dynamic/traefik.toml

# Create and set permissions for the ACME certificates storage
RUN touch /etc/traefik/acme/acme.json && chmod 600 /etc/traefik/acme/acme.json

HEALTHCHECK --interval=5s --timeout=3s --retries=3 \
  CMD traefik healthcheck
