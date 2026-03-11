"""Integration tests for name creation, listing, and deletion."""

import pytest

from zrok.name import create_name, delete_name, list_names, list_namespaces


class TestNameLifecycle:
    def test_list_namespaces(self, enabled_root):
        """The bootstrap creates a 'public' namespace — verify it exists."""
        nss = list_namespaces(enabled_root)
        assert len(nss) >= 1
        names = [ns.Name for ns in nss]
        assert "public" in names

    def test_create_and_delete_name(self, enabled_root):
        """Create a name in the public namespace, verify listing, then delete."""
        # Get the public namespace token
        nss = list_namespaces(enabled_root)
        public_ns = next(ns for ns in nss if ns.Name == "public")

        entry = create_name(enabled_root, "integ-test-name", namespace_token=public_ns.NamespaceToken)
        assert entry.Name == "integ-test-name"

        try:
            # Verify it appears in listing
            names = list_names(enabled_root, namespace_token=public_ns.NamespaceToken)
            found = [n for n in names if n.Name == "integ-test-name"]
            assert len(found) == 1
        finally:
            delete_name(enabled_root, "integ-test-name", namespace_token=public_ns.NamespaceToken)

        # Verify it's gone
        names_after = list_names(enabled_root, namespace_token=public_ns.NamespaceToken)
        found_after = [n for n in names_after if n.Name == "integ-test-name"]
        assert len(found_after) == 0
