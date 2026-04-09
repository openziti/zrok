---
sidebar_position: 20
---

# Configure metrics

Configure the OpenZiti controller, metrics bridge, and zrok controller to collect and store usage metrics in InfluxDB.

## Configure the OpenZiti controller

1. Add the following stanza to the OpenZiti controller configuration to append `fabric.usage` events to a file:

    ```yaml
    events:
      jsonLogger:
        subscriptions:
          - type: fabric.usage
            version: 3
        handler:
          type: file
          format: json
          path: /tmp/fabric-usage.json
    ```

    Adjust `events/jsonLogger/handler/path` to wherever you want to send these events for ingestion into zrok. Consult the OpenZiti docs for additional options that control file rotation.

2. Add the following to the `network` stanza of the OpenZiti controller configuration to increase the reporting frequency. By default, the OpenZiti events infrastructure reports and batches events in 1-minute buckets — too large an interval for a responsive zrok metrics experience. This increases the frequency to every 5 seconds:

    ```yaml
    network:
      intervalAgeThreshold: 5s
      metricsReportInterval: 5s
    ```

3. Add the following stanza to the router configuration for every router on your OpenZiti network:

    ```yaml
    metrics:
      reportInterval: 5s
      intervalAgeThreshold: 5s
    ```

4. Restart all components of your OpenZiti network for the configuration changes to take effect.

## Configure the zrok metrics bridge

zrok uses a metrics bridge component (running as a separate process) to consume `fabric.usage` events from the OpenZiti controller and publish them onto an AMQP queue.

1. Add the following stanza to your zrok controller configuration:

    ```yaml
    bridge:
      source:
        type:           fileSource
        path:           /tmp/fabric-usage.json
      sink:
        type:           amqpSink
        url:            amqp://guest:guest@localhost:5672
        queue_name:     events
    ```

    This consumes `fabric.usage` events from the file specified in the OpenZiti controller configuration and publishes them onto an AMQP queue.

2. Start RabbitMQ as your AMQP implementation. The default RabbitMQ configuration works as a Docker container:

    ```bash
    docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.11-management
    ```

3. Start the zrok metrics bridge by pointing it at your zrok controller configuration:

    ```bash
    zrok2 controller metrics bridge <path/to/zrok-controller.yaml>
    ```

## Configure zrok metrics

1. Add the following `metrics` section to your zrok controller configuration:

    ```yaml
    metrics:
      agent:
        source:
          type:         amqpSource
          url:          amqp://guest:guest@localhost:5672
          queue_name:   events
      influx:
        url:            "http://127.0.0.1:8086"
        bucket:         zrok  # the bucket and org must be
        org:            zrok  # created in advance in InfluxDB
        token:          "<secret token>"
    ```

    This configures the zrok controller to consume usage events from the AMQP queue and write them to InfluxDB. The InfluxDB organization and bucket must be created in advance — the zrok controller will not create them for you.

## Test metrics

With all components configured and running, use `zrok test loop` or manually create shares to generate traffic on the zrok instance. If everything is working correctly, log messages from the controller will look like this:

```text
[5339.658]    INFO zrok/controller/metrics.(*influxWriter).Handle: share: 736z80mr4syu, circuit: Ad1V-6y48 backend {rx: 4.5 kB, tx: 4.6 kB} frontend {rx: 4.6 kB, tx: 4.5 kB}
[5349.652]    INFO zrok/controller/metrics.(*influxWriter).Handle: share: 736z80mr4syu, circuit: Ad1V-6y48 backend {rx: 2.5 kB, tx: 2.6 kB} frontend {rx: 2.6 kB, tx: 2.5 kB}
[5354.657]    INFO zrok/controller/metrics.(*influxWriter).Handle: share: 5a4u7lqxb7pa, circuit: iG1--6H4S backend {rx: 13.2 kB, tx: 13.3 kB} frontend {rx: 13.3 kB, tx: 13.2 kB}
```

The zrok web console should also show activity for your shares:

![zrok web console activity](images/zrok-console-activity.png)

With metrics configured, you might be interested in [limits](configuring-limits.md).
