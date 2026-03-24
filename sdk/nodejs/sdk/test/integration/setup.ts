import {Root, Metadata, Environment, Config} from "../../src/zrok/environment";
import {enable, disable} from "../../src/zrok/enable";
import {AdminApi, Configuration} from "../../src/api";
import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";

export const getEndpoint = (): string | undefined => {
    return process.env.ZROK2_API_ENDPOINT;
}

export const getAdminToken = (): string | undefined => {
    return process.env.ZROK2_ADMIN_TOKEN;
}

export const shouldSkip = (): boolean => {
    return !getEndpoint() || !getAdminToken();
}

export const createTestAccount = async (endpoint: string, adminToken: string): Promise<string> => {
    let cfg = new Configuration({
        basePath: endpoint + "/api/v2",
        apiKey: adminToken,
    });
    let adminApi = new AdminApi(cfg);
    let res = await adminApi.createAccount({
        body: {
            email: `test-${Date.now()}@integration.zrok.io`,
            password: `test-password-${Date.now()}`,
        },
    });
    return res.accountToken!;
}

export const createEnabledRoot = async (endpoint: string, accountToken: string): Promise<Root> => {
    let tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "zrok2-integ-"));

    let metadata = new Metadata("v0.4", tmpDir);
    let config = new Config(endpoint);
    let root = new Root(metadata, undefined, config);

    // write metadata file so rootExists() returns true
    let metaDir = tmpDir;
    fs.mkdirSync(metaDir, {recursive: true});
    fs.writeFileSync(path.join(metaDir, "metadata.json"), JSON.stringify({v: "v0.4"}));

    await enable(root, accountToken, "integration test", os.hostname());

    return root;
}

export const cleanupRoot = async (root: Root): Promise<void> => {
    try {
        await disable(root);
    } catch {
        // ignore disable errors
    }
    let rootPath = root.metadata.rootPath;
    if (rootPath && fs.existsSync(rootPath)) {
        fs.rmSync(rootPath, {recursive: true, force: true});
    }
}
