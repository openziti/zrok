"""Tests for zrok.share — CreateShare, DeleteShare, ModifyShare, GetShareDetail."""

import pytest
from unittest.mock import patch, MagicMock

from zrok.share import CreateShare, DeleteShare, ModifyShare, GetShareDetail, ReleaseReservedShare
import zrok.model as model


def _make_mock_api_share_request():
    """Return a MagicMock that stands in for the pydantic ShareRequest from zrok_api."""
    m = MagicMock()
    m.auth_scheme = model.AUTH_SCHEME_NONE
    m.auth_users = []
    return m


class TestCreateShare:
    def test_not_enabled_raises(self, disabled_root):
        req = model.ShareRequest(
            BackendMode=model.PROXY_BACKEND_MODE,
            ShareMode=model.PUBLIC_SHARE_MODE,
            Target="http://localhost:8080",
        )
        with pytest.raises(Exception, match="not enabled"):
            CreateShare(disabled_root, req)

    def test_create_public_share(self, mock_root):
        mock_share_api = MagicMock()
        mock_response = MagicMock()
        mock_response.data = MagicMock()
        mock_response.data.share_token = "shr_abc123"
        mock_response.data.frontend_proxy_endpoints = ["https://abc123.share.zrok.io"]
        mock_share_api.share_with_http_info.return_value = mock_response

        req = model.ShareRequest(
            BackendMode=model.PROXY_BACKEND_MODE,
            ShareMode=model.PUBLIC_SHARE_MODE,
            Target="http://localhost:8080",
            Frontends=["public"],
        )

        with patch("zrok.share.ShareApi", return_value=mock_share_api), \
             patch("zrok.share._ShareRequest__newPublicShare", return_value=_make_mock_api_share_request(), create=True) as _, \
             patch("zrok.share.ShareRequest", return_value=_make_mock_api_share_request()):
            share = CreateShare(mock_root, req)
            assert share.Token == "shr_abc123"
            assert share.FrontendEndpoints == ["https://abc123.share.zrok.io"]

    def test_create_private_share(self, mock_root):
        mock_share_api = MagicMock()
        mock_response = MagicMock()
        mock_response.data = MagicMock()
        mock_response.data.share_token = "shr_priv"
        mock_response.data.frontend_proxy_endpoints = []
        mock_share_api.share_with_http_info.return_value = mock_response

        req = model.ShareRequest(
            BackendMode=model.TCP_TUNNEL_BACKEND_MODE,
            ShareMode=model.PRIVATE_SHARE_MODE,
            Target="tcp://localhost:25565",
        )

        with patch("zrok.share.ShareApi", return_value=mock_share_api), \
             patch("zrok.share.ShareRequest", return_value=_make_mock_api_share_request()):
            share = CreateShare(mock_root, req)
            assert share.Token == "shr_priv"

    def test_create_share_unknown_mode_raises(self, mock_root):
        req = model.ShareRequest(
            BackendMode=model.PROXY_BACKEND_MODE,
            ShareMode="invalid",
            Target="http://localhost:8080",
        )
        with pytest.raises(Exception, match="unknown share mode"):
            CreateShare(mock_root, req)

    def test_create_share_with_basic_auth(self, mock_root):
        mock_share_api = MagicMock()
        mock_response = MagicMock()
        mock_response.data = MagicMock()
        mock_response.data.share_token = "shr_auth"
        mock_response.data.frontend_proxy_endpoints = []
        mock_share_api.share_with_http_info.return_value = mock_response

        req = model.ShareRequest(
            BackendMode=model.PROXY_BACKEND_MODE,
            ShareMode=model.PUBLIC_SHARE_MODE,
            Target="http://localhost:8080",
            Frontends=["public"],
            BasicAuth=["user1:pass1"],
        )

        with patch("zrok.share.ShareApi", return_value=mock_share_api), \
             patch("zrok.share.ShareRequest", return_value=_make_mock_api_share_request()):
            share = CreateShare(mock_root, req)
            assert share.Token == "shr_auth"


class TestDeleteShare:
    def test_delete_share(self, mock_root):
        mock_share_api = MagicMock()
        shr = model.Share(Token="shr_abc", FrontendEndpoints=[])

        with patch("zrok.share.ShareApi", return_value=mock_share_api):
            DeleteShare(mock_root, shr)
            mock_share_api.unshare_with_http_info.assert_called_once()

    def test_delete_share_raises_on_error(self, mock_root):
        mock_share_api = MagicMock()
        mock_share_api.unshare_with_http_info.side_effect = Exception("404")
        shr = model.Share(Token="shr_bad", FrontendEndpoints=[])

        with patch("zrok.share.ShareApi", return_value=mock_share_api):
            with pytest.raises(Exception, match="error deleting share"):
                DeleteShare(mock_root, shr)


class TestReleaseReservedShare:
    def test_release_reserved_share(self, mock_root):
        mock_share_api = MagicMock()
        shr = model.Share(Token="shr_reserved", FrontendEndpoints=[])

        with patch("zrok.share.ShareApi", return_value=mock_share_api):
            ReleaseReservedShare(mock_root, shr)
            mock_share_api.unshare_with_http_info.assert_called_once()
            call_args = mock_share_api.unshare_with_http_info.call_args
            body = call_args.kwargs.get("body") or call_args[1].get("body")
            # UnshareRequest doesn't have 'reserved' in pydantic model;
            # verify the share_token was passed correctly
            assert body.share_token == "shr_reserved"


class TestModifyShare:
    def test_not_enabled_raises(self, disabled_root):
        with pytest.raises(Exception, match="not enabled"):
            ModifyShare(disabled_root, "shr_x")

    def test_modify_share_adds_grants(self, mock_root):
        mock_share_api = MagicMock()

        with patch("zrok.share.ShareApi", return_value=mock_share_api):
            ModifyShare(mock_root, "shr_abc", add_access_grants=["grant1"])
            mock_share_api.update_share_with_http_info.assert_called_once()

    def test_modify_share_api_error(self, mock_root):
        mock_share_api = MagicMock()
        mock_share_api.update_share_with_http_info.side_effect = Exception("500")

        with patch("zrok.share.ShareApi", return_value=mock_share_api):
            with pytest.raises(Exception, match="error modifying share"):
                ModifyShare(mock_root, "shr_abc", add_access_grants=["grant1"])


class TestGetShareDetail:
    def test_not_enabled_raises(self, disabled_root):
        with pytest.raises(Exception, match="not enabled"):
            GetShareDetail(disabled_root, "shr_x")

    def test_get_share_detail(self, mock_root):
        mock_metadata_api = MagicMock()
        mock_res = MagicMock()
        mock_res.share_token = "shr_abc"
        mock_res.z_id = "zid"
        mock_res.env_zid = "env_zid"
        mock_res.share_mode = "public"
        mock_res.backend_mode = "proxy"
        mock_res.frontend_endpoints = ["https://x.io"]
        mock_res.target = "http://localhost:8080"
        mock_res.limited = False
        mock_res.created_at = 1000
        mock_res.updated_at = 2000
        mock_metadata_api.get_share_detail.return_value = mock_res

        with patch("zrok.share.MetadataApi", return_value=mock_metadata_api):
            detail = GetShareDetail(mock_root, "shr_abc")
            assert detail.Token == "shr_abc"
            assert detail.ShareMode == "public"
            assert detail.Target == "http://localhost:8080"
            assert detail.CreatedAt == 1000
