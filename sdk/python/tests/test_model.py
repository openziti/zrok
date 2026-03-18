"""Tests for zrok.model — dataclass construction and constants."""

import zrok2.model as model


class TestConstants:
    def test_backend_modes(self):
        assert model.PROXY_BACKEND_MODE == "proxy"
        assert model.TCP_TUNNEL_BACKEND_MODE == "tcpTunnel"
        assert model.UDP_TUNNEL_BACKEND_MODE == "udpTunnel"
        assert model.WEB_BACKEND_MODE == "web"
        assert model.CADDY_BACKEND_MODE == "caddy"
        assert model.DRIVE_BACKEND_MODE == "drive"
        assert model.SOCKS_BACKEND_MODE == "socks"

    def test_share_modes(self):
        assert model.PRIVATE_SHARE_MODE == "private"
        assert model.PUBLIC_SHARE_MODE == "public"

    def test_permission_modes(self):
        assert model.OPEN_PERMISSION_MODE == "open"
        assert model.CLOSED_PERMISSION_MODE == "closed"

    def test_auth_schemes(self):
        assert model.AUTH_SCHEME_NONE == "none"
        assert model.AUTH_SCHEME_BASIC == "basic"
        assert model.AUTH_SCHEME_OAUTH == "oauth"


class TestShareRequest:
    def test_defaults(self):
        req = model.ShareRequest(
            BackendMode=model.PROXY_BACKEND_MODE,
            ShareMode=model.PUBLIC_SHARE_MODE,
            Target="http://localhost:8080",
        )
        assert req.Frontends == []
        assert req.BasicAuth == []
        assert req.OauthProvider == ""
        assert req.Reserved is False
        assert req.PermissionMode == model.OPEN_PERMISSION_MODE
        assert req.AccessGrants == []
        assert req.NameSelections == []

    def test_with_name_selections(self):
        ns = model.NameSelection(NamespaceToken="ns1", Name="myname")
        req = model.ShareRequest(
            BackendMode=model.TCP_TUNNEL_BACKEND_MODE,
            ShareMode=model.PRIVATE_SHARE_MODE,
            Target="tcp://localhost:25565",
            NameSelections=[ns],
        )
        assert len(req.NameSelections) == 1
        assert req.NameSelections[0].Name == "myname"


class TestShareDetail:
    def test_defaults(self):
        detail = model.ShareDetail()
        assert detail.Token == ""
        assert detail.Limited is False
        assert detail.FrontendEndpoints == []


class TestAccessDetail:
    def test_defaults(self):
        detail = model.AccessDetail()
        assert detail.Id == 0
        assert detail.FrontendToken == ""


class TestNameEntry:
    def test_construction(self):
        entry = model.NameEntry(
            NamespaceToken="ns1",
            NamespaceName="public",
            Name="myname",
            ShareToken="shr_abc",
            Reserved=True,
            CreatedAt=1000,
        )
        assert entry.Name == "myname"
        assert entry.Reserved is True


class TestNamespace:
    def test_construction(self):
        ns = model.Namespace(NamespaceToken="ns1", Name="public", Description="desc")
        assert ns.NamespaceToken == "ns1"


class TestStatus:
    def test_defaults(self):
        s = model.Status()
        assert s.Enabled is False
        assert s.ApiEndpoint == ""
        assert s.Token == ""
