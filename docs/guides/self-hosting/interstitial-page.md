---
title: Interstitial Page Configuration
sidebar_label: Interstitial Pages
sidebar_position: 18
---

On large zrok installations that support open registration and shared public frontends, abuse can become an issue. In order to mitigate phishing and other similar forms of abuse, zrok offers an interstitial page that announces to the visiting user that the share is hosted through zrok, and probably isn't their financial institution.

Interstitial pages can be enabled on a per-frontend basis, allowing the interstitial to be enabled on shared public frontends, but not private, closed frontends. The interstitial page requirement can also be overridden on a per-account basis, allowing shares created by specific accounts to bypass the interstitial requirement on frontends that enable it.

By default, if you do not specifically enable interstitial pages, then your self-hosted service instance will not offer them.

Let's take a look at how the interstitial pages mechanism works:

![zrok_interstitial_rendezvous](../../images/zrok_interstitial_rendezvous.png)

Every zrok share has a _config_ recorded in the underlying OpenZiti network. The config is of type `zrok.proxy.v1`. The frontend uses the information in this config to understand the disposition of the share. The config can contain an `interstitial: true` setting. If the config has this setting, and the frontend is configured to enable interstitial pages, then end users accessing the share will receive the interstitial page on first visit.

By default the zrok controller will record `interstitial: true` in the share config _unless_ a row is present in the `skip_interstitial_grants` table in the underlying database. The `skip_interstitial_grants` table is a basic SQL structure that allows inserting a row per account. 

```
create table skip_interstitial_grants (
    id          serial         primary key,

    account_id  integer        references accounts (id) not null,

    created_at  timestamptz    not null default(current_timestamp),
    updated_at  timestamptz    not null default(current_timestamp),
    deleted     boolean        not null default(false)
);
```

If an account has a row present in this table when creating a share, then the controller will write `interstitial: false` into the config for the share, which will bypass the interstitial regardless of frontend configuration.

The frontend config looks like this:

```
# Setting the `interstitial` setting to `true` will allow this frontend 
# to offer interstitial pages if they are configured on the share by the 
# controller.
#
#interstitial: true
```

Simply setting `interstitial: true` in the controller config will allow the configured frontend to offer interstitial pages.

## Bypassing the Interstitial

End users can offer an HTTP header of `skip_zrok_interstitial`, set to any value to bypass the interstitial page. Setting this header means that the user most likely understands what a zrok share is and will hopefully not fall for a phishing attack.

This header is especially useful for API clients (like `curl`).