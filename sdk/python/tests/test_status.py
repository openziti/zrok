"""Tests for zrok.status — status()."""

from zrok2.status import status
from zrok2.environment.root import Root


class TestStatus:
    def test_status_enabled(self, mock_root):
        s = status(mock_root)
        assert s.Enabled is True
        assert s.Token == "test-token-abc123"
        assert s.ZitiIdentity == "test-ziti-id"
        assert s.ApiEndpoint == "https://test.zrok.io"
        assert s.ApiEndpointSource == "env"

    def test_status_disabled(self, disabled_root):
        s = status(disabled_root)
        assert s.Enabled is False
        assert s.Token == ""
        assert s.ZitiIdentity == ""
        assert s.ApiEndpoint == "https://test.zrok.io"
        assert s.ApiEndpointSource == "config"

    def test_status_default_endpoint(self):
        root = Root()
        s = status(root)
        assert s.Enabled is False
        assert s.ApiEndpoint == "https://api-v2.zrok.io"
        assert s.ApiEndpointSource == "binary"
