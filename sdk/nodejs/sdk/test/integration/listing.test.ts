import {describe, it, expect, beforeAll, afterAll} from "vitest";
import {shouldSkip, getEndpoint, getAdminToken, createTestAccount, createEnabledRoot, cleanupRoot} from "./setup";
import {listShares, listAccesses} from "../../src/zrok/listing";
import {Root} from "../../src/zrok/environment";

describe.skipIf(shouldSkip())("listing integration", () => {
    let root: Root;

    beforeAll(async () => {
        let endpoint = getEndpoint()!;
        let adminToken = getAdminToken()!;
        let accountToken = await createTestAccount(endpoint, adminToken);
        root = await createEnabledRoot(endpoint, accountToken);
    });

    afterAll(async () => {
        if (root) {
            await cleanupRoot(root);
        }
    });

    it("should return empty shares list for fresh environment", async () => {
        let shares = await listShares(root);
        expect(shares).toEqual([]);
    });

    it("should return empty accesses list for fresh environment", async () => {
        let accesses = await listAccesses(root);
        expect(accesses).toEqual([]);
    });
});
