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

The limits agent is a component of the zrok controller. It can be enabled and configured through the zrok controller configuration.

The limits agent is responsible for controlling the number of resources in use (environments, shares, etc.) and also for ensuring that accounts are held below the configured data transfer bandwidth thresholds. The limits agent exists to manage resource consumption for larger, multi-user zrok installations.

### Types of Limits

Limits can be specified that control the number of environments, shares, reserved shares, unique names, and frontends per-share that can be created by an account. Limits that control the allowed number of resources are called _resource count limits_.

Limits can be specified to control the amount of data that can be transferred within a time period. Limits that control the amount of data that can be transferred are called _bandwidth limits_.

zrok limits can be specified _globally_, applying to all users in a service instance. Limit _classes_ can be created to provide additional levels of resource allocation. Limit classes can then be _applied_ to multiple accounts, to alter their limit allocation beyond what's configured in the global configuration.

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
  share_frontends:  -1
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

The `environments`, `shares`, `reserved_shares`, `unique_names`, and `share_frontends` specify the resource count limits, globally for the service instance. 

These resource counts will be applied to all users in the service instance by default.

### Global Bandwidth Limits

The `bandwidth` section defines the global bandwidth limits for all users in the service instance.

There are two levels of bandwidth limits that can be specified in the global configuration. The first limit defines a _warning_ threshold where the user will receive an email that they are using increased data transfer amounts and will ultimately be subject to a limit. If you do not want this warning email to be sent, then configure all of the values to `-1` (unlimited).

The second limit defines the the actual _limit_ threshold, where the limits agent will disabled traffic for the account's shares.

Bandwidth limits can be specified in terms of `tx` (or _transmitted_ data), `rx` (or _received_ data), and the `total` bytes that are sent in either direction. If you only want to set the `total` transferred limit, you can set `rx` and `tx` to `-1` (for _unlimited_). You can configure any combination of these these values at either the limit or warning levels.

The `period` specifies the time window for the bandwidth limit. See the documentation for [`time.Duration.ParseDuration`](https://pkg.go.dev/time#ParseDuration) for details about the format used for these durations. If the `period` is set to 5 minutes, then the limits agent will monitor the transmitted and receivde traffic for the account for the last 5 minutes, and if the amount of data is greater than either the `warning` or the `limit` threshold, action will be taken.

In the global configuration example above users are allowed to transfer a total of `10485760` bytes in a `5m` period, and they will receive a warning email after they transfer more than `7242880` bytes in a `5m` period.

## Limit Classes

The zrok limits agent includes a concept called _limit classes_. Limit classes can be used to define resource count and bandwidth limits that can be selectively applied to individual accounts in a service instance.

Limit classes are created by creating a record in the `limit_classes` table in the zrok controller database. The table has this schema:

```sql
CREATE TABLE public.limit_classes (
    id integer NOT NULL,
    label VARCHAR(32),
    backend_mode public.backend_mode,
    environments integer DEFAULT '-1'::integer NOT NULL,
    shares integer DEFAULT '-1'::integer NOT NULL,
    reserved_shares integer DEFAULT '-1'::integer NOT NULL,
    unique_names integer DEFAULT '-1'::integer NOT NULL,
    share_frontends integer DEFAULT '-1'::integer NOT NULL,
    period_minutes integer DEFAULT 1440 NOT NULL,
    rx_bytes bigint DEFAULT '-1'::integer NOT NULL,
    tx_bytes bigint DEFAULT '-1'::integer NOT NULL,
    total_bytes bigint DEFAULT '-1'::integer NOT NULL,
    limit_action public.limit_action DEFAULT 'limit'::public.limit_action NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);

```

This schema supports constructing the 3 different types of limits classes that the system supports.

After defining a limit class in the database, it can be applied to specific user accounts (overriding the relevant parts of the global configuration) by inserting a row into the `applied_limit_classes` table:

```sql
CREATE TABLE public.applied_limit_classes (
    id integer NOT NULL,
    account_id integer NOT NULL,
    limit_class_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted boolean DEFAULT false NOT NULL
);
```

Create a row in this table linking the `account_id` to the `limit_class_id` to apply the limit class to a specific user account.

### Unscoped Resource Count Classes

To support overriding the resource count limits defined in the global limits configuration, a site administrator can create a limit class by inserting a row into the `limit_classes` table structured like this:

```sql
insert into limit_classes (environments, shares, reserved_shares, unique_names, share_frontends) values (1, 1, 1, 1, 1);
```

This creates a limit class that sets the `environments`, `shares`, `reserved_shares`, and `unique_names` all to `1`.

When this limit class is applied to a user account those values would override the default resource count values configured globally.

Applying an unscoped resource count class _does not_ affect the bandwidth limits (either globally configured, or via a limit class).

### Unscoped Bandwidth Classes

To support overriding the bandwidth limits defined in the global configuration, a site administrator can create a limit class by inserting a row into the `limit_classes` table structured like this:

```sql
insert into limit_classes (period_minutes, total_bytes, limit_action) values (2, 204800, 'limit');
```

This inserts a limit class that allows for a total bandwidth transfer of `204800` bytes every `2` minutes.

When this limit class is applied to a user account, those values would override the default bandwidth values configured globally.

Applying an unscoped bandwidth class _does not_ affect the resource count limits (either globally configured, or via a limit class).

### Scoped Classes

A scoped limit class specifies _both_ the resource counts (`shares`, `reserved_shares`, and `unique_names`, but *NOT* `environments`) for a *specific* backend mode. Insert a row like this:

```sql
insert into limit_classes (backend_mode, shares, reserved_shares, unique_names, period_minutes, total_bytes, limit_action) values ('web', 2, 1, 1, 2, 4096000, 'limit');
```

Scoped limits are designed to _increase_ the limits for a specific backend mode beyond what the global configuration and the unscoped classes provide. The general approach is to use the global configuration and the unscoped classes to provide the general account limits, and then the scoped classes can be used to further increase (or potentially _decrease_)  the limits for a specific backend mode.

If a scoped limit class exists for a specific backend mode, then the limits agent will use that limit in making a decision about limiting the resource count or bandwidth. All other types of shares will fall back to the unscoped classes or the global configuration.

## Limit Actions

When an account exceeds a bandwidth limit, the limits agent will seek to limit the affected shares (based on the combination of global configuration, unscoped limit classes, and scoped limit classes). It applies the limit by removing the underlying OpenZiti dial policies for any frontends that are trying to access the share.

This means that public frontends will simply return a `404` as if the share is no longer there. Private frontends will also return `404` errors. When the limit is relaxed, the dial policies are put back in place and the share will continue operating normally.

## Unlimited Accounts

The `accounts` table in the database includes a `limitless` column. When this column is set to `true` the account is not subject to any of the limits in the system.

## Experimental Limits Locking

zrok versions prior to `v0.4.31` had a potential race condition when enforcing resource count limits. This usually only manifested in cases where shares or environments were being allocated programmatically (and fast enough to win the limits race). 

This occurs due to a lack of transactional database locking around the limited structures. `v0.4.31` includes a pessimistic locking facility that can be enabled _only_ on the PostgreSQL store implemention.

If you're running PostgreSQL for your service instance and you want to enable the new experimental locking facility that eliminates the potential resource count race condition, add the `enable_locking: true` flag to your `store` definition:

```yaml
store:
  enable_locking: true
```

## Caveats

There are a number of caveats that are important to understand when using the limits agent with more complicated limits scenarios:

### Aggregate Bandwidth

The zrok limits agent is a work in progress. The system currently does not track bandwidth individually for each backend mode type, which means all bandwidth values are aggregated between all of the share types that an account might be using. This will likely change in an upcoming release.

### Administration Through SQL

There are currently no administrative API endpoints (or corresponding CLI tools) to support creating and applying limit classes in the current release. The limits agent infrastructure was designed to support software integrations that directly manipulate the underlying database structures.

A future release may provide API and CLI tooling to support the human administration of the limits agent.

### Performance

Be sure to minimize the number of different periods used for specifying bandwidth limits. Specifying limits in multiple different periods can cause a multiplicity of queries to be executed against the metrics store (InfluxDB). Standardizing on a period like `24h` or `6h` and using that consistently is the best way to to manage the performance of the metrics store.