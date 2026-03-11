"""Fixtures for integration tests against a live zrok2 instance.

Required environment variables:
  ZROK2_API_ENDPOINT  — zrok2 controller URL (e.g. http://localhost:18080)
  ZROK2_ADMIN_TOKEN   — admin secret for account creation
"""

import os
import json
import uuid
import pytest

import zrok_api as zrok
from zrok_api.api import AdminApi
from zrok_api.configuration import Configuration
from zrok_api.models.login_request import LoginRequest
from zrok.environment.root import Root, Metadata, Config, Environment, V
from zrok.environment.enable import enable, disable


@pytest.fixture(scope="session")
def zrok2_endpoint():
    """Return the zrok2 controller API endpoint from env."""
    ep = os.environ.get("ZROK2_API_ENDPOINT")
    if not ep:
        pytest.skip("ZROK2_API_ENDPOINT not set — skipping integration tests")
    return ep


@pytest.fixture(scope="session")
def admin_token():
    """Return the admin token from env."""
    token = os.environ.get("ZROK2_ADMIN_TOKEN")
    if not token:
        pytest.skip("ZROK2_ADMIN_TOKEN not set — skipping integration tests")
    return token


@pytest.fixture(scope="session")
def account_token(zrok2_endpoint, admin_token):
    """Create a test account via the admin API, return the account token."""
    cfg = Configuration()
    cfg.host = zrok2_endpoint + "/api/v2"
    cfg.api_key["key"] = admin_token

    client = zrok.ApiClient(configuration=cfg)
    admin_api = AdminApi(client)

    unique_id = uuid.uuid4().hex[:8]
    req = LoginRequest(
        email=f"integration-test-{unique_id}@zrok2.local",
        password="integration-test-password-1234",
    )
    custom_headers = {
        'Accept': 'application/json, application/zrok.v1+json'
    }

    try:
        resp = admin_api.create_account_with_http_info(body=req, _headers=custom_headers)
        if hasattr(resp, 'data') and resp.data is not None:
            return resp.data.account_token
        raise Exception("create_account returned no data")
    except Exception as e:
        # Handle the content type mismatch like other SDK calls
        from zrok_api.exceptions import ApiException
        if isinstance(e, ApiException) and hasattr(e, 'body') and e.body:
            data = json.loads(e.body)
            return data.get("accountToken") or data.get("account_token")
        raise


@pytest.fixture
def fresh_root(tmp_path, zrok2_endpoint):
    """Return a Root pointed at the live controller, not yet enabled."""
    root = Root(
        meta=Metadata(V=V, RootPath=str(tmp_path)),
        cfg=Config(ApiEndpoint=zrok2_endpoint),
        env=Environment(),
    )
    return root


@pytest.fixture
def enabled_root(tmp_path, account_token, zrok2_endpoint):
    """Enable a fresh environment, yield the Root, disable on teardown."""
    zrok_dir = tmp_path / ".zrok"
    zrok_dir.mkdir()
    (zrok_dir / "identities").mkdir()

    # Write metadata so Load() would work
    with open(zrok_dir / "metadata.json", "w") as f:
        json.dump({"v": V}, f)

    root = Root(
        meta=Metadata(V=V, RootPath=str(zrok_dir)),
        cfg=Config(ApiEndpoint=zrok2_endpoint),
        env=Environment(),
    )

    # Patch dirs to use tmp_path
    import zrok.environment.root as root_mod
    orig_env_file = root_mod.environmentFile
    orig_id_dir = root_mod.identitiesDir
    orig_id_file = root_mod.identityFile

    env_file = str(zrok_dir / "environment.json")
    id_dir = str(zrok_dir / "identities")

    root_mod.environmentFile = lambda: env_file
    root_mod.identitiesDir = lambda: id_dir
    root_mod.identityFile = lambda name: os.path.join(id_dir, name + ".json")

    try:
        env = enable(root, account_token, description="integration-test")
        assert root.IsEnabled()
        yield root
    finally:
        try:
            if root.IsEnabled():
                disable(root)
        except Exception:
            pass
        root_mod.environmentFile = orig_env_file
        root_mod.identitiesDir = orig_id_dir
        root_mod.identityFile = orig_id_file
