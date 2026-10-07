# Garbage collection

`zrok2 admin gc <configPath> [--delete] [--min-age 24h]` (`cmd/zrok2/adminGc.go`, `controller/gc.go`) runs as its own process against the controller's store and OpenZiti controller, and removes OpenZiti objects left behind by shares that no longer exist. It opens the store with auto-migration off (`controller/adminStore.go`), so even a dry run never changes the schema; migrations belong to the controller and `zrok2 admin migrate`.

## What gc owns

gc examines services, configs, service policies and service edge router policies tagged `zrok`. An object belongs to a share when it carries a `zrokShareToken` tag:

- token belongs to a live (not deleted) row in `shares`: live, kept;
- token belongs to no live row: orphaned, a candidate;
- no `zrokShareToken` tag: not share-owned, never touched.

Names are never consulted. The last group covers agent-remote objects (`zrokAgentRemote`), frontend identities and their policies, the dynamic proxy controller's service, and anything else carrying only the `zrok` tag. An access-path dial policy carries the share token as well as `zrokFrontendToken`, so it follows its share. Identities and edge router policies belong to environments and are handled by `disable`, not gc. Live shares come from the shared `shares` table, so the v1 line's shares count as live too.

## Age guard

A candidate whose `createdAt` is younger than `--min-age` (default 24 hours) is skipped and counted as too young. A share create allocates its OpenZiti objects before its row commits, so for the length of that request its objects look orphaned; the guard keeps gc off them, and a later run picks up a real orphan. The store is read before OpenZiti is listed, so an object created after the read is younger than the run.

The default is a day rather than minutes for two reasons. `createdAt` is stamped by the OpenZiti controller's clock and compared against the clock of the host running gc, and a day is well clear of any skew between the two. And since v2.0.6 stopped the leak, the orphans left to reclaim are old; nothing younger than a day needs collecting, so the longer guard costs nothing. A shorter `--min-age` is for a deliberate cleanup, such as on a dev environment, where the clocks are known to agree.

## Modes

Every kind is listed in full, 500 at a time with an advancing offset (`automation.FindAll`). The report goes to stdout in both modes: for each kind the counts of live, orphaned, too young and not share-owned objects, then the orphans' ids and names grouped by share token. The JSON log names every orphan and too-young object.

Without `--delete` it is a dry run and nothing is deleted. With `--delete` the orphans are deleted by id in the order service edge router policies, service policies, configs, services, so a service is never left referenced by a policy or config that outlives it. An object already gone counts as collected; any other failure is logged and the run continues, and the command exits with an error naming the number of failed deletes. A table of deleted, already gone and failed counts per kind follows the report.

## Production procedure

1. Run the dry run and keep its output.
2. Review it: every live share's objects should be counted live, the orphans listed by token should be shares you expect to be gone, and nothing without a share token appears in the orphan list.
3. Run with `--delete`. Objects created since the dry run are either live or too young.
4. Run the dry run again; it should report no orphans.
