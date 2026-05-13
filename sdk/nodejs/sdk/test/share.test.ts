import {describe, it, expect, vi, beforeEach} from "vitest";
import {createShare, deleteShare, releaseReservedShare, modifyShare, getShareDetail, toApiBackendMode} from "../src/zrok/share";
import {mockRoot, disabledRoot} from "./fixtures";
import {
    ShareRequest,
    Share,
    PRIVATE_SHARE_MODE,
    PUBLIC_SHARE_MODE,
    PROXY_BACKEND_MODE,
    WEB_BACKEND_MODE,
    TCP_TUNNEL_BACKEND_MODE,
    UDP_TUNNEL_BACKEND_MODE,
    CADDY_BACKEND_MODE,
    DRIVE_BACKEND_MODE,
    SOCKS_BACKEND_MODE,
    AUTH_SCHEME_BASIC,
    AUTH_SCHEME_OAUTH,
    NameSelection,
} from "../src/zrok/model";
import {ShareRequestBackendModeEnum} from "../src/api";

// mock the API modules
vi.mock("../src/api", async (importOriginal) => {
    let original = await importOriginal() as any;
    return {
        ...original,
        ShareApi: vi.fn().mockImplementation(() => ({
            share: vi.fn().mockResolvedValue({
                shareToken: "test-share-token",
                frontendProxyEndpoints: ["https://test.share.zrok.io"],
            }),
            unshare: vi.fn().mockResolvedValue(undefined),
            updateShare: vi.fn().mockResolvedValue(undefined),
        })),
        MetadataApi: vi.fn().mockImplementation(() => ({
            clientVersionCheck: vi.fn().mockResolvedValue(undefined),
            getShareDetail: vi.fn().mockResolvedValue({
                shareToken: "detail-token",
                zId: "detail-zid",
                envZId: "detail-envzid",
                shareMode: "public",
                backendMode: "proxy",
                frontendEndpoints: ["https://detail.share.zrok.io"],
                target: "http://localhost:8080",
                limited: false,
                createdAt: 1000,
                updatedAt: 2000,
            }),
        })),
    };
});

describe("createShare", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        let req = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:8080");
        await expect(createShare(root, req)).rejects.toThrow("environment is not enabled");
    });

    it("should create a public share", async () => {
        let root = mockRoot();
        let req = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:8080");
        let shr = await createShare(root, req);
        expect(shr.shareToken).toBe("test-share-token");
        expect(shr.frontendEndpoints).toEqual(["https://test.share.zrok.io"]);
    });

    it("should create a private share", async () => {
        let root = mockRoot();
        let req = new ShareRequest(PRIVATE_SHARE_MODE, TCP_TUNNEL_BACKEND_MODE, "tcp://localhost:5432");
        req.accessGrants = ["user@example.com"];
        let shr = await createShare(root, req);
        expect(shr.shareToken).toBe("test-share-token");
    });

    it("should throw on unknown share mode", async () => {
        let root = mockRoot();
        let req = new ShareRequest("invalid", PROXY_BACKEND_MODE, "http://localhost:8080");
        await expect(createShare(root, req)).rejects.toThrow("unknown share mode");
    });
});

describe("deleteShare", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        let shr = new Share("token", []);
        await expect(deleteShare(root, shr)).rejects.toThrow("environment is not enabled");
    });

    it("should call unshare", async () => {
        let root = mockRoot();
        let shr = new Share("token-to-delete", ["https://endpoint.zrok.io"]);
        await expect(deleteShare(root, shr)).resolves.not.toThrow();
    });
});

describe("releaseReservedShare", () => {
    it("should call unshare", async () => {
        let root = mockRoot();
        let shr = new Share("reserved-token", []);
        await expect(releaseReservedShare(root, shr)).resolves.not.toThrow();
    });
});

describe("modifyShare", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        await expect(modifyShare(root, "token")).rejects.toThrow("environment is not enabled");
    });

    it("should call updateShare with grants", async () => {
        let root = mockRoot();
        await expect(
            modifyShare(root, "token", ["user@add.com"], ["user@remove.com"])
        ).resolves.not.toThrow();
    });
});

describe("getShareDetail", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        await expect(getShareDetail(root, "token")).rejects.toThrow("environment is not enabled");
    });

    it("should return mapped ShareDetail", async () => {
        let root = mockRoot();
        let detail = await getShareDetail(root, "detail-token");
        expect(detail.token).toBe("detail-token");
        expect(detail.zId).toBe("detail-zid");
        expect(detail.envZId).toBe("detail-envzid");
        expect(detail.shareMode).toBe("public");
        expect(detail.backendMode).toBe("proxy");
        expect(detail.frontendEndpoints).toEqual(["https://detail.share.zrok.io"]);
        expect(detail.target).toBe("http://localhost:8080");
        expect(detail.limited).toBe(false);
        expect(detail.createdAt).toBe(1000);
        expect(detail.updatedAt).toBe(2000);
    });
});

describe("toApiBackendMode", () => {
    it("should map all 7 backend modes", () => {
        expect(toApiBackendMode(PROXY_BACKEND_MODE)).toBe(ShareRequestBackendModeEnum.Proxy);
        expect(toApiBackendMode(WEB_BACKEND_MODE)).toBe(ShareRequestBackendModeEnum.Web);
        expect(toApiBackendMode(TCP_TUNNEL_BACKEND_MODE)).toBe(ShareRequestBackendModeEnum.TcpTunnel);
        expect(toApiBackendMode(UDP_TUNNEL_BACKEND_MODE)).toBe(ShareRequestBackendModeEnum.UdpTunnel);
        expect(toApiBackendMode(CADDY_BACKEND_MODE)).toBe(ShareRequestBackendModeEnum.Caddy);
        expect(toApiBackendMode(DRIVE_BACKEND_MODE)).toBe(ShareRequestBackendModeEnum.Drive);
        expect(toApiBackendMode(SOCKS_BACKEND_MODE)).toBe(ShareRequestBackendModeEnum.Socks);
    });

    it("should return undefined for unknown mode", () => {
        expect(toApiBackendMode("unknown")).toBeUndefined();
    });
});
