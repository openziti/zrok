import {describe, it, expect, beforeAll, afterAll} from "vitest";
import {shouldSkip, getEndpoint, getAdminToken, createTestAccount, createEnabledRoot, cleanupRoot} from "./setup";
import {createShare, deleteShare} from "../../src/zrok/share";
import {createAccess, deleteAccess} from "../../src/zrok/access";
import {listAccesses} from "../../src/zrok/listing";
import {Root} from "../../src/zrok/environment";
import {
    ShareRequest,
    Share,
    Access,
    AccessRequest,
    PRIVATE_SHARE_MODE,
    TCP_TUNNEL_BACKEND_MODE,
} from "../../src/zrok/model";

describe.skipIf(shouldSkip())("access lifecycle integration", () => {
    let root: Root;
    let shr: Share;
    let acc: Access;

    beforeAll(async () => {
        let endpoint = getEndpoint()!;
        let adminToken = getAdminToken()!;
        let accountToken = await createTestAccount(endpoint, adminToken);
        root = await createEnabledRoot(endpoint, accountToken);

        let req = new ShareRequest(PRIVATE_SHARE_MODE, TCP_TUNNEL_BACKEND_MODE, "tcp://localhost:5432");
        shr = await createShare(root, req);
    });

    afterAll(async () => {
        if (acc) {
            try {
                await deleteAccess(root, acc);
            } catch {
                // ignore
            }
        }
        if (shr) {
            try {
                await deleteShare(root, shr);
            } catch {
                // ignore
            }
        }
        if (root) {
            await cleanupRoot(root);
        }
    });

    it("should create access to the share", async () => {
        let req = new AccessRequest(shr.shareToken);
        acc = await createAccess(root, req);
        expect(acc.frontendToken).toBeTruthy();
        expect(acc.shareToken).toBe(shr.shareToken);
    });

    it("should appear in listAccesses", async () => {
        let accesses = await listAccesses(root);
        let found = accesses.find(a => a.frontendToken === acc.frontendToken);
        expect(found).toBeDefined();
    });

    it("should delete the access", async () => {
        await deleteAccess(root, acc);
        let accesses = await listAccesses(root);
        let found = accesses.find(a => a.frontendToken === acc.frontendToken);
        expect(found).toBeUndefined();
        acc = undefined as any;
    });
});
