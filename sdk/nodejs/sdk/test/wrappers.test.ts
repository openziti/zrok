import {describe, it, expect, vi, beforeEach} from "vitest";
import {mockRoot} from "./fixtures";

const mockCreateShare = vi.fn();
const mockDeleteShare = vi.fn();
const mockReleaseReservedShare = vi.fn();
const mockCreateAccess = vi.fn();
const mockDeleteAccess = vi.fn();
const mockGetOverview = vi.fn();

vi.mock("../src/zrok/share", () => ({
    createShare: (...args: any[]) => mockCreateShare(...args),
    deleteShare: (...args: any[]) => mockDeleteShare(...args),
    releaseReservedShare: (...args: any[]) => mockReleaseReservedShare(...args),
    toApiBackendMode: vi.fn(),
}));

vi.mock("../src/zrok/access", () => ({
    createAccess: (...args: any[]) => mockCreateAccess(...args),
    deleteAccess: (...args: any[]) => mockDeleteAccess(...args),
}));

vi.mock("../src/zrok/overview", () => ({
    getOverview: (...args: any[]) => mockGetOverview(...args),
}));

import {withShare, withAccess, ProxyShare} from "../src/zrok/wrappers";
import {
    ShareRequest,
    Share,
    AccessRequest,
    Access,
    PUBLIC_SHARE_MODE,
    PROXY_BACKEND_MODE,
} from "../src/zrok/model";

describe("withShare", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        mockCreateShare.mockResolvedValue(new Share("test-token", ["https://test.zrok.io"]));
        mockDeleteShare.mockResolvedValue(undefined);
    });

    it("should create share, invoke callback, and delete share", async () => {
        let root = mockRoot();
        let req = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:8080");
        let result = await withShare(root, req, async (shr) => {
            expect(shr.shareToken).toBe("test-token");
            return "done";
        });
        expect(result).toBe("done");
        expect(mockCreateShare).toHaveBeenCalled();
        expect(mockDeleteShare).toHaveBeenCalled();
    });

    it("should skip delete when reserved is true", async () => {
        let root = mockRoot();
        let req = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:8080");
        req.reserved = true;
        await withShare(root, req, async () => "ok");
        expect(mockCreateShare).toHaveBeenCalled();
        expect(mockDeleteShare).not.toHaveBeenCalled();
    });

    it("should delete share even when callback throws", async () => {
        let root = mockRoot();
        let req = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:8080");
        await expect(
            withShare(root, req, async () => {
                throw new Error("callback error");
            })
        ).rejects.toThrow("callback error");
        expect(mockDeleteShare).toHaveBeenCalled();
    });
});

describe("withAccess", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        mockCreateAccess.mockResolvedValue(new Access("ft-123", "st-123", "proxy"));
        mockDeleteAccess.mockResolvedValue(undefined);
    });

    it("should create access, invoke callback, and delete access", async () => {
        let root = mockRoot();
        let req = new AccessRequest("share-token");
        let result = await withAccess(root, req, async (acc) => {
            expect(acc.frontendToken).toBe("ft-123");
            return "accessed";
        });
        expect(result).toBe("accessed");
        expect(mockCreateAccess).toHaveBeenCalled();
        expect(mockDeleteAccess).toHaveBeenCalled();
    });

    it("should delete access even when callback throws", async () => {
        let root = mockRoot();
        let req = new AccessRequest("share-token");
        await expect(
            withAccess(root, req, async () => {
                throw new Error("access error");
            })
        ).rejects.toThrow("access error");
        expect(mockDeleteAccess).toHaveBeenCalled();
    });
});

describe("ProxyShare", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        mockCreateShare.mockResolvedValue(new Share("proxy-token", ["https://proxy.zrok.io"]));
        mockGetOverview.mockResolvedValue({environments: []});
        mockReleaseReservedShare.mockResolvedValue(undefined);
    });

    it("should create a reserved share", async () => {
        let root = mockRoot();
        let proxy = await ProxyShare.create(root, "http://localhost:3000");
        expect(proxy.token).toBe("proxy-token");
        expect(proxy.endpoints).toEqual(["https://proxy.zrok.io"]);
        expect(mockCreateShare).toHaveBeenCalled();
    });

    it("should register cleanup when no unique name", async () => {
        let root = mockRoot();
        let proxy = await ProxyShare.create(root, "http://localhost:3000");
        // cleanup should be registered
        expect(proxy.token).toBe("proxy-token");
    });
});
