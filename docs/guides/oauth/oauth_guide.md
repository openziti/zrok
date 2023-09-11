## Enabling Oauth on Access Point

There is a new stanza in the access point configuration. 

```yaml
oauth:
  port: <host-port> #port to listen on oauth callbacks from
  redirect_url: <host-url> #redirect url to feed into oauth flow
  hash_key_raw: "<your-key>" #key we will use to sign our access token
  providers: #which providers we configure to use.
    - name: <provider-name>
      client_id: <client-id> #the client id you get from your oauth provider
      client_secret: <client-secret> #the client secret you get from your oauth provider
```
Currently we support the following Oauth providers:
- google
- github

We spin up a zitadel oidc server on the specified port that handled all of the oauth handshaking. With the response we create a cookie with the name `zrok-access`.

## Enabling Oath on Share

To utilize the oauth integration on the access point we need to add a few more flags to our share command. There are three new flags:
- `provider` : This is the provider to authenticate against. Options are the same as above dependant on what the acess point is configured for
- `oauth-domains` : A list of valid email domains that are allowed to access the service. for example `gmail.com`
- `oauth-check-interval` : How long a `zrok-access` token is valid for before reinitializing the oauth flow. This is defaultly 3 hours.

That's all it takes!

Now when a user connects to your share they will be prompted with the chosen oauth provider and allowed based on your allowed domains. Simply restarting the service won't force a reauth for users either. Changing the `provider` or `oauth-check-interval` will, however. 