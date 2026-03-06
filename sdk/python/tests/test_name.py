"""Tests for zrok.name — create_name, delete_name, list_names, list_namespaces."""

import pytest
from unittest.mock import patch, MagicMock

from zrok2.name import create_name, delete_name, list_names, list_namespaces


class TestCreateName:
    def test_not_enabled_raises(self, disabled_root):
        with pytest.raises(Exception, match="not enabled"):
            create_name(disabled_root, "myname")

    def test_create_name(self, mock_root):
        mock_share_api = MagicMock()

        with patch("zrok.name.ShareApi", return_value=mock_share_api):
            entry = create_name(mock_root, "myname", namespace_token="ns1")
            assert entry.Name == "myname"
            assert entry.NamespaceToken == "ns1"
            mock_share_api.create_share_name_with_http_info.assert_called_once()

    def test_create_name_default_namespace(self, mock_root):
        mock_share_api = MagicMock()

        with patch("zrok.name.ShareApi", return_value=mock_share_api):
            entry = create_name(mock_root, "myname")
            assert entry.NamespaceToken == ""


class TestDeleteName:
    def test_not_enabled_raises(self, disabled_root):
        with pytest.raises(Exception, match="not enabled"):
            delete_name(disabled_root, "myname")

    def test_delete_name(self, mock_root):
        mock_share_api = MagicMock()

        with patch("zrok.name.ShareApi", return_value=mock_share_api):
            delete_name(mock_root, "myname", namespace_token="ns1")
            mock_share_api.delete_share_name_with_http_info.assert_called_once()


class TestListNames:
    def test_not_enabled_raises(self, disabled_root):
        with pytest.raises(Exception, match="not enabled"):
            list_names(disabled_root)

    def test_list_names_in_namespace(self, mock_root):
        mock_share_api = MagicMock()
        mock_name = MagicMock()
        mock_name.namespace_token = "ns1"
        mock_name.namespace_name = "public"
        mock_name.name = "myname"
        mock_name.share_token = "shr_abc"
        mock_name.reserved = True
        mock_name.created_at = 1000
        mock_share_api.list_names_for_namespace.return_value = [mock_name]

        with patch("zrok.name.ShareApi", return_value=mock_share_api):
            names = list_names(mock_root, namespace_token="ns1")
            assert len(names) == 1
            assert names[0].Name == "myname"
            assert names[0].Reserved is True

    def test_list_names_all_namespaces(self, mock_root):
        mock_share_api = MagicMock()
        mock_ns = MagicMock()
        mock_ns.namespace_token = "ns1"
        mock_share_api.list_share_namespaces.return_value = [mock_ns]

        mock_name = MagicMock()
        mock_name.namespace_token = "ns1"
        mock_name.namespace_name = "public"
        mock_name.name = "myname"
        mock_name.share_token = ""
        mock_name.reserved = False
        mock_name.created_at = 0
        mock_share_api.list_names_for_namespace.return_value = [mock_name]

        with patch("zrok.name.ShareApi", return_value=mock_share_api):
            names = list_names(mock_root)
            assert len(names) == 1


class TestListNamespaces:
    def test_not_enabled_raises(self, disabled_root):
        with pytest.raises(Exception, match="not enabled"):
            list_namespaces(disabled_root)

    def test_list_namespaces(self, mock_root):
        mock_share_api = MagicMock()
        mock_ns = MagicMock()
        mock_ns.namespace_token = "ns1"
        mock_ns.name = "public"
        mock_ns.description = "Public namespace"
        mock_share_api.list_share_namespaces.return_value = [mock_ns]

        with patch("zrok.name.ShareApi", return_value=mock_share_api):
            nss = list_namespaces(mock_root)
            assert len(nss) == 1
            assert nss[0].NamespaceToken == "ns1"
            assert nss[0].Name == "public"

    def test_list_namespaces_empty(self, mock_root):
        mock_share_api = MagicMock()
        mock_share_api.list_share_namespaces.return_value = []

        with patch("zrok.name.ShareApi", return_value=mock_share_api):
            nss = list_namespaces(mock_root)
            assert nss == []
