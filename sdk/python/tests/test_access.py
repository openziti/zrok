"""Tests for zrok.access — CreateAccess, DeleteAccess, Access context manager."""

import pytest
from unittest.mock import patch, MagicMock

from zrok2.access import CreateAccess, DeleteAccess, Access
import zrok2.model as model


class TestCreateAccess:
    def test_not_enabled_raises(self, disabled_root):
        req = model.AccessRequest(ShareToken="shr_abc")
        with pytest.raises(Exception, match="not enabled"):
            CreateAccess(disabled_root, req)

    def test_create_access(self, mock_root):
        mock_share_api = MagicMock()
        mock_res = MagicMock()
        mock_res.frontend_token = "fe_tok"
        mock_res.backend_mode = "tcpTunnel"
        mock_share_api.access.return_value = mock_res

        req = model.AccessRequest(ShareToken="shr_abc")

        with patch("zrok.access.ShareApi", return_value=mock_share_api):
            acc = CreateAccess(mock_root, req)
            assert acc.Token == "fe_tok"
            assert acc.ShareToken == "shr_abc"
            assert acc.BackendMode == "tcpTunnel"

    def test_create_access_api_error(self, mock_root):
        mock_share_api = MagicMock()
        mock_share_api.access.side_effect = Exception("404 not found")

        req = model.AccessRequest(ShareToken="shr_bad")

        with patch("zrok.access.ShareApi", return_value=mock_share_api):
            with pytest.raises(Exception, match="unable to create access"):
                CreateAccess(mock_root, req)


class TestDeleteAccess:
    def test_delete_access(self, mock_root):
        mock_share_api = MagicMock()
        acc = model.Access(Token="fe_tok", ShareToken="shr_abc", BackendMode="proxy")

        with patch("zrok.access.ShareApi", return_value=mock_share_api):
            DeleteAccess(mock_root, acc)
            mock_share_api.unaccess.assert_called_once()

    def test_delete_access_error(self, mock_root):
        mock_share_api = MagicMock()
        mock_share_api.unaccess.side_effect = Exception("500")
        acc = model.Access(Token="fe_tok", ShareToken="shr_abc", BackendMode="proxy")

        with patch("zrok.access.ShareApi", return_value=mock_share_api):
            with pytest.raises(Exception, match="error deleting access"):
                DeleteAccess(mock_root, acc)


class TestAccessContextManager:
    def test_context_manager_creates_and_deletes(self, mock_root):
        mock_share_api = MagicMock()
        mock_res = MagicMock()
        mock_res.frontend_token = "fe_tok"
        mock_res.backend_mode = "proxy"
        mock_share_api.access.return_value = mock_res

        req = model.AccessRequest(ShareToken="shr_abc")

        with patch("zrok.access.ShareApi", return_value=mock_share_api):
            with Access(mock_root, req) as acc:
                assert acc.Token == "fe_tok"
            # unaccess should have been called on exit
            mock_share_api.unaccess.assert_called_once()
