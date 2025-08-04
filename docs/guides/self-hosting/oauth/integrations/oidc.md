---
sidebar_position: 30
---

# Generic OIDC Setup

This guide covers setting up OpenID Connect (OIDC) providers for your zrok public frontend. OIDC is supported by many identity providers including Keycloak, Auth0, Okta, Azure AD, and others.

## Provider Requirements

Your OIDC provider must support:
- Authorization Code flow
- Discovery endpoint (optional but recommended)
- PKCE (Proof Key for Code Exchange) - optional but recommended for security

## Configure OIDC Provider

1. Create a new OAuth/OIDC client in your provider's admin interface
2. Set the **redirect URI** to: `https://your-oauth-frontend-domain:port/oidc/oauth`
3. Configure required scopes: `openid`, `email`, `profile`
4. Note the **client ID**, **client secret**, and **issuer URL**

## Frontend Configuration

Add the OIDC provider to your `frontend.yml`:

```yaml
oauth:
  providers:
    - name: "my-oidc-provider"
      type: "oidc"
      client_id: "<your-oidc-client-id>"
      client_secret: "<your-oidc-client-secret>"
      scopes: ["openid", "email", "profile"]
      issuer: "https://your-oidc-provider.com"
      supports_pkce: true  # recommended for security
```

### Configuration Options

- **`name`**: Unique identifier for this provider (used in share commands)
- **`type`**: Must be `"oidc"` for OpenID Connect providers
- **`client_id`** and **`client_secret`**: OAuth client credentials from your provider
- **`scopes`**: OAuth scopes to request (typically `["openid", "email", "profile"]`)
- **`issuer`**: The OIDC issuer URL (used for auto-discovery)
- **`discovery_url`**: Optional explicit discovery endpoint URL (if not using issuer auto-discovery)
- **`supports_pkce`**: Whether the provider supports PKCE (recommended: `true`)

## Common OIDC Providers

### Keycloak
```yaml
issuer: "https://your-keycloak.com/realms/your-realm"
```

### Auth0
```yaml
issuer: "https://your-domain.auth0.com/"
```

### Azure AD
```yaml
issuer: "https://login.microsoftonline.com/<tenant-id>/v2.0"
```

### Okta
```yaml
issuer: "https://your-domain.okta.com/oauth2/default"
```

## Redirect URL Format

For OIDC providers, the redirect URL should be:
```
https://your-oauth-frontend-domain:port/oidc/oauth
```
