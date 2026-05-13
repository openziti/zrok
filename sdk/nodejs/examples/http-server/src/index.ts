import {Command} from "commander";
import {
    createShare,
    deleteShare,
    express,
    init,
    loadRoot,
    PROXY_BACKEND_MODE,
    PUBLIC_SHARE_MODE,
    setLogLevel,
    ShareRequest
} from "@openziti/zrok2";

const httpServer = async () => {
    const root = loadRoot();
    setLogLevel(0);
    await init(root)
        .catch((err: Error) => {
            console.log(err);
            return process.exit(1);
        });
    const shr = await createShare(root, new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http-server"));

    const app = express(shr);
    app.get("/", (_req: any, res: any) => {
        res.write("hello, world!\n");
        res.end();
    });
    app.listen(undefined, () => {
        console.log("listening at '" + shr.frontendEndpoints + "'");
    });

    process.on("SIGINT", async () => {
        await deleteShare(root, shr);
        process.exit(0);
    });
}

const program = new Command();
program.command("http-server").description("A simple HTTP server example, sharing directly to a zrok share").action(httpServer);
program.parse(process.argv);
