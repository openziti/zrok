const { Command } = require("commander");
const zrok = require("@openziti/zrok")
var readlineSync = require('readline-sync');


const program = new Command();

program
  .command('copyto')
  .version("1.0.0")
  .description("command to host content to be pastedfrom'd")
  .action(async () => {
    var data = readlineSync.question('Input some text... ');

    console.log("data is: ", data)

    var enc = new TextEncoder();
    var buf = enc.encode(data); //

    zrok.setLogLevel(0)

    let root = zrok.Load()
    await zrok.init( root ).catch(( err: Error ) => { console.error(err); return process.exit(1) });
    console.log("setting up zrok.CreateShare...")
    let shr = await zrok.CreateShare(root, new zrok.ShareRequest(zrok.TCP_TUNNEL_BACKEND_MODE, zrok.PRIVATE_SHARE_MODE, "pastebin", []));
    console.log(`access your pastebin using 'pastefrom ${shr.Token}'`)

    zrok.listener(
      shr.Token,
      (data: any) => {      // listenCallback
      },
      (data: any) => {      // listenClientCallback
      },
      (data: any) => {      // clientConnectCallback
        // when we receive a client connection, then write the data to them
        zrok.write(
          data.client,
          buf,
          (data: any) => {  // writeCallback
          }
        );
      },
      (data: any) => {      // clientDataCallback
      },
    );

    // Delete the private share upon CTRL-C
    process.on('SIGINT', async () => { 
      console.log("Now deleting your private zrok share...")
      await zrok.DeleteShare(root, shr)
      process.exit(15);
    });
    
  });

program
  .command('pastefrom <shrToken>')
  .version("1.0.0")
  .description("command to paste content from coptyo")
  .action(async (shrToken: string) => {

    zrok.setLogLevel(0)

    let root = zrok.Load();
    await zrok.init(root).catch((err: any) => {
      console.log(err)
    });
    let acc = await zrok.CreateAccess(root, new zrok.AccessRequest(shrToken))

    var dec = new TextDecoder("utf-8");

    zrok.dialer(
      root, 
      shrToken, 
      (data: any) => {  // on_connect_cb
      },
      (data: any) => {  // on_data_cb
        console.log("data is: ", dec.decode(data));
        process.exit(0);
      },
    );
  
    process.on('SIGINT', async () => { 
      process.exit(15);
    });

  });

program.parse(process.argv)
const options = program.opts();
