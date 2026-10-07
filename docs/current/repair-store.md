# Store repair

`zrok2 admin repair-store <configPath> [--apply] [--batch 500]` (`cmd/zrok2/adminRepairStore.go`, `controller/repairStore.go`) runs as its own process against the controller's store and releases the store rows a teardown would have released had it run: rows left by shares deleted before `teardownShare` released names on every path, and by teardowns that never ran or stopped part-way. It opens the store with auto-migration off, as `admin gc` and `admin repair-dial-policies` do (`controller/adminStore.go`): a store behind the binary's schema fails the command rather than being migrated by it. `healSeveredName` repairs a name when someone asks for it (see `share-teardown.md`); this sweep reaches the rest, which no request will ever visit.

## What it repairs

Each condition is one a teardown would have handled had it run. They run in this order.

- **A.** A live environment of a deleted account, and a live share in a deleted environment (or in a live environment of a deleted account, which A is about to delete). The environment goes through `removeEnvironmentFromStore`, the store half of `disable`: its own frontends and the environment row are soft-deleted. The share goes through `releaseShareFromStore`, the store half of `teardownShare`: its name mappings and allocated names, frontend mappings, access frontends, access grants and the share row. Environments go first, a batch of environments at a time, then shares. This is store-side only. It makes no OpenZiti calls: once the share row is deleted its token is no longer live, and the next `admin gc` run collects its objects.
- **B.** A live access frontend (`frontends.private_share_id` set) whose share has no live row. Soft-deleted. Shares torn down before step 13 left these, because `teardownShare` removed their dial policies but not the rows.
- **C.** A live name of a deleted account. Soft-deleted, with any live mapping on it soft-deleted first. Account deletes before step 13 left these.
- **D.** A live auto-allocated name that no live mapping holds. Soft-deleted.

Then the mapping conditions:

1. A live `share_name_mappings` row whose share has no live row. The mapping is soft-deleted; if its name is auto-allocated (not reserved) and still live, the name is soft-deleted too, as `teardownShare` does. Reserved names survive. A mapping whose name is also deleted counts here, not under 2.
2. A live `share_name_mappings` row of a live share whose name is deleted. The availability check looks mappings up by the live name's id, so it never sees these. Soft-deleted.
3. A `frontend_mappings` row whose share token belongs to no live share: the share is deleted, or no share row carries the token at all. The availability check treats the second kind as a live holder and never heals it; only this sweep removes it. Deleted; the table has no `deleted` column.

A row can match more than one condition when the survey runs: an allocated name of a deleted account held by a stranded share is counted under A's shares and under C. The earlier condition repairs it, and the later one finds it already done, so its repaired count is lower than its found count. The survey's found counts are therefore not additive.

A share token is unique only among live shares, so condition 3 asks whether any live share carries the token, not whether a row with the token is deleted.

## Modes

Without `--apply` it is a dry run. One transaction, rolled back, counts each condition (A split into environments and shares, 1 into reserved and auto-allocated), counts the auto-allocated names condition 1 would release, and lists up to twenty rows of each: the row's table and id, then whichever it has of name, namespace token (frontend token for condition 3 and B, OpenZiti id for an environment), share token, and the environment or account it belongs to, marking a frontend mapping whose token has no share row. The report goes to stdout and the totals to the log.

With `--apply` the survey is printed and then each condition is repaired in order: A environments, A shares, B, C, D, 1 reserved, 1 auto-allocated, 2, 3. A goes first because a released share takes its name mappings, frontend mappings and access frontends with it. Among the mapping conditions, reserved names go first so the names users chose are freed first if a run is interrupted. Each batch is its own transaction: a plain `select ... order by id limit N` of the matching rows, then an update or delete by those ids, in SQL both engines accept. Progress is logged after each batch. A failed batch rolls back, is reported, and stops the run with a non-zero exit; the batches before it stay committed. Repaired rows no longer match, so a re-run resumes where the last one stopped. The report ends with found and repaired per condition, the names condition 1 released, batches run and failed batches.

## No notifications

The command touches the store only. It publishes no unbinds: dynamic frontends drop the removed frontend mappings at their next periodic reconciliation (`dynamic-proxy-mappings.md`). It runs outside the controller, which owns the publisher, and a frontend mapping it removes serves no live share, so nothing is lost while the frontend waits.

## What is left in OpenZiti

The command makes no OpenZiti calls; it is store-side by decision. The OpenZiti side of the same debt is `admin gc`'s (`garbage-collection.md`), and that covers the objects of shares released under A: once the share row is deleted its token is no longer live, and the next gc run collects them.

gc does not cover environments. The identity and edge router policy of an environment soft-deleted under A stay in OpenZiti. After condition A, the apply report lists those identities' ids under a heading saying they are left in OpenZiti, for removal one at a time with `zrok2 admin delete identity <zId>`, which deletes the edge router policy named after the identity and then the identity. Only environments in committed batches are listed. The dial policies of the frontends in those environments are tagged with share tokens that may still be live, so gc keeps them; they also stay.

## Production procedure

1. Run the dry run and keep its output.
2. Review the sample. Live shares are expected under "shares in deleted environments", whose environments are gone, and under "mappings to deleted names", whose names are gone. Under every other condition each share token should be a share you expect to be gone, and no live share should appear. Under "allocated names with no mapping", check that the names are not ones their owners un-reserved on purpose (see below).
3. Run with `--apply`. Rows severed since the dry run, if any, are repaired too; the counts are the dry run's, the repaired column is what was done.
4. Run the dry run again; it should find nothing.
5. Remove the identities the apply report listed as left in OpenZiti, with `zrok2 admin delete identity <zId>`, then run `admin gc` to collect the released shares' objects.

## Unmapped allocated names

`updateShareName` lets an owner set `reserved` to false on a name no share holds. That leaves a live, auto-allocated name with no mapping, which is exactly condition D. A run of the sweep releases such a name, as the owner's next teardown of a share on it would. The production snapshot had three, which may be this path rather than a leak.
