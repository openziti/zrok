---
sidebar_position: 40
---

# Configuring Limits

:::note
This guide is current as of zrok version `v0.4.31`.
:::

:::warning
If you have not yet configured [metrics](configuring-metrics.md), please visit the [metrics guide](configuring-metrics.md) first before working through the limits configuration.
:::

## Understanding the zrok Limits Agent

The limits agent in zrok is used to control the amount of resources that can be consumed by any account in a service instance. 

Limits can be specified that control the number of environments, shares, reserved shares, and unique names. Limits that control the number of resources are called _resource count limits_.

Limits can be specified that control the amount of data that can be transferred for different types of share backend modes. Limits that control the amount of data that can be transferred are called _bandwidth limits_.

The limits agent in zrok is responsible for controlling the number of resources in use (environments, shares) and also for ensuring that any single account, environment, or share is held below the configured thresholds.

zrok limits can be specified _globally_, applying to all users in a service instance. Individual limits can be specified and applied to individual accounts using a new facility called _limit classes_. Limit classes can be used to specify resource count and bandwidth limit defaults per-account. Separate limits for each type share backend can also be specified and applied to user accounts.

## The Global Configuration

The reference configuration for the zrok controller (found at [`etc/ctrl.yaml`](https://github.com/openziti/zrok/blob/main/etc/ctrl.yml) in the [repository](https://github.com/openziti/zrok)) contains the global limits configuration, which looks like this:

```yaml
# Service instance limits global configuration.
#
# See `docs/guides/metrics-and-limits/configuring-limits.md` for details.
#
limits:
  environments:     -1
  shares:           -1
  reserved_shares:  -1
  unique_names:     -1
  bandwidth:
    period:         5m
    warning:
      rx:           -1
      tx:           -1
      total:        7242880
    limit:
      rx:           -1
      tx:           -1
      total:        10485760
  enforcing:        false
  cycle:            5m
```

:::note
A value of `-1` appearing in the limits configuration mean the value is _unlimited_.
:::

The `enforcing` boolean specifies whether or not limits are enabled in the service instance. By default, limits is disabled. No matter what else is configured in this stanza, if `enforcing` is set to `false`, there will be no limits placed on any account in the service instance.

The `cycle` value controls how frequently the limits agent will evaluate enforced limits. When a user exceeds a limit and has their shares disabled, the limits agent will evaluate their bandwidth usage on this interval looking to "relax" the limit once their usage falls below the threshold.

### Global Resouce Count Limits

The `environments`, `shares`, `reserved_shares`, and `unique_names` specify the resource count limits, globally for the service instance. 

These resource counts will be applied to all users in the service instance by default.

### Global Bandwidth Limits

The `bandwidth` section defines the global bandwidth limits for all users in the service instance.

There are two levels of bandwidth limits that can be specified in the global configuration. The first limit defines a _warning_ threshold where the user will receive an email that they are using increased data transfer amounts and will ultimately be subject to a limit.

The second limit defines the the actual _limit_ threshold, where the limits agent will disabled traffic for the account's shares.

Bandwidth limits can be specified in terms of `tx` (or _transmitted_ data), `rx` (or _received_ data), and the `total` bytes that are sent in either direction. If you only want to set the `total` transferred limit, you can set `rx` and `tx` to `-1` (for _unlimited_). You can configure any combination of these these values at either the limit or warning levels.

The `period` specifies the time window for the bandwidth limit. See the documentation for [`time.Duration.ParseDuration`](https://pkg.go.dev/time#ParseDuration) for details about the format used for these durations. If the `period` is set to 5 minutes, then the limits agent will monitor the transmitted and receivde traffic for the account for the last 5 minutes, and if the amount of data is greater than either the `warning` or the `limit` threshold, action will be taken.

## Limit Actions

When an account exceeds a bandwidth limit, the limits agent will seek to limit the affected shares (based on the combination of global configuration, unscoped limit classes, and scoped limit classes). It applies the limit by removing the underlying OpenZiti dial policies for any frontends that are trying to access the share.

This means that public frontends will simply return a `404` as if the share is no longer there. Private frontends will also return `404` errors. When the limit is relaxed, the dial policies are put back in place and the share will continue operating normally.

## Unlimited Accounts

The `accounts` table in the database includes a `limitless` column. When this column is set to `true` the account is not subject to any of the limits in the system.