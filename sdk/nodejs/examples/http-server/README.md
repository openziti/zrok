# "http-server" SDK Example

This `http-server` example is a minimal `zrok` application that surfaces a basic http server over a private zrok share.

## Implementation

```js
    let root = zrok.Load()
```

The `root` is a structure that contains all of the user's environment detail and allows the SDK application to access the `zrok` service instance and the underlying OpenZiti network.

```js
    let shr = await zrok.CreateShare(root, new zrok.ShareRequest(zrok.PROXY_BACKEND_MODE, zrok.PRIVATE_SHARE_MODE, "http-server", ["private"]));
    console.log(`access your private HTTP Server on another machine using:  'zrok access private ${shr.Token}'`)
```

The `zrok.CreateShare` call uses the loaded `environment` root along with the details of the share request (`zrok.ShareRequest`) to create the share that will be used to access the `http-server`.

We are using the `zrok.PROXY_BACKEND_MODE` to handle tcp traffic. This time we are using `zrok.PRIVATE_SHARE_MODE` to take advantage of a private share that is running. With that we set which frontends to listen on, so we use whatever is configured, `private` here.

Further down we emit where to access the service.

Then we create a NodeJS Express web server that is configured to listen for ONLY incoming Ziti connections.  No TCP connections from the open internet are possible.

```js
    let app = zrok.express( shr.Token );

    ...

    app.listen(undefined, () => {
      console.log(`private HTTP Server is now listening for incoming requests`)
    })

    ...

    app.get('/', function(_: Request, res: any){
		...
	}
```

Then we create a signal handler that catches CTRL-C.  When that signal is caught, we tear down teh zrok share, and exit the process.

```js
    process.on('SIGINT', async () => { 
      console.log("Now deleting your private zrok share...")
      await zrok.DeleteShare(root, shr)
      process.exit(15);
    });
```

## How to execute this example

You should first set up and enable your local `zrok` environment. Then execute the following cmds:

```cmd
npm install
node dist/index.js http-server
```

When the example app begins execution, you should see output on the console that resembles the following:

```
Now creating your private zrok share...
access your private HTTP Server on another machine using:  'zrok access private <SOME_TOKEN>'
private HTTP Server is now listening for incoming requests
```

Later, when you are done using the http-server, you can press CTRL_C, and the app will shut down, like this:

```
Now deleting your private zrok share...
```
