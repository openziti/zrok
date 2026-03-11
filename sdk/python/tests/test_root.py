"""Tests for zrok.environment.root — Root, Load, Default, ApiEndpoint."""

import json
import os
import sys
import pytest
from unittest.mock import patch, MagicMock

from zrok.environment.root import (
    Root, Metadata, Config, Environment, Default, Load, V, ApiEndpoint as ApiEndpointTuple,
)


class TestRootHasConfig:
    def test_has_config_when_set(self):
        root = Root(cfg=Config(ApiEndpoint="https://x.io"))
        assert root.HasConfig() is True

    def test_no_config_when_default(self):
        root = Root()
        assert root.HasConfig() is False


class TestRootIsEnabled:
    def test_enabled_with_environment(self):
        root = Root(env=Environment(Token="t", ZitiIdentity="z", ApiEndpoint="e"))
        assert root.IsEnabled() is True

    def test_not_enabled_when_default(self):
        root = Root()
        assert root.IsEnabled() is False


class TestApiEndpoint:
    def test_default_endpoint(self):
        root = Root()
        ep = root.ApiEndpoint()
        assert ep.endpoint == "https://api-v2.zrok.io"
        assert ep.frm == "binary"

    def test_config_overrides_default(self):
        root = Root(cfg=Config(ApiEndpoint="https://config.zrok.io"))
        ep = root.ApiEndpoint()
        assert ep.endpoint == "https://config.zrok.io"
        assert ep.frm == "config"

    @patch.dict(os.environ, {"ZROK2_API_ENDPOINT": "https://env2.zrok.io"}, clear=False)
    def test_zrok2_env_overrides_config(self):
        root = Root(cfg=Config(ApiEndpoint="https://config.zrok.io"))
        ep = root.ApiEndpoint()
        assert ep.endpoint == "https://env2.zrok.io"
        assert ep.frm == "ZROK2_API_ENDPOINT"

    @patch.dict(os.environ, {"ZROK_API_ENDPOINT": "https://legacy.zrok.io"}, clear=False)
    def test_legacy_env_fallback_with_warning(self, capsys):
        # Ensure ZROK2_ is not set
        os.environ.pop("ZROK2_API_ENDPOINT", None)
        root = Root()
        ep = root.ApiEndpoint()
        assert ep.endpoint == "https://legacy.zrok.io"
        assert ep.frm == "ZROK_API_ENDPOINT"
        captured = capsys.readouterr()
        assert "deprecated" in captured.err.lower()

    @patch.dict(os.environ, {
        "ZROK2_API_ENDPOINT": "https://new.zrok.io",
        "ZROK_API_ENDPOINT": "https://old.zrok.io",
    }, clear=False)
    def test_zrok2_takes_priority_over_legacy(self, capsys):
        root = Root()
        ep = root.ApiEndpoint()
        assert ep.endpoint == "https://new.zrok.io"
        assert ep.frm == "ZROK2_API_ENDPOINT"
        # No deprecation warning when ZROK2_ is set
        captured = capsys.readouterr()
        assert "deprecated" not in captured.err.lower()

    def test_env_overrides_all_when_enabled(self):
        root = Root(
            cfg=Config(ApiEndpoint="https://config.zrok.io"),
            env=Environment(Token="t", ZitiIdentity="z", ApiEndpoint="https://env.zrok.io"),
        )
        ep = root.ApiEndpoint()
        assert ep.endpoint == "https://env.zrok.io"
        assert ep.frm == "env"

    def test_trailing_slash_stripped(self):
        root = Root(cfg=Config(ApiEndpoint="https://config.zrok.io/"))
        ep = root.ApiEndpoint()
        assert not ep.endpoint.endswith("/")


class TestRootPersistence:
    def test_set_and_delete_environment(self, tmp_zrok_dir):
        root = Root(meta=Metadata(V=V, RootPath=str(tmp_zrok_dir)))

        # Patch dirs to use our tmp dir
        env_file = str(tmp_zrok_dir / "environment.json")
        id_dir = str(tmp_zrok_dir / "identities")
        id_file = str(tmp_zrok_dir / "identities" / "environment.json")

        with patch("zrok.environment.root.environmentFile", return_value=env_file), \
             patch("zrok.environment.root.identitiesDir", return_value=id_dir), \
             patch("zrok.environment.root.identityFile", return_value=id_file):
            env = Environment(Token="tok", ZitiIdentity="zid", ApiEndpoint="https://x.io")
            root.SetEnvironment(env)

            assert os.path.isfile(env_file)
            with open(env_file) as f:
                data = json.load(f)
            assert data["zrok_token"] == "tok"
            assert data["ziti_identity"] == "zid"
            assert data["api_endpoint"] == "https://x.io"
            assert root.IsEnabled()

            # Save identity
            root.SaveZitiIdentityNamed("environment", '{"ztAPI":"https://x"}')
            assert os.path.isfile(id_file)

            # Delete environment
            root.DeleteEnvironment()
            assert not os.path.isfile(env_file)
            assert not root.IsEnabled()

    def test_delete_environment_noop_when_no_files(self, tmp_zrok_dir):
        root = Root(meta=Metadata(V=V, RootPath=str(tmp_zrok_dir)))
        env_file = str(tmp_zrok_dir / "environment.json")
        id_file = str(tmp_zrok_dir / "identities" / "environment.json")
        with patch("zrok.environment.root.environmentFile", return_value=env_file), \
             patch("zrok.environment.root.identityFile", return_value=id_file):
            # Should not raise
            root.DeleteEnvironment()


class TestDefault:
    def test_default_creates_root_with_version(self):
        with patch("zrok.environment.root.rootDir", return_value="/tmp/fakezrok"):
            r = Default()
            assert r.meta.V == V
            assert r.meta.RootPath == "/tmp/fakezrok"
            assert not r.IsEnabled()


class TestClientVersionCheck:
    def test_client_raises_on_version_mismatch(self):
        root = Root(
            cfg=Config(ApiEndpoint="https://test.zrok.io"),
            env=Environment(Token="t", ZitiIdentity="z", ApiEndpoint="https://test.zrok.io"),
        )
        mock_client = MagicMock()
        mock_metadata_api = MagicMock()
        mock_response = MagicMock()
        mock_response.status_code = 400
        mock_metadata_api.client_version_check_with_http_info.return_value = mock_response

        with patch("zrok.environment.root.zrok.MetadataApi", return_value=mock_metadata_api):
            with pytest.raises(Exception, match="Client version check failed"):
                root.client_version_check(mock_client)

    def test_client_passes_on_200(self):
        root = Root(
            cfg=Config(ApiEndpoint="https://test.zrok.io"),
            env=Environment(Token="t", ZitiIdentity="z", ApiEndpoint="https://test.zrok.io"),
        )
        mock_client = MagicMock()
        mock_metadata_api = MagicMock()
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_metadata_api.client_version_check_with_http_info.return_value = mock_response

        with patch("zrok.environment.root.zrok.MetadataApi", return_value=mock_metadata_api):
            # Should not raise
            root.client_version_check(mock_client)
