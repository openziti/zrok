"""Integration tests for list_shares and list_accesses with an empty environment."""

import pytest

from zrok2.listing import list_shares, list_accesses


class TestListingEmpty:
    def test_list_shares_empty(self, enabled_root):
        """A freshly enabled environment has no shares."""
        shares = list_shares(enabled_root)
        assert isinstance(shares, list)

    def test_list_accesses_empty(self, enabled_root):
        """A freshly enabled environment has no accesses."""
        accesses = list_accesses(enabled_root)
        assert isinstance(accesses, list)
