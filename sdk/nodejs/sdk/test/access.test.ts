import {describe, it, expect, vi} from "vitest";
import {createAccess, deleteAccess} from "../src/zrok/access";
import {mockRoot, disabledRoot} from "./fixtures";
import {Access, AccessRequest, PROXY_BACKEND_MODE} from "../src/zrok/model";

vi.mock("../src/api", async (importOriginal) => {
    let original = await importOriginal() as any;
    return {
        ...original,
        ShareApi: vi.fn().mockImplementation(() => ({
            access: vi.fn().mockResolvedValue({
                frontendToken: "frontend-token-123",
                backendMode: "proxy",
            }),
            unaccess: vi.fn().mockResolvedValue(undefined),
        })),
        MetadataApi: vi.fn().mockImplementation(() => ({
            clientVersionCheck: vi.fn().mockResolvedValue(undefined),
        })),
    };
});

describe("createAccess", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        let req = new AccessRequest("share-token");
        await expect(createAccess(root, req)).rejects.toThrow("environment is not enabled");
    });

    it("should create access with share token", async () => {
        let root = mockRoot();
        let req = new AccessRequest("share-token");
        let acc = await createAccess(root, req);
        expect(acc.frontendToken).toBe("frontend-token-123");
        expect(acc.shareToken).toBe("share-token");
        expect(acc.backendMode).toBe("proxy");
    });

    it("should pass bind address", async () => {
        let root = mockRoot();
        let req = new AccessRequest("share-token", "127.0.0.1:9090");
        let acc = await createAccess(root, req);
        expect(acc.frontendToken).toBe("frontend-token-123");
    });
});

describe("deleteAccess", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        let acc = new Access("ft", "st", PROXY_BACKEND_MODE);
        await expect(deleteAccess(root, acc)).rejects.toThrow("environment is not enabled");
    });

    it("should call unaccess", async () => {
        let root = mockRoot();
        let acc = new Access("frontend-token", "share-token", PROXY_BACKEND_MODE);
        await expect(deleteAccess(root, acc)).resolves.not.toThrow();
    });
});
