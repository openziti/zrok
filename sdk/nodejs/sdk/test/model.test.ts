import {describe, it, expect} from "vitest";
import {
    PROXY_BACKEND_MODE,
    WEB_BACKEND_MODE,
    TCP_TUNNEL_BACKEND_MODE,
    UDP_TUNNEL_BACKEND_MODE,
    CADDY_BACKEND_MODE,
    DRIVE_BACKEND_MODE,
    SOCKS_BACKEND_MODE,
    PRIVATE_SHARE_MODE,
    PUBLIC_SHARE_MODE,
    OPEN_PERMISSION_MODE,
    CLOSED_PERMISSION_MODE,
    AUTH_SCHEME_NONE,
    AUTH_SCHEME_BASIC,
    AUTH_SCHEME_OAUTH,
    ShareRequest,
    Share,
    AccessRequest,
    Access,
    EnableRequest,
    ShareDetail,
    AccessDetail,
    NameEntry,
    Namespace,
    Status,
    NameSelection,
} from "../src/zrok/model";

describe("backend mode constants", () => {
    it("should have all 7 backend modes", () => {
        expect(PROXY_BACKEND_MODE).toBe("proxy");
        expect(WEB_BACKEND_MODE).toBe("web");
        expect(TCP_TUNNEL_BACKEND_MODE).toBe("tcpTunnel");
        expect(UDP_TUNNEL_BACKEND_MODE).toBe("udpTunnel");
        expect(CADDY_BACKEND_MODE).toBe("caddy");
        expect(DRIVE_BACKEND_MODE).toBe("drive");
        expect(SOCKS_BACKEND_MODE).toBe("socks");
    });
});

describe("share mode constants", () => {
    it("should have private and public modes", () => {
        expect(PRIVATE_SHARE_MODE).toBe("private");
        expect(PUBLIC_SHARE_MODE).toBe("public");
    });
});

describe("permission mode constants", () => {
    it("should have open and closed modes", () => {
        expect(OPEN_PERMISSION_MODE).toBe("open");
        expect(CLOSED_PERMISSION_MODE).toBe("closed");
    });
});

describe("auth scheme constants", () => {
    it("should have none, basic, and oauth schemes", () => {
        expect(AUTH_SCHEME_NONE).toBe("none");
        expect(AUTH_SCHEME_BASIC).toBe("basic");
        expect(AUTH_SCHEME_OAUTH).toBe("oauth");
    });
});

describe("ShareRequest", () => {
    it("should have correct defaults for public share", () => {
        let req = new ShareRequest(PUBLIC_SHARE_MODE, PROXY_BACKEND_MODE, "http://localhost:8080");
        expect(req.reserved).toBe(false);
        expect(req.uniqueName).toBeUndefined();
        expect(req.shareMode).toBe(PUBLIC_SHARE_MODE);
        expect(req.backendMode).toBe(PROXY_BACKEND_MODE);
        expect(req.target).toBe("http://localhost:8080");
        expect(req.permissionMode).toBe(CLOSED_PERMISSION_MODE);
        expect(req.nameSelections).toHaveLength(1);
        expect(req.nameSelections![0].namespaceToken).toBe("public");
        expect(req.basicAuth).toBeUndefined();
        expect(req.oauthProvider).toBeUndefined();
        expect(req.accessGrants).toBeUndefined();
    });

    it("should have no name selections for private share", () => {
        let req = new ShareRequest(PRIVATE_SHARE_MODE, TCP_TUNNEL_BACKEND_MODE, "tcp://localhost:5432");
        expect(req.nameSelections).toBeUndefined();
    });
});

describe("Share", () => {
    it("should construct with token and endpoints", () => {
        let shr = new Share("abc123", ["https://abc123.share.zrok.io"]);
        expect(shr.shareToken).toBe("abc123");
        expect(shr.frontendEndpoints).toEqual(["https://abc123.share.zrok.io"]);
    });
});

describe("AccessRequest", () => {
    it("should construct with share token and optional bind address", () => {
        let req = new AccessRequest("share-token");
        expect(req.shareToken).toBe("share-token");
        expect(req.bindAddress).toBeUndefined();

        let reqWithBind = new AccessRequest("share-token", "127.0.0.1:8080");
        expect(reqWithBind.bindAddress).toBe("127.0.0.1:8080");
    });
});

describe("Access", () => {
    it("should construct with all fields", () => {
        let acc = new Access("frontend-token", "share-token", PROXY_BACKEND_MODE);
        expect(acc.frontendToken).toBe("frontend-token");
        expect(acc.shareToken).toBe("share-token");
        expect(acc.backendMode).toBe(PROXY_BACKEND_MODE);
    });
});

describe("EnableRequest", () => {
    it("should have correct defaults", () => {
        let req = new EnableRequest();
        expect(req.description).toBe("");
        expect(req.host).toBe("");
    });

    it("should accept custom values", () => {
        let req = new EnableRequest("my env", "my-host");
        expect(req.description).toBe("my env");
        expect(req.host).toBe("my-host");
    });
});

describe("ShareDetail", () => {
    it("should be constructable as an object literal", () => {
        const detail: ShareDetail = {
            token: "t", zId: "z", envZId: "e", shareMode: "public",
            backendMode: "proxy", frontendEndpoints: ["https://x.zrok.io"],
            target: "http://localhost", limited: false, createdAt: 1, updatedAt: 2,
        };
        expect(detail.token).toBe("t");
        expect(detail.frontendEndpoints).toEqual(["https://x.zrok.io"]);
    });
});

describe("AccessDetail", () => {
    it("should be constructable as an object literal", () => {
        const detail: AccessDetail = {
            id: 1, frontendToken: "ft", envZId: "e", shareToken: "st",
            backendMode: "proxy", bindAddress: "127.0.0.1:8080",
            description: "test", limited: false, createdAt: 1, updatedAt: 2,
        };
        expect(detail.id).toBe(1);
        expect(detail.bindAddress).toBe("127.0.0.1:8080");
    });
});

describe("NameEntry", () => {
    it("should be constructable as an object literal", () => {
        const entry: NameEntry = {
            namespaceToken: "public", namespaceName: "public",
            name: "my-name", shareToken: "st", reserved: true, createdAt: 1,
        };
        expect(entry.name).toBe("my-name");
        expect(entry.reserved).toBe(true);
    });
});

describe("Namespace", () => {
    it("should be constructable as an object literal", () => {
        const ns: Namespace = {namespaceToken: "public", name: "public", description: "the public ns"};
        expect(ns.namespaceToken).toBe("public");
        expect(ns.description).toBe("the public ns");
    });
});

describe("Status", () => {
    it("should be constructable as an object literal", () => {
        const s: Status = {
            enabled: true, apiEndpoint: "https://api.zrok.io",
            apiEndpointSource: "config", token: "tok", zitiIdentity: "zid",
        };
        expect(s.enabled).toBe(true);
        expect(s.token).toBe("tok");
    });
});

describe("NameSelection", () => {
    it("should construct with token and optional name", () => {
        let ns = new NameSelection("public");
        expect(ns.namespaceToken).toBe("public");
        expect(ns.name).toBeUndefined();

        let nsWithName = new NameSelection("public", "my-name");
        expect(nsWithName.name).toBe("my-name");
    });
});
