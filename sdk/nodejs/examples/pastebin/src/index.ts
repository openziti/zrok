import {Command} from "commander";
// @ts-ignore
import {
    AccessRequest,
    createAccess,
    createShare,
    deleteAccess,
    deleteShare,
    dialer,
    init,
    listener,
    loadRoot,
    PRIVATE_SHARE_MODE,
    ShareRequest,
    TCP_TUNNEL_BACKEND_MODE,
    write
} from "@openziti/zrok";
import readlineSync = require('readline-sync');

const copyto = async () => {
    let text = readlineSync.question("enter some text: ");
    let root = loadRoot();
    await init(root)
        .catch((err: Error) => {
            console.log(err);
            return process.exit(1);
        });
    let shr = await createShare(root, new ShareRequest(PRIVATE_SHARE_MODE, TCP_TUNNEL_BACKEND_MODE, "copyto"));

    console.log("connect with 'pastefrom " + shr.shareToken + "'");

    listener(shr, (data: any) => {
        write(data.client, new TextEncoder().encode(text + "\n"));
    });

    process.on("SIGINT", async () => {
        deleteShare(root, shr);
    });
}

const pastefrom = async (shareToken: string) => {
    let root = loadRoot();
    await init(root).catch((err: Error) => {
        console.log(err);
        return process.exit(1);
    });
    let acc = await createAccess(root, new AccessRequest(shareToken));

    dialer(acc, () => {}, async (data: any) => {
        console.log(new TextDecoder().decode(data));
        await deleteAccess(root, acc);
        process.exit(0);
    });
}

const program = new Command();
program.command("copyto").description("serve a copy buffer").action(copyto);
program.command("pastefrom <shareToken>").description("receive a copy buffer").action(pastefrom);
program.parse(process.argv);