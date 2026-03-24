"""Tests for zrok.environment.enable — enable() and disable()."""

import pytest
from unittest.mock import patch, MagicMock

from zrok2.environment.enable import enable, disable
from zrok2.environment.root import Root, Metadata, Config, Environment, V


@pytest.fixture
def fresh_root(tmp_zrok_dir):
    """A Root that is not yet enabled, with Client() mocked."""
    root = Root(
        meta=Metadata(V=V, RootPath=str(tmp_zrok_dir)),
        cfg=Config(ApiEndpoint="https://test.zrok.io"),
        env=Environment(),
    )
    return root


class TestEnable:
    def test_enable_returns_existing_when_already_enabled(self, mock_root):
        result = enable(mock_root, "new-token")
        assert result.Token == "test-token-abc123"

    def test_enable_creates_environment(self, fresh_root, tmp_zrok_dir):
        mock_client = MagicMock()
        mock_env_api = MagicMock()

        mock_response = MagicMock()
        mock_response.data = MagicMock()
        mock_response.data.identity = "new-ziti-id"
        mock_response.data.cfg = '{"ztAPI":"https://x"}'
        mock_env_api.enable_with_http_info.return_value = mock_response

        env_file = str(tmp_zrok_dir / "environment.json")
        id_dir = str(tmp_zrok_dir / "identities")
        id_file = str(tmp_zrok_dir / "identities" / "environment.json")

        with patch.object(fresh_root, "Client", return_value=mock_client), \
             patch("zrok.environment.enable.EnvironmentApi", return_value=mock_env_api), \
             patch("zrok.environment.root.environmentFile", return_value=env_file), \
             patch("zrok.environment.root.identitiesDir", return_value=id_dir), \
             patch("zrok.environment.root.identityFile", return_value=id_file):
            env = enable(fresh_root, "my-token", description="test host")
            assert env.Token == "my-token"
            assert env.ZitiIdentity == "new-ziti-id"
            assert env.ApiEndpoint == "https://test.zrok.io"
            assert fresh_root.IsEnabled()

    def test_enable_resets_env_on_client_error(self, fresh_root):
        with patch.object(fresh_root, "Client", side_effect=Exception("conn refused")):
            with pytest.raises(Exception, match="error getting zrok client"):
                enable(fresh_root, "my-token")
        assert not fresh_root.IsEnabled()

    def test_enable_resets_env_on_api_error(self, fresh_root):
        mock_client = MagicMock()
        mock_env_api = MagicMock()
        mock_env_api.enable_with_http_info.side_effect = Exception("500 server error")

        with patch.object(fresh_root, "Client", return_value=mock_client), \
             patch("zrok.environment.enable.EnvironmentApi", return_value=mock_env_api):
            with pytest.raises(Exception, match="unable to enable environment"):
                enable(fresh_root, "my-token")
        assert not fresh_root.IsEnabled()


class TestDisable:
    def test_disable_noop_when_not_enabled(self, disabled_root):
        # Should not raise or call any API
        disable(disabled_root)

    def test_disable_calls_api_and_cleans_up(self, mock_root, tmp_zrok_dir):
        mock_root.Client()
        mock_env_api = MagicMock()

        env_file = str(tmp_zrok_dir / "environment.json")
        id_file = str(tmp_zrok_dir / "identities" / "environment.json")

        with patch("zrok.environment.enable.EnvironmentApi", return_value=mock_env_api), \
             patch("zrok.environment.root.environmentFile", return_value=env_file), \
             patch("zrok.environment.root.identityFile", return_value=id_file):
            disable(mock_root)
            mock_env_api.disable_with_http_info.assert_called_once()
            assert not mock_root.IsEnabled()

    def test_disable_raises_on_client_error(self):
        root = Root(
            env=Environment(Token="t", ZitiIdentity="z", ApiEndpoint="https://x.io"),
        )
        with patch.object(root, "Client", side_effect=Exception("conn refused")):
            with pytest.raises(Exception, match="error getting zrok client"):
                disable(root)
