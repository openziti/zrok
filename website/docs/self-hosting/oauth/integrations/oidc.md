---
sidebar_position: 30
---

# Set up a generic OIDC provider

Configure an OpenID Connect (OIDC) provider as an authentication provider for your zrok public frontend. OIDC is supported by many identity providers including Keycloak, Auth0, Okta, Azure AD, and others.

## Provider requirements

Your OIDC provider must support:
- Authorization Code flow
- Discovery endpoint (optional but recommended)
- PKCE (Proof Key for Code Exchange)—optional but recommended for security

## Configure an OIDC provider

1. Create a new OAuth/OIDC client in your provider's admin interface
2. Set the **redirect URI** to `https://your-oauth-frontend-domain:port/<provider-name>/auth/callback`
3. Configure required scopes: `openid`, `email`, `profile`
4. Note the **client ID**, **client secret**, and **issuer URL**

## Add the OIDC provider to your frontend configuration

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
      prompt: "login"
      supports_pkce: true  # recommended for security
```

### Configuration options

- **`name`**: Unique identifier for this provider (used in share commands)
- **`type`**: Must be `"oidc"` for OpenID Connect providers
- **`client_id`** and **`client_secret`**: OAuth client credentials from your provider
- **`scopes`**: OAuth scopes to request (typically `["openid", "email", "profile"]`)
- **`issuer`**: The OIDC issuer URL (used for auto-discovery)
- **`discovery_url`**: Optional explicit discovery endpoint URL (if not using issuer auto-discovery)
- **`supports_pkce`**: Whether the provider supports PKCE (recommended: `true`)
- **`prompt`**: Optional prompt parameter for OIDC authentication (e.g., "login", "consent"), defaults to "login"

## Common OIDC providers

These are the `issuer` URLs for popular OIDC providers.

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

## Redirect URL format

For OIDC providers, the redirect URL should use your configured provider name:

```
https://your-oauth-frontend-domain:port/<provider-name>/auth/callback
```

For example, with the provider name `"my-oidc-provider"`:

```
https://your-oauth-frontend-domain:port/my-oidc-provider/auth/callback
```
