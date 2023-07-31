# "pastebin" SDK Example

This `pastebin` example is a minimal `zrok` SDK application that implements a wormhole that makes redirecting file contents between multiple `zrok` environments very easy.

The `pastebin` example is split into two separate commands. The `copyto` command takes a copy buffer from standard input. You can use it like this:

```
$ echo "this is a pastebin test" | copyto
access your pastebin using 'pastefrom b46p9j82z81f'
```

And then using another terminal window, you can access your pastebin data like this:

```
$ pastefrom b46p9j82z81f
this is a pastebin test
```

## The `copyto` Implementation

The `copyto` utility is an illustration of how to implement an application that creates a share and exposes it to the `zrok` network. Let's look at each section of the implementation:

```go
	data, err := loadData()
	if err != nil {
		panic(err)
	}
```

This first block of code is responsible for calling the `loadData` function, which loads the pastebin with data from `os.Stdin`.

All SDK applications need to load the user's "root" from the `environment` package, like this:

```go
	root, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}
```

The `root` is a structure that contains all of the user's environment detail and allows the SDK application to access the `zrok` service instance and the underlying OpenZiti network.

Next, `copyto` will create a `zrok` share:

```go
	shr, err := sdk.CreateShare(root, &sdk.ShareRequest{
		BackendMode: sdk.TcpTunnelBackendMode,
		ShareMode:   sdk.PrivateShareMode,
		Target:      "pastebin",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("access your pastebin using 'pastefrom %v'\n", shr.Token)
```

The `sdk.CreateShare` call uses the loaded `environment` root along with the details of the share request (`sdk.ShareRequest`) to create the share that will be used to access the `pastebin`.

For the `pastebin` application, we're using a `sdk.TcpTunnelBackendMode` backend mode (we're just using a single network connection that implements a reliable byte stream, so TCP works great). Tunnel backends only work with `private` shares as of `zrok` `v0.4`, so we're using  `sdk.PrivateShareMode`. 

We'll set the `Target` to be `pastebin`, as that's just metadata describing the application.

Finally, we emit the share token so the user can access the `pastebin` using the `pastefrom` command.

Next, we'll use the SDK to create a listener for this share:

```go
	listener, err := sdk.NewListener(shr.Token, root)
	if err != nil {
		panic(err)
	}
```

The `sdk.NewListener` establishes a network listener for the newly created share. This listener works just like a `net.Listener`.

Next, we're going to add a shutdown hook so that `copyto` will delete the share when the application is terminated using `^C`:

```go
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := sdk.DeleteShare(root, shr); err != nil {
			panic(err)
		}
		_ = listener.Close()
		os.Exit(0)
	}()
```

This anonymous function runs waiting for a signal to exit. When that is received, it runs the `sdk.DeleteShare` function to remove the share that was created. This is how ephemeral shares work for the `zrok share` commands as well.

And finally, we run in an infinite loop waiting for requests for the `pastebin` data from the network:

```go
	for {
		if conn, err := listener.Accept(); err == nil {
			go handle(conn, data)
		} else {
			panic(err)
		}
	}	
```

## The "pastefrom" Implementation

The `pastefrom` application works very similarly to `copyto`. The primary difference is that it "dials" the share through the SDK using `sdk.NewDialer`, which returns a `net.Conn`:

```go
	conn, err := sdk.NewDialer(shrToken, root)
	if err != nil {
		panic(err)
	}
```

When this `sdk.NewDialer` function returns without an error, a bidirectional `net.Conn` has been established between the `copyto` "server" and the `pastefrom` "client". `pastefrom` then just reads the available data from the `net.Conn` and emits it to `os.Stdout`.