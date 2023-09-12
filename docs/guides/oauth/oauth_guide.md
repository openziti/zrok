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

In your oauth provider of choice's setup you would be prompted to create a client for accessing their services. This is where you will find the client_id and client_secret.

The port you choose is entirely up to the deployment. Just make sure it is open to receive callbacks from your configured oauth providers.

redirect_url is what we will tell the oauth providers to callback with the authorization result. This will be whatever domain you've chosen to host the access server against. This will get combined with the above port.

We then secure the response data within a zrok-access cookie. This is secured with the hash_key_raw. This can be any raw string.

### Required Scopes:
- google
- - Need access to a user's email: ./auth/userinfo.email 

### Example

An example config would look something like:
```yaml
oauth:
  port: 28080
  redirect_url: zrok.io
  hash_key_raw: "test1234test1234"
  providers:
    - name: google
      client_id: ohfwerouyr972t3riugdf89032r8y230ry.apps.googleusercontent.com
      client_secret: SDAFOHWER-qafsfgghrWERFfeqo13g 
```

Note that the client id and secret are jumbled text and do not correlate to actual secrets.

We spin up a zitadel oidc server on the specified port that handled all of the oauth handshaking. With the response we create a cookie with the name `zrok-access`.

## Enabling Oath on Share

To utilize the oauth integration on the access point we need to add a few more flags to our share command. There are three new flags:
- `provider` : This is the provider to authenticate against. Options are the same as above dependant on what the acess point is configured for
- `oauth-domains` : A list of valid email domains that are allowed to access the service. for example `gmail.com`
- `oauth-check-interval` : How long a `zrok-access` token is valid for before reinitializing the oauth flow. This is defaultly 3 hours.

That's all it takes!

Now when a user connects to your share they will be prompted with the chosen oauth provider and allowed based on your allowed domains. Simply restarting the service won't force a reauth for users either. Changing the `provider` or `oauth-check-interval` will, however. 