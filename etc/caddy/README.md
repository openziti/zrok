
# Caddyfile Samples

The Caddyfile samples in this directory are for use with `--backend-mode caddy ./my.Caddyfile` which runs an embedded
Caddy server.

The Caddyfile must have this structure because it is rendered as a Go template by zrok to bind the HTTP listener.

```console
http:// {
    bind {{ .ZrokBindAddress }}
    # customize reverse_proxy, file_server, etc.
}
```

## Notes

multiple_upstream.Caddyfile is bundled in the zrok2-agent package for Linux as an example Caddyfile.
