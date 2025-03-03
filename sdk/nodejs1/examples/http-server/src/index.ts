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

const program = new Command();

program
    .command("http-server")
    .version("1.0.0")
    .description("A simple HTTP server example, sharing directly to a zrok share")
    .action(async () => {
        let root = loadRoot();
        await init(root)
            .catch((err: Error) => {
                console.log(err);
                return process.exit(1);
            });
        let req = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http-server");
        req.frontends = ["public"];
        let shr = await createShare(root, req);

        let app = express(shr);
        app.get("/", (r: Request, res: any) => {
            res.write("hello, world!");
            res.end();
        });
        app.listen(undefined, () => {
            console.log("listening at '" + shr.frontendEndpoints + "'");
        });

        process.on("SIGINT", async () => {
            deleteShare(root, shr);
        });
    });

program.parse(process.argv);