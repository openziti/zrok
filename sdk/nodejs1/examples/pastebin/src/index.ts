import {Command} from "commander";
// @ts-ignore
import {
    createShare,
    deleteShare,
    init,
    listener,
    loadRoot,
    PRIVATE_SHARE_MODE, setLogLevel,
    ShareRequest,
    TCP_TUNNEL_BACKEND_MODE,
    write
} from "@openziti/zrok";
import readlineSync = require('readline-sync');

const program = new Command();

const copyto = async () => {
    let text = readlineSync.question("enter some text: ");
    let root = loadRoot();
    await init(root)
        .catch((err: Error) => {
            console.log(err);
            return process.exit(1);
        });
    let shr = await createShare(root, new ShareRequest(PRIVATE_SHARE_MODE, TCP_TUNNEL_BACKEND_MODE, "copyto"));
    console.log("connect with 'copyto " + shr.shareToken + "'");
    listener(shr, {clientConnectCallback: (data: any) => {
        console.log("data", data);
        write(data.client, text) ;
    }});

    process.on("SIGINT", async () => {
        console.log("cleaning up");
        deleteShare(root, shr);
    });
}

program.command("copyto").version("1.0.0").description("serve a copy buffer").action(copyto);
program.parse(process.argv);