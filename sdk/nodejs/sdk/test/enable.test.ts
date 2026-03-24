import {describe, it, expect, vi, beforeEach} from "vitest";
import {enable, disable} from "../src/zrok/enable";
import {mockRoot, disabledRoot} from "./fixtures";
import {Environment} from "../src/zrok/environment";
import {Configuration} from "../src/api";

vi.mock("../src/api", async (importOriginal) => {
    let original = await importOriginal() as any;
    return {
        ...original,
        EnvironmentApi: vi.fn().mockImplementation(() => ({
            enable: vi.fn().mockResolvedValue({
                identity: "new-ziti-identity",
                cfg: '{"ztAPI":"https://ctrl.zrok.io"}',
            }),
            disable: vi.fn().mockResolvedValue(undefined),
        })),
        MetadataApi: vi.fn().mockImplementation(() => ({
            clientVersionCheck: vi.fn().mockResolvedValue(undefined),
        })),
    };
});

describe("enable", () => {
    it("should return existing environment when already enabled", async () => {
        let root = mockRoot({token: "existing-token", zId: "existing-zid"});
        let env = await enable(root, "new-token");
        expect(env.accountToken).toBe("existing-token");
        expect(env.zId).toBe("existing-zid");
    });

    it("should enable a disabled environment", async () => {
        let root = disabledRoot();
        // mock client() to return a valid configuration
        root.client = vi.fn().mockResolvedValue(
            new Configuration({basePath: "https://api-v2.zrok.io/api/v2"})
        );
        // mock setEnvironment and saveZitiIdentityNamed
        root.setEnvironment = vi.fn();
        root.saveZitiIdentityNamed = vi.fn();

        let env = await enable(root, "test-token", "test description", "test-host");
        expect(env.accountToken).toBe("test-token");
        expect(env.zId).toBe("new-ziti-identity");
        expect(root.setEnvironment).toHaveBeenCalled();
        expect(root.saveZitiIdentityNamed).toHaveBeenCalledWith(
            "environment",
            '{"ztAPI":"https://ctrl.zrok.io"}'
        );
    });

    it("should reset environment on client error", async () => {
        let root = disabledRoot();
        root.client = vi.fn().mockRejectedValue(new Error("connection failed"));

        await expect(enable(root, "test-token")).rejects.toThrow("error getting zrok client");
        expect(root.environment).toBeUndefined();
    });
});

describe("disable", () => {
    it("should be a no-op when not enabled", async () => {
        let root = disabledRoot();
        await expect(disable(root)).resolves.not.toThrow();
    });

    it("should disable an enabled environment", async () => {
        let root = mockRoot();
        root.deleteEnvironment = vi.fn();
        await disable(root);
        expect(root.deleteEnvironment).toHaveBeenCalled();
    });
});
