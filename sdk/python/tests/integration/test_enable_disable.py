"""Integration tests for enable/disable lifecycle."""

import os
import json
import pytest

from zrok.environment.enable import enable, disable
from zrok.environment.root import Root, Metadata, Config, Environment, V
from zrok.status import status


class TestEnableDisable:
    def test_enable_creates_environment(self, enabled_root):
        """Verify enable() produces a valid enabled Root."""
        assert enabled_root.IsEnabled()
        assert enabled_root.env.Token != ""
        assert enabled_root.env.ZitiIdentity != ""
        assert enabled_root.env.ApiEndpoint != ""

    def test_status_after_enable(self, enabled_root):
        """Verify status() reports enabled state."""
        s = status(enabled_root)
        assert s.Enabled is True
        assert s.Token != ""
        assert s.ZitiIdentity != ""

    def test_idempotent_enable(self, enabled_root, account_token):
        """Calling enable() again returns existing environment without error."""
        original_identity = enabled_root.env.ZitiIdentity
        env = enable(enabled_root, account_token)
        assert env.ZitiIdentity == original_identity

    def test_disable_and_reenable(self, tmp_path, account_token, zrok2_endpoint):
        """Full enable → disable → verify disabled cycle."""
        import zrok.environment.root as root_mod

        zrok_dir = tmp_path / ".zrok2_cycle"
        zrok_dir.mkdir()
        (zrok_dir / "identities").mkdir()

        env_file = str(zrok_dir / "environment.json")
        id_dir = str(zrok_dir / "identities")

        orig_env_file = root_mod.environmentFile
        orig_id_dir = root_mod.identitiesDir
        orig_id_file = root_mod.identityFile

        root_mod.environmentFile = lambda: env_file
        root_mod.identitiesDir = lambda: id_dir
        root_mod.identityFile = lambda name: os.path.join(id_dir, name + ".json")

        try:
            root = Root(
                meta=Metadata(V=V, RootPath=str(zrok_dir)),
                cfg=Config(ApiEndpoint=zrok2_endpoint),
                env=Environment(),
            )

            # Enable
            enable(root, account_token, description="cycle-test")
            assert root.IsEnabled()

            # Disable
            disable(root)
            assert not root.IsEnabled()
            assert not os.path.exists(env_file)
        finally:
            root_mod.environmentFile = orig_env_file
            root_mod.identitiesDir = orig_id_dir
            root_mod.identityFile = orig_id_file
