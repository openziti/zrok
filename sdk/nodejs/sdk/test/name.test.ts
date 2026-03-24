import {describe, it, expect, vi} from "vitest";
import {createName, deleteName, listNames, listNamespaces} from "../src/zrok/name";
import {mockRoot, disabledRoot} from "./fixtures";

vi.mock("../src/api", async (importOriginal) => {
    let original = await importOriginal() as any;
    return {
        ...original,
        ShareApi: vi.fn().mockImplementation(() => ({
            createShareName: vi.fn().mockResolvedValue(undefined),
            deleteShareName: vi.fn().mockResolvedValue(undefined),
            listAllNames: vi.fn().mockResolvedValue([
                {
                    namespaceToken: "public",
                    namespaceName: "public",
                    name: "my-name",
                    shareToken: "share-123",
                    reserved: true,
                    createdAt: 1000,
                },
            ]),
            listNamesForNamespace: vi.fn().mockResolvedValue([
                {
                    namespaceToken: "ns-1",
                    namespaceName: "my-ns",
                    name: "ns-name",
                    shareToken: "share-456",
                    reserved: false,
                    createdAt: 2000,
                },
            ]),
            listShareNamespaces: vi.fn().mockResolvedValue([
                {
                    namespaceToken: "public",
                    name: "public",
                    description: "the public namespace",
                },
                {
                    namespaceToken: "ns-1",
                    name: "my-ns",
                    description: "a custom namespace",
                },
            ]),
        })),
        MetadataApi: vi.fn().mockImplementation(() => ({
            clientVersionCheck: vi.fn().mockResolvedValue(undefined),
        })),
    };
});

describe("createName", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        await expect(createName(root, "my-name")).rejects.toThrow("environment is not enabled");
    });

    it("should create a name and return NameEntry", async () => {
        let root = mockRoot();
        let entry = await createName(root, "my-name", "public");
        expect(entry.name).toBe("my-name");
        expect(entry.namespaceToken).toBe("public");
    });
});

describe("deleteName", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        await expect(deleteName(root, "my-name")).rejects.toThrow("environment is not enabled");
    });

    it("should delete a name without error", async () => {
        let root = mockRoot();
        await expect(deleteName(root, "my-name", "public")).resolves.not.toThrow();
    });
});

describe("listNames", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        await expect(listNames(root)).rejects.toThrow("environment is not enabled");
    });

    it("should list all names without namespace", async () => {
        let root = mockRoot();
        let names = await listNames(root);
        expect(names).toHaveLength(1);
        expect(names[0].name).toBe("my-name");
        expect(names[0].namespaceToken).toBe("public");
        expect(names[0].shareToken).toBe("share-123");
        expect(names[0].reserved).toBe(true);
    });

    it("should list names for specific namespace", async () => {
        let root = mockRoot();
        let names = await listNames(root, "ns-1");
        expect(names).toHaveLength(1);
        expect(names[0].name).toBe("ns-name");
        expect(names[0].namespaceToken).toBe("ns-1");
    });
});

describe("listNamespaces", () => {
    it("should throw when root is not enabled", async () => {
        let root = disabledRoot();
        await expect(listNamespaces(root)).rejects.toThrow("environment is not enabled");
    });

    it("should list namespaces", async () => {
        let root = mockRoot();
        let nss = await listNamespaces(root);
        expect(nss).toHaveLength(2);
        expect(nss[0].namespaceToken).toBe("public");
        expect(nss[0].name).toBe("public");
        expect(nss[1].namespaceToken).toBe("ns-1");
        expect(nss[1].description).toBe("a custom namespace");
    });
});
