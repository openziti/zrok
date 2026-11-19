# Bandwidth-limit relaxation

The controller checks limited accounts after their bandwidth usage drops below the configured threshold. It restores each share's dial policies and removes the account's bandwidth-limit journal entry only when every applicable share succeeds. Failed accounts retain their entries for a later cycle, while other accounts continue. Policy lookup by deterministic name makes a partial retry safe, and a second cycle over restored policies creates nothing.

The cycle uses one database transaction. A store error ends the cycle because PostgreSQL aborts the transaction after a failed statement.

## Which dial policies a share should have

`controller/limits/dialPolicies.go` computes a share's desired dial policies separately from ensuring them; relax and the repair command share it.

- **Public share frontend policy**, `<envZId>-<shrZId>-dial`, tagged with the share token. Its identity roles are the frontends of every namespace the share's live name mappings point at (mapping, name and namespace all undeleted), de-duplicated, which is how `allocatePublicResources` creates it. `frontend_selection` is the v1 column and v2 never writes it, so it is used only for a v1-created row with no live mappings. A share with neither, or whose namespaces have no frontends, has no frontend policy; its create made none, and relax counts it as done.
- **Private access policy**, `<frontendToken>-<envZId>-<shrZId>-dial`, one per frontend row with `private_share_id`. `access private` does not check the share mode, so these exist for public shares too, and relax restores them for both modes.

## Influx deadline

Each limits-agent Influx query is bounded by `limits.query_timeout` (default `30s`) and by the agent's context, which `Stop` cancels. A query that times out or fails part-way through its result is an error, never zero usage, so it cannot relax an account.

## Enforcement fast path

The limits agent enforces each usage event that has a share token. Before it runs enforcement it asks two in-process caches whether the event could require anything; when they show it cannot, the event is finished without a transaction and without an Influx query.

- **Account cache**, by account: whether the account is limitless and its bandwidth classes, trusted for `limits.account_cache_lifetime`. Beside them it keeps, per class, when the journal last confirmed a crossing of that class.
- **Usage cache**, by account and period: the received and transmitted totals Influx returned, trusted for `limits.usage_cache_lifetime`.

The fast path finishes an event when the account is limitless, or when the class enforcement would select for the event's backend mode, which is the last exceeded class in that mode's array, is not exceeded by the cached totals, or has a crossing recorded. A recorded crossing of an earlier class never stands in for a later one, so a recorded warning does not hide a limit crossed since. An event with no account, or with no backend mode, always takes the full path.

Only the store and Influx decide anything that writes. When a cache has no fresh entry, or when it shows the selected class exceeded and unrecorded, enforcement runs unchanged: its own transaction, account, share row (for the backend mode), limit classes, Influx totals and journal. No cached value is passed into it. What a successful run read refreshes the caches; a failed or panicking run refreshes nothing. A stale cache can make enforcement late, never wrong.

A crossing record is trusted for one account lifetime from its own confirmation, whether the journal entry already existed or was just created. Refreshing the account does not extend it, and a run that confirms one class leaves the other classes' records as they were. So a relaxed account that crosses the same class again is journaled within the account lifetime, however busy its other traffic is.

| Change | Takes effect within |
| --- | --- |
| An administrator applies, removes or changes an account's limit classes, or its limitless flag | `limits.account_cache_lifetime` |
| An account crosses a warning or limit threshold | `limits.usage_cache_lifetime` |
| An account is relaxed and crosses the same class again | `limits.account_cache_lifetime` |

The caches belong to the agent's run goroutine, are per controller process, take no lock and are never invalidated; an administrative change is seen when entries expire. Each map is bounded in code: when it is full, expired entries are dropped, and if none have expired it is cleared. The API's limit checks and the relax cycle do not use the caches.

| Configuration path | Default |
| --- | --- |
| `limits.account_cache_lifetime` | `60s` |
| `limits.usage_cache_lifetime` | `60s` |
| `limits.cache_report_interval` | `5m` |

Each value is resolved where it is used: unset or zero means the default, and a negative value turns that cache, or the report line, off. With either cache off the fast path never finishes an event for an account that is not limitless; with the account cache off it never finishes any, which is the behaviour before the caches existed. The crossing records share the account lifetime and are off when it is.

After each event, once `limits.cache_report_interval` has passed and events were handled since the last report, the agent logs one info line with the events the fast path finished and the events that took the full path, then resets the counts.

## Repairing shares the journal no longer reaches

Incidents cleared by hand, and cycles that crashed before step 1, left live shares with no dial policy and no journal entry; relax never visits them. `zrok2 admin repair-dial-policies <configPath>` walks every live share, computes its desired dial policies and reports each one missing by account, share token and policy name. It honours the journal: an account with a global `limit` entry, or a `limit` entry whose limit class has no backend mode, is skipped as a whole; a `limit` entry whose class has a backend mode skips that mode's shares. It never deletes.

Production procedure:

1. Run `zrok2 admin repair-dial-policies <configPath>` (a dry run) and review the missing policies and skipped accounts.
2. Run it again with `--apply`. It creates the missing policies through the same find-then-create as relax, continues past individual OpenZiti failures, lists under failed shares any v1-created share whose selected frontend no longer exists (with the frontend it selects), and prints checked, missing, created and failed counts; any failure exits non-zero after the run.
3. Run the dry run again; it should report nothing missing.
