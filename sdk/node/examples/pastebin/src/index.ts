const { Command } = require("commander");
const zrok = require("zrok")
const ziti =  require('@openziti/ziti-sdk-nodejs')
const express = require('express')


const program = new Command();

program
  .command('copyto')
  .version("1.0.0")
  .description("command to host content to be pastedfrom'd")
  .action(async () => {
    let root = zrok.Load()
    await zrok.init(root);
    //await ziti.init( root.env.ZitiIdentity ).catch(( err: Error ) => { console.error(err); return process.exit(1) });
    ziti.setLogLevel(10)
    let shr = await zrok.CreateShare(root, new zrok.ShareRequest(zrok.TCP_TUNNEL_BACKEND_MODE, zrok.PUBLIC_SHARE_MODE, "pastebin", ["public"]));
    console.log("setting up app")
    let service = "ns5ix2brb61f"
    console.log("attempting to bind to service: "+ shr.Token)
    let app = ziti.express( express, shr.Token );
    console.log("after setting up app")
    app.get('/',function(_: Request,res: any){
      res.write("Test")
    });
    console.log("after setting up get")
    app.listen(undefined, () => {
      console.log(`Example app listening!`)
    })
    console.log("after listen")
    zrok.DeleteShare(root, shr);
  });

program
  .command('pastefrom <shrToken>')
  .version("1.0.0")
  .description("command to paste content from coptyo")
  .action(async (shrToken: string) => {
    
    //ziti.setLogLevel(10)
    console.log('pastefrom command called', shrToken);
    let root = zrok.Load();
    await zrok.init(root).catch((err: any) => {
      console.log(err)
    });
    let acc = await zrok.CreateAccess(root, new zrok.AccessRequest(shrToken))
    console.log("about to dial")
    zrok.dialer(root, shrToken, (data: any) => {
      console.log("in callback")
      console.log(data.toString())
    });
  });

program.parse(process.argv)
const options = program.opts();
