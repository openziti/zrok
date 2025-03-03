import {Command} from "commander";
import {loadRoot} from "@openziti/zrok/dist/environment";


const program = new Command();

program
    .command("share")
    .version("1.0.0")
    .description("sharing smoke test")
    .action(() => {
        let root = loadRoot();
        console.log("root.isEnabled", root.isEnabled());
    });

program.parse(process.argv);