"""Shared fixtures for zrok SDK unit tests."""

import json
import os
import pytest
from unittest.mock import MagicMock, patch

from zrok.environment.root import Root, Metadata, Config, Environment, V


@pytest.fixture
def tmp_zrok_dir(tmp_path):
    """Create a temporary .zrok directory structure."""
    zrok_dir = tmp_path / ".zrok"
    zrok_dir.mkdir()
    (zrok_dir / "identities").mkdir()

    # Write metadata
    with open(zrok_dir / "metadata.json", "w") as f:
        json.dump({"v": V}, f)

    return zrok_dir


@pytest.fixture
def mock_api_client():
    """A mock zrok_api.ApiClient."""
    client = MagicMock()
    client.configuration = MagicMock()
    return client


@pytest.fixture
def mock_root(tmp_zrok_dir, mock_api_client):
    """A Root with an enabled environment and mocked Client()."""
    root = Root(
        meta=Metadata(V=V, RootPath=str(tmp_zrok_dir)),
        cfg=Config(ApiEndpoint="https://test.zrok.io"),
        env=Environment(
            Token="test-token-abc123",
            ZitiIdentity="test-ziti-id",
            ApiEndpoint="https://test.zrok.io",
        ),
    )

    # Patch Client() to return our mock without making real HTTP calls
    with patch.object(root, "Client", return_value=mock_api_client):
        yield root


@pytest.fixture
def disabled_root(tmp_zrok_dir):
    """A Root that is NOT enabled (empty Environment)."""
    root = Root(
        meta=Metadata(V=V, RootPath=str(tmp_zrok_dir)),
        cfg=Config(ApiEndpoint="https://test.zrok.io"),
        env=Environment(),  # not enabled
    )
    return root
