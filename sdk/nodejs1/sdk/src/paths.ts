import { homedir } from "os";
import { join } from "node:path";

export const rootDir = (): string => {
    return join(homedir(), ".zrok");
}

