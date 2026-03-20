import {vi} from "vitest";
import {Root, Metadata, Environment, Config} from "../src/zrok/environment";
import {Configuration} from "../src/api";
import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";

export const mockRoot = (opts?: {
    token?: string;
    zId?: string;
    apiEndpoint?: string;
    config?: Config;
}): Root => {
    let env = new Environment(
        opts?.token || "test-token",
        opts?.zId || "test-zid",
        opts?.apiEndpoint || "https://test.zrok.io",
    );
    let metadata = new Metadata("v0.4", "/tmp/test-zrok2");
    let root = new Root(metadata, env, opts?.config);

    // mock client() to skip the version check
    root.client = vi.fn().mockResolvedValue(
        new Configuration({
            basePath: env.apiEndpoint + "/api/v2",
            apiKey: env.accountToken,
        })
    );

    return root;
}

export const disabledRoot = (): Root => {
    let metadata = new Metadata("v0.4", "/tmp/test-zrok2");
    let root = new Root(metadata, undefined);
    return root;
}

export const tmpZrokDir = (): {dir: string; cleanup: () => void} => {
    let dir = fs.mkdtempSync(path.join(os.tmpdir(), "zrok2-test-"));
    return {
        dir,
        cleanup: () => {
            fs.rmSync(dir, {recursive: true, force: true});
        },
    };
}
