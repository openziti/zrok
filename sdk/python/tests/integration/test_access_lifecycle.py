"""Integration tests for access creation, listing, and deletion."""

import pytest

from zrok.share import CreateShare, DeleteShare
from zrok.access import CreateAccess, DeleteAccess
from zrok.listing import list_accesses
import zrok.model as model


class TestAccessLifecycle:
    def test_create_and_delete_access(self, enabled_root):
        """Create a private share, access it, verify listing, then clean up."""
        share_req = model.ShareRequest(
            BackendMode=model.TCP_TUNNEL_BACKEND_MODE,
            ShareMode=model.PRIVATE_SHARE_MODE,
            Target="tcp://localhost:9998",
        )

        share = CreateShare(enabled_root, share_req)
        assert share.Token != ""

        try:
            # Create access
            access_req = model.AccessRequest(ShareToken=share.Token)
            acc = CreateAccess(enabled_root, access_req)
            assert acc.Token != ""
            assert acc.ShareToken == share.Token

            try:
                # Verify it appears in listing
                accesses = list_accesses(enabled_root)
                tokens = [a.FrontendToken for a in accesses]
                assert acc.Token in tokens
            finally:
                DeleteAccess(enabled_root, acc)

            # Verify access is gone
            accesses_after = list_accesses(enabled_root)
            tokens_after = [a.FrontendToken for a in accesses_after]
            assert acc.Token not in tokens_after
        finally:
            DeleteShare(enabled_root, share)
