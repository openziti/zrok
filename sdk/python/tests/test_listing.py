"""Tests for zrok.listing — list_shares, list_accesses."""

import pytest
from unittest.mock import patch, MagicMock

from zrok.listing import list_shares, list_accesses
import zrok.model as model


class TestListShares:
    def test_not_enabled_raises(self, disabled_root):
        with pytest.raises(Exception, match="not enabled"):
            list_shares(disabled_root)

    def test_list_shares(self, mock_root):
        mock_metadata_api = MagicMock()
        mock_share = MagicMock()
        mock_share.share_token = "shr_abc"
        mock_share.z_id = "zid"
        mock_share.env_zid = "env_zid"
        mock_share.share_mode = "public"
        mock_share.backend_mode = "proxy"
        mock_share.frontend_endpoints = ["https://x.io"]
        mock_share.target = "http://localhost:8080"
        mock_share.limited = False
        mock_share.created_at = 1000
        mock_share.updated_at = 2000
        mock_res = MagicMock()
        mock_res.shares = [mock_share]
        mock_metadata_api.list_shares.return_value = mock_res

        with patch("zrok.listing.MetadataApi", return_value=mock_metadata_api):
            shares = list_shares(mock_root)
            assert len(shares) == 1
            assert shares[0].Token == "shr_abc"
            assert shares[0].ShareMode == "public"

    def test_list_shares_empty(self, mock_root):
        mock_metadata_api = MagicMock()
        mock_res = MagicMock()
        mock_res.shares = []
        mock_metadata_api.list_shares.return_value = mock_res

        with patch("zrok.listing.MetadataApi", return_value=mock_metadata_api):
            shares = list_shares(mock_root)
            assert shares == []

    def test_list_shares_none_shares(self, mock_root):
        mock_metadata_api = MagicMock()
        mock_res = MagicMock()
        mock_res.shares = None
        mock_metadata_api.list_shares.return_value = mock_res

        with patch("zrok.listing.MetadataApi", return_value=mock_metadata_api):
            shares = list_shares(mock_root)
            assert shares == []


class TestListAccesses:
    def test_not_enabled_raises(self, disabled_root):
        with pytest.raises(Exception, match="not enabled"):
            list_accesses(disabled_root)

    def test_list_accesses(self, mock_root):
        mock_metadata_api = MagicMock()
        mock_access = MagicMock()
        mock_access.id = 1
        mock_access.frontend_token = "fe_tok"
        mock_access.env_zid = "env_zid"
        mock_access.share_token = "shr_abc"
        mock_access.backend_mode = "tcpTunnel"
        mock_access.bind_address = "127.0.0.1:8080"
        mock_access.description = "test access"
        mock_access.limited = False
        mock_access.created_at = 1000
        mock_access.updated_at = 2000
        mock_res = MagicMock()
        mock_res.accesses = [mock_access]
        mock_metadata_api.list_accesses.return_value = mock_res

        with patch("zrok.listing.MetadataApi", return_value=mock_metadata_api):
            accesses = list_accesses(mock_root)
            assert len(accesses) == 1
            assert accesses[0].FrontendToken == "fe_tok"
            assert accesses[0].ShareToken == "shr_abc"

    def test_list_accesses_empty(self, mock_root):
        mock_metadata_api = MagicMock()
        mock_res = MagicMock()
        mock_res.accesses = []
        mock_metadata_api.list_accesses.return_value = mock_res

        with patch("zrok.listing.MetadataApi", return_value=mock_metadata_api):
            accesses = list_accesses(mock_root)
            assert accesses == []
