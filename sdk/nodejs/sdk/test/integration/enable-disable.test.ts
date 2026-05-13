import {describe, it, expect, beforeAll, afterAll} from "vitest";
import {shouldSkip, getEndpoint, getAdminToken, createTestAccount, createEnabledRoot, cleanupRoot} from "./setup";
import {enable, disable} from "../../src/zrok/enable";
import {status} from "../../src/zrok/status";
import {Root} from "../../src/zrok/environment";

describe.skipIf(shouldSkip())("enable/disable integration", () => {
    let root: Root;
    let endpoint: string;
    let accountToken: string;

    beforeAll(async () => {
        endpoint = getEndpoint()!;
        let adminToken = getAdminToken()!;
        accountToken = await createTestAccount(endpoint, adminToken);
        root = await createEnabledRoot(endpoint, accountToken);
    });

    afterAll(async () => {
        if (root) {
            await cleanupRoot(root);
        }
    });

    it("should have a valid environment after enable", () => {
        expect(root.isEnabled()).toBe(true);
        expect(root.environment?.accountToken).toBe(accountToken);
        expect(root.environment?.zId).toBeTruthy();
        expect(root.environment?.apiEndpoint).toBe(endpoint);
    });

    it("should report enabled status", () => {
        let s = status(root);
        expect(s.enabled).toBe(true);
        expect(s.token).toBe(accountToken);
        expect(s.zitiIdentity).toBeTruthy();
    });

    it("should be idempotent on second enable", async () => {
        let env = await enable(root, "different-token");
        expect(env.accountToken).toBe(accountToken); // returns existing, not new token
    });
});
