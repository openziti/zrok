# Configuring Metrics

A fully configured, production-scale `zrok` service instance looks like this:

![zrok Metrics Architecture](images/metrics-architecture.png)

`zrok` metrics builds on top of the `fabric.usage` event type from OpenZiti. The OpenZiti controller has a number of way to emit events. The `zrok` controller has several ways to consume `fabric.usage` events. Smaller installations could be configured in these ways:

![zrok simplified metrics architecture](images/metrics-architecture-simple.png)

Environments that horizontally scale the `zrok` control plane with multiple controllers should use an AMQP-based queue to "fan out" the metrics workload across the entire control plane. Simpler installations that use a single `zrok` controller can collect `fabric.usage` events from the OpenZiti controller by "tailing" the events log file, or collecting them from the OpenZiti controller's websocket implementation.