import {describe, it, expect, beforeAll, afterAll} from "vitest";
import {shouldSkip, getEndpoint, getAdminToken, createTestAccount, createEnabledRoot, cleanupRoot} from "./setup";
import {createShare, deleteShare, getShareDetail} from "../../src/zrok/share";
import {listShares} from "../../src/zrok/listing";
import {Root} from "../../src/zrok/environment";
import {ShareRequest, Share, PRIVATE_SHARE_MODE, TCP_TUNNEL_BACKEND_MODE} from "../../src/zrok/model";

describe.skipIf(shouldSkip())("share lifecycle integration", () => {
    let root: Root;
    let shr: Share;

    beforeAll(async () => {
        let endpoint = getEndpoint()!;
        let adminToken = getAdminToken()!;
        let accountToken = await createTestAccount(endpoint, adminToken);
        root = await createEnabledRoot(endpoint, accountToken);
    });

    afterAll(async () => {
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

    it("should create a private tcpTunnel share", async () => {
        let req = new ShareRequest(PRIVATE_SHARE_MODE, TCP_TUNNEL_BACKEND_MODE, "tcp://localhost:5432");
        shr = await createShare(root, req);
        expect(shr.shareToken).toBeTruthy();
    });

    it("should appear in listShares", async () => {
        let shares = await listShares(root);
        let found = shares.find(s => s.token === shr.shareToken);
        expect(found).toBeDefined();
    });

    it("should return detail for the share", async () => {
        let detail = await getShareDetail(root, shr.shareToken);
        expect(detail.token).toBe(shr.shareToken);
        expect(detail.backendMode).toBe(TCP_TUNNEL_BACKEND_MODE);
    });

    it("should delete the share", async () => {
        await deleteShare(root, shr);
        let shares = await listShares(root);
        let found = shares.find(s => s.token === shr.shareToken);
        expect(found).toBeUndefined();
        shr = undefined as any; // prevent afterAll double-delete
    });
});
