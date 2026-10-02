# Bandwidth-limit relaxation

The controller checks limited accounts after their bandwidth usage drops below the configured threshold. It restores each share's dial policies and removes the account's bandwidth-limit journal entry only when every applicable share succeeds. Failed accounts retain their entries for a later cycle, while other accounts continue. Policy lookup by deterministic name makes a partial retry safe, and a second cycle over restored policies creates nothing.

The cycle uses one database transaction. A store error ends the cycle because PostgreSQL aborts the transaction after a failed statement.

## Which dial policies a share should have

`controller/limits/dialPolicies.go` computes a share's desired dial policies separately from ensuring them; relax and the repair command share it.

- **Public share frontend policy**, `<envZId>-<shrZId>-dial`, tagged with the share token. Its identity roles are the frontends of every namespace the share's live name mappings point at (mapping, name and namespace all undeleted), de-duplicated, which is how `allocatePublicResources` creates it. `frontend_selection` is the v1 column and v2 never writes it, so it is used only for a v1-created row with no live mappings. A share with neither, or whose namespaces have no frontends, has no frontend policy; its create made none, and relax counts it as done.
- **Private access policy**, `<frontendToken>-<envZId>-<shrZId>-dial`, one per frontend row with `private_share_id`. `access private` does not check the share mode, so these exist for public shares too, and relax restores them for both modes.

## Influx deadline

Each limits-agent Influx query is bounded by `limits.query_timeout` (default `30s`) and by the agent's context, which `Stop` cancels. A query that times out or fails part-way through its result is an error, never zero usage, so it cannot relax an account.

## Repairing shares the journal no longer reaches

Incidents cleared by hand, and cycles that crashed before step 1, left live shares with no dial policy and no journal entry; relax never visits them. `zrok2 admin repair-dial-policies <configPath>` walks every live share, computes its desired dial policies and reports each one missing by account, share token and policy name. It honours the journal: an account with a global `limit` entry, or a `limit` entry whose limit class has no backend mode, is skipped as a whole; a `limit` entry whose class has a backend mode skips that mode's shares. It never deletes.

Production procedure:

1. Run `zrok2 admin repair-dial-policies <configPath>` (a dry run) and review the missing policies and skipped accounts.
2. Run it again with `--apply`. It creates the missing policies through the same find-then-create as relax, continues past individual OpenZiti failures, lists under failed shares any v1-created share whose selected frontend no longer exists (with the frontend it selects), and prints checked, missing, created and failed counts; any failure exits non-zero after the run.
3. Run the dry run again; it should report nothing missing.
