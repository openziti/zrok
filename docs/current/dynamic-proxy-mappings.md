# Dynamic proxy mappings

The dynamic frontend (`zrok2 access dynamicProxy`, `endpoints/dynamicProxy/`) serves requests from an in-memory map of hostname to share token. The controller's `frontend_mappings` rows are authoritative; the map learns them three ways, all applied by one loop in `mappings.go`:

| Source | When | Can learn | Cannot learn |
| --- | --- | --- | --- |
| Full pull (`FrontendMappings`, id 0) | at start, then every `mapping_reconcile_interval` (default `10m`) | additions, removals, replacements | nothing, but it is tens of thousands of rows for a busy frontend, so it runs on minutes |
| Delta pull (id above the highest held) | every `mapping_refresh_interval` (default `5m`) | every row with a higher id than any held: additions, and replacements, since a name deleted and re-inserted gets a higher id | deletions: a deleted row never appears in a higher-id pull |
| AMQP bind and unbind (`amqpSubscriber.go`) | as the controller publishes, after commit | additions and removals, immediately | anything published while the broker or the frontend's connection was down |

The delta pull never removes a name, by design, and stays that way; removals are the subscriber's job and, when an update is lost, the reconciliation's.

## Start

The initial full pull retries with backoff, one second doubling to thirty, until it succeeds or the frontend stops, logging each failure. Until it succeeds the map is empty and requests answer not-found.

## Reconciliation

On each tick the frontend pulls the complete set and, under the map's lock, makes the map match it:

- a name absent from the set is dropped;
- a name missing from the map is added;
- a name whose share token or id differs from the set's is replaced with the set's row;
- a name whose row is identical is left alone.

Each difference is logged at info with the name and both tokens, and one summary line gives the counts and elapsed time. A failed pull is logged at error and the map is left untouched.

A complete set that comes back empty while the map is non-empty is logged at warn and not applied the first time. For a frontend carrying traffic a single empty answer is more likely a failure the client did not see than a world with no shares, and applying it would drop every mapping at once. If the very next reconciliation is also empty it is applied like any other set, each drop and the summary logged as usual, so a frontend whose last mappings really were removed stops serving them within two intervals. A non-empty set or a failed pull in between resets the count.

## Subscriber acknowledgement

The subscriber consumes its own queue, exclusive and deleted with the process, with manual acknowledgement and a prefetch of `amqp_subscriber.prefetch` (default 64).

| Delivery | Settlement |
| --- | --- |
| parsed bind or unbind, handed to the loop | ack |
| parsed, but the handoff channel (`amqp_subscriber.queue_depth`) is full | ack, logged at warn; reconciliation covers it |
| body that does not parse, or an unknown operation | nack without requeue, logged at error with the body's first hundred bytes; not forwarded |
| in flight when the frontend stops | nack without requeue |

Nothing requeues. The queue has no other consumer and dies with the process, so a requeued message can only return to the same handler, and one it cannot process would return forever. A message lost with the queue is what reconciliation recovers.
