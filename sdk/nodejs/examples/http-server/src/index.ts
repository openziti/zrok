const { Command } = require("commander");
const zrok = require("zrok")


const program = new Command();

program
  .command('http-server')
  .version("1.0.0")
  .description("command to host an HTTP Server")
  .action(async () => {

    // Load the zrok env
    let root = zrok.Load()

    // Authenticate with the Ziti network
    await zrok.init( root ).catch(( err: Error ) => { console.error(err); return process.exit(1) });

    // Set this to a larger value (e.g. 10) to trace lower-level Ziti SDK activity
    zrok.setLogLevel(0)

    // Create the zrok private share that will represent this web server
    console.log("Now creating your private zrok share...")
    let shr = await zrok.CreateShare(root, new zrok.ShareRequest(zrok.PROXY_BACKEND_MODE, zrok.PRIVATE_SHARE_MODE, "http-server", ["private"]));
    console.log(`access your private HTTP Server on another machine using:  'zrok access private ${shr.Token}'`)

    // Create a NodeJS Express web server that listens NOT on a TCP port, but for incoming Ziti connections to the private zrok share
    let app = zrok.express( shr.Token );

    // Set up a simple route
    let reqCtr = 0;
    app.get('/', function(_: Request, res: any){
      reqCtr++;
      console.log(`received a GET request... reqCtr[${reqCtr}]`);
      res.write(`Hello zrok! reqCtr[${reqCtr}]`)
      res.end()
    });

    // Start listening for incoming requests
    app.listen(undefined, () => {
      console.log(`private HTTP Server is now listening for incoming requests`)
    })

    // Delete the private share upon CTRL-C
    process.on('SIGINT', async () => { 
      console.log("Now deleting your private zrok share...")
      await zrok.DeleteShare(root, shr)
      process.exit(15);
    });

  });


program.parse(process.argv)
const options = program.opts();
