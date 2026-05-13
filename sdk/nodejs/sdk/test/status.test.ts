import {describe, it, expect} from "vitest";
import {status} from "../src/zrok/status";
import {mockRoot, disabledRoot} from "./fixtures";

describe("status", () => {
    it("should return enabled status for enabled root", () => {
        let root = mockRoot({token: "my-token", zId: "my-zid", apiEndpoint: "https://api.zrok.io"});
        let s = status(root);
        expect(s.enabled).toBe(true);
        expect(s.apiEndpoint).toBe("https://api.zrok.io");
        expect(s.apiEndpointSource).toBe("env");
        expect(s.token).toBe("my-token");
        expect(s.zitiIdentity).toBe("my-zid");
    });

    it("should return disabled status for disabled root", () => {
        let root = disabledRoot();
        let s = status(root);
        expect(s.enabled).toBe(false);
        expect(s.token).toBe("");
        expect(s.zitiIdentity).toBe("");
    });

    it("should return default endpoint for empty root", () => {
        let root = disabledRoot();
        root.config = undefined;
        let s = status(root);
        expect(s.apiEndpoint).toBe("https://api-v2.zrok.io");
        expect(s.apiEndpointSource).toBe("binary");
    });
});
