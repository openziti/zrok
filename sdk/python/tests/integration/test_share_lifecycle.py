"""Integration tests for share creation, listing, detail, and deletion."""

import pytest

from zrok2.share import CreateShare, DeleteShare, GetShareDetail
from zrok2.listing import list_shares
import zrok2.model as model


class TestShareLifecycle:
    def test_create_and_delete_private_share(self, enabled_root):
        """Create a private tcpTunnel share, verify it appears in listing, then delete."""
        req = model.ShareRequest(
            BackendMode=model.TCP_TUNNEL_BACKEND_MODE,
            ShareMode=model.PRIVATE_SHARE_MODE,
            Target="tcp://localhost:9999",
        )

        share = CreateShare(enabled_root, req)
        assert share.Token != ""

        try:
            # Verify it shows up in list_shares
            shares = list_shares(enabled_root)
            tokens = [s.Token for s in shares]
            assert share.Token in tokens

            # Get detail
            detail = GetShareDetail(enabled_root, share.Token)
            assert detail.Token == share.Token
            assert detail.BackendMode == model.TCP_TUNNEL_BACKEND_MODE
            assert detail.ShareMode == model.PRIVATE_SHARE_MODE
        finally:
            DeleteShare(enabled_root, share)

        # Verify it's gone from listing
        shares_after = list_shares(enabled_root)
        tokens_after = [s.Token for s in shares_after]
        assert share.Token not in tokens_after
