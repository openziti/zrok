const { Command } = require("commander");
const zrok = require("zrok")
const ziti =  require('@openziti/ziti-sdk-nodejs')
var readlineSync = require('readline-sync');


const program = new Command();

program
  .command('copyto')
  .version("1.0.0")
  .description("command to host content to be pastedfrom'd")
  .action(async () => {
    var data = readlineSync.question('Input some text... ');

    console.log("data is: ", data)
    let root = zrok.Load()
    await zrok.init( root ).catch(( err: Error ) => { console.error(err); return process.exit(1) });
    zrok.setLogLevel(0)
    console.log("setting up zrok.CreateShare...")
    let shr = await zrok.CreateShare(root, new zrok.ShareRequest(zrok.TCP_TUNNEL_BACKEND_MODE, zrok.PRIVATE_SHARE_MODE, "pastebin", ["private"]));
    console.log("access your pastebin using 'pastefrom ", shr.Token)
    let app = zrok.express( shr.Token );
    app.get('/',function(_: Request,res: any){
      res.write(data)
      res.end()
    });
    app.listen(undefined, () => {
    })

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
    
    let root = zrok.Load();
    await zrok.init(root).catch((err: any) => {
      console.log(err)
    });
    let acc = await zrok.CreateAccess(root, new zrok.AccessRequest(shrToken))

    ziti.httpRequest( 
      shrToken, 
      undefined, 
      'GET', 
      '/', 
      [],
      (data: any) => {  // on_req_cb
        console.log("in on_req_cb")
        console.log("data is: ", data)
      },
      (data: any) => {  // on_resp_cb
        console.log("in on_resp_cb")
        console.log("data is: ", data)
      },
      async (data: any) => {  // on_resp_data_cb
        console.log("in on_resp_data_cb")
        console.log("data is: ", data)
        if (data.body) {
          console.log('----------- pastefrom is: ', data.body.toString());

          await zrok.DeleteAccess(root, acc)

          process.exit(0);
        }

      }
    );
  });

program.parse(process.argv)
const options = program.opts();
