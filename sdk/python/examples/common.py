"""Shared environment setup for zrok2 examples.

Handles the full account + environment lifecycle:

1. If the environment is already enabled, uses it as-is.
2. If ZROK2_ENABLE_TOKEN is set, enables with that token.
3. If ZROK2_ADMIN_TOKEN is set (and no account token), creates a fresh
   account via the admin API, then enables the environment.

Registers atexit cleanup to disable the environment if this module enabled it.

Usage in an example script:

    from common import get_root
    root = get_root()
"""

import atexit
import json
import os
import time

import urllib3

from zrok2.environment import root as root_mod
from zrok2.environment.enable import enable, disable


def _create_account(api_endpoint, admin_token):
    """Create a test account via the admin API, return the account token."""
    http = urllib3.PoolManager()
    resp = http.request(
        "POST",
        f"{api_endpoint}/api/v2/account",
        headers={
            "X-TOKEN": admin_token,
            "Content-Type": "application/zrok.v1+json",
        },
        body=json.dumps({
            "email": f"example-{os.getpid()}-{int(time.time())}@zrok.internal",
            "password": "example-test-password",
        }).encode(),
    )
    if resp.status != 201:
        raise RuntimeError(
            f"failed to create account: HTTP {resp.status} {resp.data.decode()}"
        )
    data = json.loads(resp.data)
    return data["accountToken"]


def get_root():
    """Load or enable a zrok2 environment root.

    Returns:
        A Root instance ready for share/access operations.
    """
    try:
        root = root_mod.Load()
        if root.IsEnabled():
            return root
    except Exception:
        pass

    # Not yet enabled — build a root from metadata + config only.
    # Load() fails when environment.json is missing, so construct manually.
    from zrok2.environment.root import Root, Metadata, Config, V, rootDir, configFile
    import json as _json
    r = Root()
    r.meta = Metadata(V=V, RootPath=rootDir())
    cf = configFile()
    if os.path.isfile(cf):
        with open(cf) as f:
            cfg = _json.load(f)
        r.cfg = Config(ApiEndpoint=cfg.get("apiEndpoint", ""))
    root = r

    account_token = os.environ.get("ZROK2_ENABLE_TOKEN", "")

    if not account_token:
        # Try to create an account automatically if admin credentials exist
        api_endpoint = os.environ.get("ZROK2_API_ENDPOINT", "")
        admin_token = os.environ.get("ZROK2_ADMIN_TOKEN", "")
        if api_endpoint and admin_token:
            account_token = _create_account(api_endpoint, admin_token)
        else:
            raise RuntimeError(
                "environment is not enabled; set ZROK2_ENABLE_TOKEN or "
                "both ZROK2_API_ENDPOINT and ZROK2_ADMIN_TOKEN"
            )

    enable(root, account_token, description="example")
    root = root_mod.Load()

    def _cleanup():
        try:
            if root.IsEnabled():
                disable(root)
        except Exception:
            pass

    atexit.register(_cleanup)
    return root
