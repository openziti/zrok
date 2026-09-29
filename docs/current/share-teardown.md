# Share teardown

`teardownShare` (`controller/teardownShare.go`) is the only code that deletes a share row. `unshare`, `disable` and admin `deleteAccount` (through `disableEnvironment`) all call it, once per share, inside the request's transaction. It runs, in order:

1. Name mappings: the share's live `share_name_mappings` rows are soft-deleted, and each auto-allocated (non-reserved) name they point at is soft-deleted with them. Reserved names survive and can be used by the next share.
2. Frontend mappings: every `frontend_mappings` row carrying the share token is deleted, whether or not a dynamic proxy controller is configured. One unbind update is returned per row.
3. Access grants for the share.
4. The share row.
5. OpenZiti objects tagged `zrokShareToken=<token>`. An object already gone counts as deleted; any other failure fails the teardown.

Any failure returns at once and the caller's transaction rolls back, leaving every row for a retry. OpenZiti goes last so that an OpenZiti failure leaves the store untouched.

Frontend mapping updates, binds from share create and unbinds from teardown, are returned to the handler and published only after its `trx.Commit()` returns nil, through `publishMappingUpdates`. A failed publish is logged at error level and not returned: the committed rows are authoritative. The dynamic frontend's periodic `FrontendMappings` pull fetches only rows with a higher id than it holds, so it recovers a lost bind but not a lost unbind; a frontend that missed an unbind keeps the mapping until it reloads its mappings from zero. Without a dynamic proxy controller the rows are still written and nothing is published.

On share create, a failed `frontend_mappings` insert fails the request; the share-create compensation then removes the OpenZiti objects the request created.
