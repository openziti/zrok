import {Command} from "commander";
import {createShare, deleteShare, loadRoot, PROXY_BACKEND_MODE, PUBLIC_SHARE_MODE, ShareRequest} from "@openziti/zrok";

const program = new Command();

program
    .command("share")
    .version("1.0.0")
    .description("sharing smoke test")
    .action(() => {
        let root = loadRoot();
        console.log("root.isEnabled", root.isEnabled());
        let req = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:8000");
        req.frontends = ["public"];
        createShare(root, req)
            .then(shr => {
                console.log(shr);
                deleteShare(root, shr);
            })
            .catch(ex => {
                console.log("exception", ex);
            });
    });

program.parse(process.argv);