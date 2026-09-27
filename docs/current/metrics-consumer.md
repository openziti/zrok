# Metrics consumer bounds

The controller consumes metrics with manual acknowledgement and a finite prefetch window. It nacks malformed JSON and unusable events without requeue immediately. A usable event has namespace `fabric.usage`, a numeric `interval_start_utc`, a string `tags.serviceId`, and a `usage` object. A missing `circuit_id` produces a warning and remains acceptable.

Within `usage`, a present `ingress.tx`, `ingress.rx`, `egress.tx`, or `egress.rx` must be numeric; a mistyped counter makes the event invalid. Absent counters remain zero, so intervals containing only ingress or only egress are accepted.

The agent holds the Influx writer as its explicit durable sink. It retries that sink locally with exponential backoff, then nacks without requeue if the attempt or elapsed-time budget runs out. Remaining sinks do not receive an event whose durable write failed. Once the durable sink succeeds, later sinks run without retries; their errors are logged and the delivery is acknowledged.

A panic during per-message processing is logged at error level with the recovered value and up to the first 512 bytes of the message body. The delivery is nacked without requeue and processing continues with the next message, preventing a poison message from repeatedly crashing the controller. Legacy sink calls carry their panics back from the guarded goroutine to this recovery boundary. The shutdown drain remains outside it and leaves deliveries unsettled.

The limits sink runs after Influx. If its queue stays full, it logs and counts the dropped handoff, then returns success so the delivery is acknowledged. The usage remains in Influx for later enforcement. `DroppedEvents()` exposes the limits handoff drop count.

The following optional configuration values use these defaults when unset:

| Configuration path | Default |
| --- | --- |
| `metrics.agent.source.prefetch` | 64 |
| `metrics.agent.retry_attempts` | 5 attempts per enrichment or durable-write operation |
| `metrics.agent.retry_initial_backoff` | `1s` |
| `metrics.agent.retry_max_backoff` | `30s` |
| `metrics.agent.retry_budget` | `2m` shared per message |
| `metrics.influx.write_timeout` | `10s` per attempt |
| `limits.handoff_timeout` | `3s` |

Share-detail enrichment and the durable write share one per-message deadline, including waiting for a store connection. Enrichment has three outcomes:

- A confirmed missing share is recognized as `errNotAShare`, debug-logged, and processing continues without retries or attribution. This is legitimate traffic for another service; the Influx writer skips events without a share token and the delivery is acknowledged.
- A share is attributed and processing continues to the durable sink. Ephemeral environments deliberately have no account: enrichment reports that condition at debug level, retains the share token and environment ID, and continues with account ID zero without retries.
- A store failure is retried with the configured attempts and backoff within the shared deadline. Exhaustion is logged and nacked without requeue, and no sinks run. Treating an unavailable store as proof that the service is unrelated would silently lose real share usage.

The consumer cancels database lookups, active writes, and retry waits at shutdown. It drains events so a source blocked on send can exit, leaving drained deliveries unsettled. A delivery held by the source when its send is cancelled is also left unsettled, as is an active delivery interrupted before durable success. Closing the AMQP connection lets the broker requeue these deliveries for recovery. This differs from failed processing: shutdown is not a processing failure, and nacking these valid messages would discard them on restart.

The built-in sinks accept cancellation through `HandleContext` while preserving the existing `UsageSink.Handle` interface. A sink implementing only `Handle` runs through a single guarded call slot: if it never returns, subsequent calls time out without starting more goroutines.

No failed-processing path requests requeue. A broker policy may route discarded deliveries to a dead-letter queue with its own length and TTL bounds; no such policy is required or configured by the controller. The limits agent's own Influx queries remain without deadlines; bounding them is deferred to the relax stage. A stuck limits query cannot block this consumer beyond its bounded handoff.
