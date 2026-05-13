import {describe, it, expect, beforeAll, afterAll} from "vitest";
import {shouldSkip, getEndpoint, getAdminToken, createTestAccount, createEnabledRoot, cleanupRoot} from "./setup";
import {createName, deleteName, listNames, listNamespaces} from "../../src/zrok/name";
import {Root} from "../../src/zrok/environment";

describe.skipIf(shouldSkip())("name lifecycle integration", () => {
    let root: Root;
    let testName: string;

    beforeAll(async () => {
        let endpoint = getEndpoint()!;
        let adminToken = getAdminToken()!;
        let accountToken = await createTestAccount(endpoint, adminToken);
        root = await createEnabledRoot(endpoint, accountToken);
        testName = `test-name-${Date.now()}`;
    });

    afterAll(async () => {
        if (testName) {
            try {
                await deleteName(root, testName, "public");
            } catch {
                // ignore
            }
        }
        if (root) {
            await cleanupRoot(root);
        }
    });

    it("should list namespaces including public", async () => {
        let nss = await listNamespaces(root);
        let publicNs = nss.find(ns => ns.name === "public");
        expect(publicNs).toBeDefined();
    });

    it("should create a name in the public namespace", async () => {
        let entry = await createName(root, testName, "public");
        expect(entry.name).toBe(testName);
    });

    it("should list names and find the created name", async () => {
        let names = await listNames(root, "public");
        let found = names.find(n => n.name === testName);
        expect(found).toBeDefined();
    });

    it("should delete the name", async () => {
        await deleteName(root, testName, "public");
        let names = await listNames(root, "public");
        let found = names.find(n => n.name === testName);
        expect(found).toBeUndefined();
        testName = ""; // prevent afterAll double-delete
    });
});
