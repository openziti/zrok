# "http-server" SDK Example

This `http-server` example is a minimal `zrok` application that surfaces a basic http server over a public zrok share.

## Implementation

```go
	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}
```

The `root` is a structure that contains all of the user's environment detail and allows the SDK application to access the `zrok` service instance and the underlying OpenZiti network.

```go
    shr, err := sdk.CreateShare(root, &sdk.ShareRequest{
		BackendMode: sdk.TcpTunnelBackendMode,
		ShareMode:   sdk.PublicShareMode,
		Frontends:   []string{"public"},
		Target:      "http-server",
	})

	if err != nil {
		panic(err)
	}
	defer func() {
		if err := sdk.DeleteShare(root, shr); err != nil {
			panic(err)
		}
	}()
    ...
	fmt.Println("Access server at the following endpoints: ", strings.Join(shr.FrontendEndpoints, "\n"))

```

The `sdk.CreateShare` call uses the loaded `environment` root along with the details of the share request (`sdk.ShareRequest`) to create the share that will be used to access the `http-server`.

We are using the `sdk.TcpTunnelBackendMode` to handle tcp traffic. This time we are using `sdk.PublicShareMode` to take advantage of a public share that is running. With that we set which frontends to listen on, so we use whatever is configured, `public` here.

Further down we emit where to access the service.

Then we create a listener and use that to server our http server:

```go
conn, err := sdk.NewListener(shr.Token, root)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

    ...

	http.HandleFunc("/", helloZrok)

	if err := http.Serve(conn, nil); err != nil {
		panic(err)
	}
```
