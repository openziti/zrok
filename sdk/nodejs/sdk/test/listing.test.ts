import {describe, it, expect, vi} from "vitest";
import {listShares, listAccesses} from "../src/zrok/listing";
import {mockRoot, disabledRoot} from "./fixtures";

vi.mock("../src/api", async (importOriginal) => {
    let original = await importOriginal() as any;
    return {
        ...original,
        MetadataApi: vi.fn().mockImplementation(() => ({
            clientVersionCheck: vi.fn().mockResolvedValue(undefined),
            listShares: vi.fn().mockResolvedValue({
                shares: [
                    {
                        shareToken: "share-1",
                        zId: "zid-1",
                        envZId: "envzid-1",
                        shareMode: "public",
                        backendMode: "proxy",
                        frontendEndpoints: ["https://share-1.zrok.io"],
                        target: "http://localhost:8080",
                        limited: false,
                        createdAt: 1000,
                        updatedAt: 2000,
                    },
                ],
            }),
            listAccesses: vi.fn().mockResolvedValue({
                accesses: [
                    {
                        id: 1,
                        frontendToken: "ft-1",
                        envZId: "envzid-1",
                        shareToken: "share-1",
                        backendMode: "proxy",
                        bindAddress: "127.0.0.1:8080",
                        description: "test access",
                        limited: false,
                        createdAt: 1000,
                        updatedAt: 2000,
                    },
                ],
            }),
        })),
    };
});

describe("listShares", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        await expect(listShares(root)).rejects.toThrow("environment is not enabled");
    });

    it("should map API response to ShareDetail[]", async () => {
        let root = mockRoot();
        let shares = await listShares(root);
        expect(shares).toHaveLength(1);
        expect(shares[0].token).toBe("share-1");
        expect(shares[0].zId).toBe("zid-1");
        expect(shares[0].shareMode).toBe("public");
        expect(shares[0].backendMode).toBe("proxy");
        expect(shares[0].frontendEndpoints).toEqual(["https://share-1.zrok.io"]);
        expect(shares[0].target).toBe("http://localhost:8080");
    });
});

describe("listAccesses", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        await expect(listAccesses(root)).rejects.toThrow("environment is not enabled");
    });

    it("should map API response to AccessDetail[]", async () => {
        let root = mockRoot();
        let accesses = await listAccesses(root);
        expect(accesses).toHaveLength(1);
        expect(accesses[0].id).toBe(1);
        expect(accesses[0].frontendToken).toBe("ft-1");
        expect(accesses[0].shareToken).toBe("share-1");
        expect(accesses[0].backendMode).toBe("proxy");
        expect(accesses[0].bindAddress).toBe("127.0.0.1:8080");
        expect(accesses[0].description).toBe("test access");
    });
});
