import {Command} from "commander";
import {
    createShare,
    deleteShare,
    express,
    init,
    loadRoot,
    PROXY_BACKEND_MODE,
    PUBLIC_SHARE_MODE,
    ShareRequest
} from "@openziti/zrok";

const httpServer = async () => {
    let root = loadRoot();
    await init(root)
        .catch((err: Error) => {
            console.log(err);
            return process.exit(1);
        });
    let shr = await createShare(root, new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http-server"));

    let app = express(shr);
    app.get("/", (r: Request, res: any) => {
        res.write("hello, world!\n");
        res.end();
    });
    app.listen(undefined, () => {
        console.log("listening at '" + shr.frontendEndpoints + "'");
    });

    process.on("SIGINT", async () => {
        deleteShare(root, shr);
    });
}

const program = new Command();
program.command("http-server").description("A simple HTTP server example, sharing directly to a zrok share").action(httpServer);
program.parse(process.argv);