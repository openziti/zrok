# "http-server" SDK Example

This `http-server` example is a minimal zrok application that surfaces a basic HTTP server over a public share.

## Implementation

```python
root = zrok.environment.root.Load()
```

The `root` is a structure that contains all of the user's environment details and allows the SDK application to access the zrok service instance and the underlying OpenZiti network.

```python
try:
    shr = zrok.share.CreateShare(root=root, request=ShareRequest(
        BackendMode=zrok.model.TCP_TUNNEL_BACKEND_MODE,
        ShareMode=zrok.model.PUBLIC_SHARE_MODE,
        Frontends=['public'],
        Target="http-server"
    ))
    shrToken = shr.Token
    print("Access server at the following endpoints: ", "\n".join(shr.FrontendEndpoints))

    def removeShare():
        zrok.share.DeleteShare(root=root, shr=shr)
        print("Deleted share")
    atexit.register(removeShare)
except Exception as e:
    print("unable to create share", e)
    sys.exit(1)
```

The `sdk.CreateShare` call uses the loaded `environment` root along with the details of the share request (`sdk.ShareRequest`) to create the share that will be used to access the `http-server`.

We are using the `sdk.TcpTunnelBackendMode` to handle tcp traffic. This time we are using `sdk.PublicShareMode` to take advantage of a public share that is running. With that we set which frontends to listen on, so we use whatever is configured, `public` here.

Next, we populate our `cfg` options for our decorator.

```python
zrok_opts['cfg'] = zrok.decor.Opts(root=root, shrToken=shrToken, bindPort=bindPort)
```

Next, we run the server which ends up calling the following:

```python
@zrok.decor.zrok(opts=zrok_opts)
def runApp():
    from waitress import serve
    # the port is only used to integrate zrok with frameworks that expect a "hostname:port" combo
    serve(app, port=bindPort)
```
