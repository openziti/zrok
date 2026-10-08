# Share teardown

`teardownShare` (`controller/teardownShare.go`) tears a share down. `unshare`, `disable` and admin `deleteAccount` (through `disableEnvironment`) all call it, once per share, inside the request's transaction. Its store half is `releaseShareFromStore`, the only code that deletes a share row; `admin repair-store` calls that half alone for shares stranded in deleted environments (see `repair-store.md`). It runs, in order:

1. Name mappings: the share's live `share_name_mappings` rows are soft-deleted, and each auto-allocated (non-reserved) name they point at is soft-deleted with them. Reserved names survive and can be used by the next share.
2. Frontend mappings: every `frontend_mappings` row carrying the share token is deleted, whether or not a dynamic proxy controller is configured. One unbind update is returned per row.
3. Access frontends: every live `frontends` row whose `private_share_id` is the share, the private accesses to it, is soft-deleted. Their dial policies carry the share token and go in step 6. `unaccess` still removes a single access on its own.
4. Access grants for the share.
5. The share row.
6. OpenZiti objects tagged `zrokShareToken=<token>`. An object already gone counts as deleted; any other failure fails the teardown.

Any failure returns at once and the caller's transaction rolls back, leaving every row for a retry. OpenZiti goes last so that an OpenZiti failure leaves the store untouched.

Frontend mapping updates, binds from share create and unbinds from teardown, are returned to the handler and published only after its `trx.Commit()` returns nil, through `publishMappingUpdates`. A failed publish is logged at error level and not returned: the committed rows are authoritative. A lost bind or unbind is recovered by the dynamic frontend's periodic full reconciliation; see `dynamic-proxy-mappings.md`. Without a dynamic proxy controller the rows are still written and nothing is published.

## Account delete

Admin `deleteAccount` disables each of the account's environments, which tears down every share in them, and then, before it deletes the account row, releases the account's names (`releaseAccountNames`): every live name of the account, reserved or not, is soft-deleted. By then no live share of the account can hold one. A mapping on one of its names to a live share (some other account's share) is a store error and fails the delete, rolling everything back; a severed mapping, to a share already deleted, is soft-deleted with its name.

## Share create

On share create, a failed `frontend_mappings` insert fails the request; the share-create compensation then removes the OpenZiti objects the request created.

## Name availability

A name whose share has been deleted can still carry a live `share_name_mappings` row, or a `frontend_mappings` row, written before `teardownShare` released names on every path. `healSeveredName` (`controller/healSeveredName.go`) is the one availability check for a name; share create (for a user-specified name), `createShareName` and `deleteShareName` all call it, inside the request's transaction. It reads the name's live mappings with their share state, and the name's frontend mappings on the namespace's dynamic frontends:

- Any mapping whose share is live, or a frontend mapping with no share row at all, is a conflict carrying the holder's token. Nothing is modified, and the handler answers 409 `name '<name>' in namespace '<ns>' is in use by share '<token>'; run 'zrok2 delete share <token>' to release it`.
- Otherwise every mapping is dead: each share name mapping is soft-deleted, each frontend mapping is removed, and each is logged at info with the dead share's token. The request then proceeds as if the name had been free.

Each frontend mapping removed yields an unbind, returned like `teardownShare`'s and published only after the handler's commit: `createShareName` and `deleteShareName` publish them on their own, and share create puts them ahead of its binds in the one list it publishes, so a frontend sees the unbind and then the bind for the same name. In share create, a conflict (a namespace or name that does not exist, a name the account does not own, a namespace it is not granted, a live holder) is a `nameSelectionConflict` and answers 409; any other error from the name check, a store failure in the heal included, answers 500.

The healing commits or rolls back with the request; a request that fails later leaves the severed rows for the next attempt to heal again. It is a repair of past damage, not a teardown path: it touches no share row and no OpenZiti object. The auto-allocated branch of share create (no name given) creates a fresh name and does not call it.
