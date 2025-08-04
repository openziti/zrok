---
sidebar_position: 20
---

# GitHub OAuth Setup

This guide covers setting up GitHub OAuth for your zrok public frontend.

## Register OAuth Application

Navigate to your GitHub account settings: **Settings > Developer Settings > OAuth Apps > Register a new application**

![](../images/github_create_oauth_application_1.png)

![](../images/github_create_oauth_application_2.png)

Configure the **Authorization callback URL** to match your OAuth frontend address with `/<provider-name>/auth/callback` appended:

![](../images/github_create_oauth_application_3.png)

Create a new client secret:

![](../images/github_create_oauth_application_4.png)

Save the client ID and client secret for your frontend configuration.

## Frontend Configuration

Add the GitHub provider to your `frontend.yml`:

```yaml
oauth:
  providers:
    - name: "github"
      type: "github"
      client_id: "<your-github-client-id>"
      client_secret: "<your-github-client-secret>"
```

## Redirect URL Format

For GitHub OAuth with the provider name `"github"`, the redirect URL should be:
```
https://your-oauth-frontend-domain:port/github/auth/callback
```

If you use a different provider name (e.g., `"gh-enterprise"`), the URL would be:
```
https://your-oauth-frontend-domain:port/gh-enterprise/auth/callback
```
