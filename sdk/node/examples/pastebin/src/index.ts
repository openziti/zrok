const { Command } = require("commander"); // add this line
const zrok = require("zrok")
//const environ = require("zrok/environment")

const program = new Command();

program
  .command('copyto')
  .version("1.0.0")
  .description("command to host content to be pastedfrom'd")
  .action(() => {
    console.log('copyto command called');
    //console.log(environ)
    let root = zrok.Load()
    let shr = zrok.CreateShare(root, new zrok.ShareRequest(zrok.TCP_TUNNEL_BACKEND_MODE, zrok.PUBLIC_SHARE_MODE, "pastebin"));
    console.log(shr)
    zrok.DeleteShare(root, shr);
  });

program
  .command('pastefrom <shrToken>')
  .version("1.0.0")
  .description("command to paste content from coptyo")
  .action((shrToken: string) => {
    console.log('pastefrom command called', shrToken);
  });

program.parse(process.argv)
const options = program.opts();
