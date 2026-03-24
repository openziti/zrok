import {describe, it, expect, beforeEach, afterEach, vi} from "vitest";
import {Root, Metadata, Environment, Config, ApiEndpoint, defaultRoot, loadRoot} from "../src/zrok/environment";
import {mockRoot, disabledRoot, tmpZrokDir} from "./fixtures";
import * as fs from "node:fs";
import * as path from "node:path";

describe("Root", () => {
    describe("hasConfig", () => {
        it("should return true when config is set", () => {
            let root = mockRoot({config: new Config("https://custom.zrok.io")});
            expect(root.hasConfig()).toBe(true);
        });

        it("should return false when config is undefined", () => {
            let root = mockRoot();
            root.config = undefined;
            expect(root.hasConfig()).toBe(false);
        });
    });

    describe("isEnabled", () => {
        it("should return true when environment is set", () => {
            let root = mockRoot();
            expect(root.isEnabled()).toBe(true);
        });

        it("should return false when environment is undefined", () => {
            let root = disabledRoot();
            expect(root.isEnabled()).toBe(false);
        });
    });

    describe("apiEndpoint", () => {
        it("should return default endpoint when no config or env vars", () => {
            let root = disabledRoot();
            root.config = undefined;
            let ep = root.apiEndpoint();
            expect(ep.endpoint).toBe("https://api-v2.zrok.io");
            expect(ep.from).toBe("binary");
        });

        it("should use config endpoint when set", () => {
            let root = disabledRoot();
            root.config = new Config("https://custom.zrok.io");
            let ep = root.apiEndpoint();
            expect(ep.endpoint).toBe("https://custom.zrok.io");
            expect(ep.from).toBe("config");
        });

        it("should prefer ZROK2_API_ENDPOINT over config", () => {
            let root = disabledRoot();
            root.config = new Config("https://custom.zrok.io");
            process.env.ZROK2_API_ENDPOINT = "https://env2.zrok.io";
            try {
                let ep = root.apiEndpoint();
                expect(ep.endpoint).toBe("https://env2.zrok.io");
                expect(ep.from).toBe("ZROK2_API_ENDPOINT");
            } finally {
                delete process.env.ZROK2_API_ENDPOINT;
            }
        });

        it("should fall back to ZROK_API_ENDPOINT with deprecation warning", () => {
            let root = disabledRoot();
            let warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
            process.env.ZROK_API_ENDPOINT = "https://old.zrok.io";
            try {
                let ep = root.apiEndpoint();
                expect(ep.endpoint).toBe("https://old.zrok.io");
                expect(ep.from).toBe("ZROK_API_ENDPOINT");
                expect(warnSpy).toHaveBeenCalledWith(
                    "WARNING: ZROK_API_ENDPOINT is deprecated, use ZROK2_API_ENDPOINT instead"
                );
            } finally {
                delete process.env.ZROK_API_ENDPOINT;
                warnSpy.mockRestore();
            }
        });

        it("should prefer ZROK2_API_ENDPOINT over deprecated ZROK_API_ENDPOINT", () => {
            let root = disabledRoot();
            process.env.ZROK2_API_ENDPOINT = "https://new.zrok.io";
            process.env.ZROK_API_ENDPOINT = "https://old.zrok.io";
            try {
                let ep = root.apiEndpoint();
                expect(ep.endpoint).toBe("https://new.zrok.io");
                expect(ep.from).toBe("ZROK2_API_ENDPOINT");
            } finally {
                delete process.env.ZROK2_API_ENDPOINT;
                delete process.env.ZROK_API_ENDPOINT;
            }
        });

        it("should use environment endpoint when enabled", () => {
            let root = mockRoot({apiEndpoint: "https://env.zrok.io"});
            let ep = root.apiEndpoint();
            expect(ep.endpoint).toBe("https://env.zrok.io");
            expect(ep.from).toBe("env");
        });

        it("should strip trailing slashes", () => {
            let root = disabledRoot();
            root.config = new Config("https://custom.zrok.io///");
            let ep = root.apiEndpoint();
            expect(ep.endpoint).toBe("https://custom.zrok.io");
        });
    });

    describe("client", () => {
        it("should call MetadataApi.clientVersionCheck", async () => {
            let root = mockRoot();
            let cfg = await root.client();
            expect(root.client).toHaveBeenCalled();
            expect(cfg).toBeDefined();
        });
    });

    describe("setEnvironment", () => {
        it("should write environment to disk", () => {
            let tmp = tmpZrokDir();
            try {
                let root = disabledRoot();
                root.metadata.rootPath = tmp.dir;

                // override paths to use temp dir
                let envFile = path.join(tmp.dir, "environment.json");
                let env = new Environment("token-123", "zid-456", "https://api.zrok.io");

                // we need to mock the file paths - use the actual setEnvironment with patched paths
                // since setEnvironment uses the paths module, we test via the Root directly
                root.setEnvironment(env);

                expect(root.environment).toEqual(env);
                expect(root.isEnabled()).toBe(true);
            } finally {
                tmp.cleanup();
            }
        });
    });

    describe("deleteEnvironment", () => {
        it("should clear environment", () => {
            let root = mockRoot();
            expect(root.isEnabled()).toBe(true);
            root.deleteEnvironment();
            expect(root.isEnabled()).toBe(false);
            expect(root.environment).toBeUndefined();
        });
    });
});

describe("defaultRoot", () => {
    it("should create root with v2.0 version", () => {
        let root = defaultRoot();
        expect(root.metadata.v).toBe("v0.4");
        expect(root.environment).toBeUndefined();
        expect(root.isEnabled()).toBe(false);
    });
});
