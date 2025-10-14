---
sidebar_position: 10
---

# OAuth Public Frontend Configuration

zrok includes OAuth integration for public frontends, allowing you to authenticate users through various OAuth providers before they can access your shared resources. You can configure multiple OAuth providers and restrict access based on email address patterns.

## Planning for the OAuth Frontend

The OAuth public frontend uses an HTTP listener with a stable name to handle redirects from OAuth providers. You'll need to configure a DNS name and port for this listener that is accessible by your end users.

The OAuth frontend address will be used as the "redirect URL" when configuring OAuth clients with your providers. Each provider will redirect authenticated users back to this address, which then forwards them to their original destination.

## Configuring your Public Frontend

Add an `oauth` section to your frontend configuration:

```yaml
oauth:
  bind_address:               "192.168.1.100:443"
  endpoint_url:               "https://oauth.your-domain.com"
  cookie_name:                "zrok-auth-session"
  cookie_domain:              "your-domain.com"
  session_lifetime:           "6h"
  intermediate_lifetime:      "5m"
  signing_key:                "your-unique-signing-key"
  encryption_key:             "your-unique-encryption-key"

  providers:
    - name:                   "google"
      type:                   "google"
      client_id:              "<google-client-id>"
      client_secret:          "<google-client-secret>"
      
    - name:                   "github"
      type:                   "github"
      client_id:              "<github-client-id>"
      client_secret:          "<github-client-secret>"
      
    - name:                   "custom-oidc"
      type:                   "oidc"
      client_id:              "<oidc-client-id>"
      client_secret:          "<oidc-client-secret>"
      scopes:                 ["openid", "email", "profile"]
      issuer:                 "https://your-oidc-provider.com"
      supports_pkce:          true
```

### Configuration Parameters

All of the following parmeters _must_ be specified in the frontend configuration. There are no defaults.

- **`bind_address`**: IP and port where the OAuth frontend will listen (format: `ip:port`)
- **`endpoint_url`**: Public base URL where OAuth redirects will be handled
- **`cookie_name`**: Name for authentication cookies (suggested to use `zrok-auth-session`)
- **`cookie_domain`**: Domain where authentication cookies should be stored
- **`session_lifetime`**: How long authentication sessions remain valid (e.g., `6h`, `24h`)
- **`intermediate_lifetime`**: Lifetime for intermediate OAuth tokens (e.g., `5m`)
- **`signing_key`**: Unique 32+ character string for securing authentication payloads
- **`encryption_key`**: Unique 24+ character string for encrypting session data

### OAuth Providers

The `providers` array supports multiple OAuth configurations. Each provider requires:

- **`name`**: Unique identifier for this provider configuration; the `name` becomes part of the OAuth URLs for this provider, for example the callback URL becomes `/<name>/auth/callback`
- **`type`**: Provider type (`google`, `github`, or `oidc`)
- **`client_id`** and **`client_secret`**: OAuth client credentials

Providers may also require additional configuration values. For detailed setup instructions for each provider type, see:
- [Google OAuth Setup](integrations/google.md)
- [GitHub OAuth Setup](integrations/github.md)  
- [Generic OIDC Setup](integrations/oidc.md)

## OAuth Identity Flow

When a user accesses a zrok public share protected with OAuth, the following flow occurs:

```mermaid
sequenceDiagram
    participant User as User Browser
    participant Share as zrok Public Share
    participant OAuth as zrok OAuth Frontend
    participant Provider as OAuth Provider<br/>(Google/GitHub/OIDC)

    Note over User, Provider: OAuth Identity Flow for zrok Public Shares

    User->>Share: 1. Initial Access<br/>GET /share-url
    Share->>Share: 2. Authentication Check<br/>Validate session cookie
    
    alt No valid session
        Share->>User: 3. Redirect to Provider<br/>302 to OAuth provider login
        User->>Provider: 4. User Authentication<br/>Login with credentials
        Provider->>OAuth: 5. Provider Callback<br/>GET /<provider-name>/auth/callback?code=xyz
        OAuth->>Provider: 6. Token Exchange<br/>POST /token (exchange code for tokens)
        Provider->>OAuth: Return access token + user info
        OAuth->>OAuth: 7. Email Validation<br/>Check email against patterns
        
        alt Email validation passes
            OAuth->>OAuth: 8. Session Creation<br/>Create session + set cookie
            OAuth->>User: 9. Final Redirect<br/>302 back to original share URL
            User->>Share: 10. Access Granted<br/>GET /share-url (with valid session)
            Share->>User: Return protected content
        else Email validation fails
            OAuth->>User: Access Denied<br/>403 Forbidden
        end
    else Valid session exists
        Share->>User: Direct Access<br/>Return protected content
    end

    Note over User, Provider: Session remains valid for configured session_lifetime
```

### Flow Steps

1. **Initial Access**: User visits the zrok public share URL
2. **Authentication Check**: zrok checks for a valid authentication session cookie
3. **Redirect to Provider**: If no valid session exists, user is redirected to the configured OAuth provider's login page
4. **User Authentication**: User authenticates with their OAuth provider (Google, GitHub, etc.)
5. **Provider Callback**: OAuth provider redirects back to zrok's OAuth frontend at `/<provider-name>/auth/callback`
6. **Token Exchange**: zrok exchanges the authorization code for access tokens and retrieves user information
7. **Email Validation**: zrok validates the user's email address against any configured `--oauth-email-address-pattern` instances
8. **Session Creation**: If validation passes, zrok creates an authenticated session and sets a session cookie
9. **Final Redirect**: User is redirected back to the original zrok share URL
10. **Access Granted**: User can now access the protected resource

### Session Management

- **Maximum Session Duration**: Controlled by the `session_lifetime` configuration
- **Re-authentication**: Users must re-authenticate when sessions expire or when `--oauth-check-interval` is reached. Some providers (like the generic OIDC provider) support token refresh and will attempt to transparently refresh at this interval, rather than provoking the user to re-authenticate
- **Cross-Share Access**: Sessions are not shared between shares using the same provider; switching zrok shares will re-start the authentication flow for the specified provider 

## Using OAuth with Public Shares

Once your public frontend is configured with OAuth providers, you can enable authentication on public shares using these command line options:

- **`--oauth-provider <name>`**: Enable OAuth using the specified provider name from your configuration
- **`--oauth-email-address-pattern <pattern>`**: Restrict access to email addresses matching the glob pattern (use multiple times for multiple patterns)
- **`--oauth-check-interval <duration>`**: How often to re-verify authentication (default: 3h)

### Example

```bash
zrok share public --backend-mode web \
  --oauth-provider google \
  --oauth-email-address-pattern '*@example.com' \
  --oauth-email-address-pattern 'admin@*' \
  ~/public
```

This creates a public share that requires Google OAuth authentication and only allows users with `@example.com` email addresses or any `admin@*` email address.

## HTTP Headers for Proxied Requests

When zrok successfully authenticates a user via OAuth, it automatically adds authentication headers to all proxied requests sent to your backend application. These headers allow your application to identify the authenticated user and make authorization decisions.

### Authentication Headers

zrok sets the following HTTP headers on every proxied request after successful OAuth authentication:

- **`zrok-auth-provider`**: The name of the OAuth provider used for authentication (e.g., `google`, `github`, `custom-oidc`)
- **`zrok-auth-email`**: The authenticated user's email address as provided by the OAuth provider
- **`zrok-auth-expires`**: The timestamp when the authentication session will expire, formatted as RFC3339 (e.g., `2024-01-15T14:30:00Z`)

### Example Usage in Backend Applications

Your backend application can read these headers to implement user-specific logic:

#### Python/Flask Example
```python
from flask import Flask, request

app = Flask(__name__)

@app.route('/')
def index():
    provider = request.headers.get('zrok-auth-provider')
    email = request.headers.get('zrok-auth-email')
    expires = request.headers.get('zrok-auth-expires')
    
    return f"Welcome {email}! Authenticated via {provider}. Session expires: {expires}"
```

#### Go Example
```go
func handler(w http.ResponseWriter, r *http.Request) {
    provider := r.Header.Get("zrok-auth-provider")
    email := r.Header.Get("zrok-auth-email")
    expires := r.Header.Get("zrok-auth-expires")
    
    fmt.Fprintf(w, "Welcome %s! Authenticated via %s. Session expires: %s", 
                email, provider, expires)
}
```

#### Node.js/Express Example
```javascript
app.get('/', (req, res) => {
    const provider = req.headers['zrok-auth-provider'];
    const email = req.headers['zrok-auth-email'];
    const expires = req.headers['zrok-auth-expires'];
    
    res.send(`Welcome ${email}! Authenticated via ${provider}. Session expires: ${expires}`);
});
```

### Security Considerations

- **Trust Boundary**: These headers are only present when requests come through zrok's OAuth-protected frontend. Direct access to your backend would not include these headers.
- **Header Validation**: Your application should validate that these headers are present when OAuth protection is expected.
- **Session Expiration**: Use the `zrok-auth-expires` header to implement client-side session warnings or automatic logout.

## Logout Endpoint

Each configured OAuth provider automatically exposes a logout endpoint at `/<providerName>/logout`. This endpoint provides a secure way for users to terminate their authenticated sessions.

### Logout Process

When a user accesses the logout endpoint, zrok performs the following actions:

1. **Token Revocation**: The OAuth access token is revoked with the respective provider:
   - **Google**: Revokes the token via Google's OAuth2 revocation endpoint
   - **GitHub**: Deletes the application token using GitHub's API
   - **OIDC**: Uses the provider's token revocation endpoint (if supported)

2. **Session Clearing**: The local authentication session cookie is cleared by setting it to expire immediately

3. **Redirect**: The user is redirected to either:
   - A custom URL specified via the `redirect_url` query parameter
   - The provider's login page (default behavior)

#### Usage Examples

##### Basic Logout
```
GET https://oauth.your-domain.com/google/logout
```
This logs the user out and redirects them to the Google OAuth login page.

##### Logout with Custom Redirect
```
GET https://oauth.your-domain.com/github/logout?redirect_url=https://example.com/goodbye
```
This logs the user out and redirects them to `https://example.com/goodbye`.

#### Implementation Notes

- The logout endpoint validates that the session belongs to the correct provider before proceeding
- If token revocation fails with the OAuth provider, the logout process will still clear the local session
- The logout process is idempotent - calling it multiple times or without an active session will not cause errors
