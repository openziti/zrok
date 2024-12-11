---
sidebar_position: 21
sidebar_label: Permission Modes
---

# Organizations

zrok (starting with `v0.4.45`) includes support for "organizations". Organizations are groups of related accounts that are typically centrally managed in some capacity. A zrok account can be a member of multiple organizations. Organization membership can also include an "admin" permission. As of `v0.4.45` organization admins are able to retrieve an "overview" (`zrok overview`) from any other account in the organization, allowing the admin to see the details of the environments, shares, and accesses created within that account.

Future zrok releases will include additional organization features, including `--closed` permission sharing functions.

## Configuring an Organization

