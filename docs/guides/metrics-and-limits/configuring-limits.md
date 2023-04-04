# Configuring Limits

The limits facility in `zrok` is responsible for controlling the number of resources in use (environments, shares) and also for ensuring that any single account, environment, or share is held below the configured thresholds.

Take this `zrok` controller configuration stanza as an example:

```yaml
limits:
  enforcing:        true
  cycle:            1m
  environments:     -1
  shares:           -1
  bandwidth:
    per_account:
      period:       5m
      warning:
        rx:         -1
        tx:         -1
        total:      7242880
      limit:
        rx:         -1
        tx:         -1
        total:      10485760
    per_environment:
      period:       5m
      warning:
        rx:         -1
        tx:         -1
        total:      -1
      limit:
        rx:         -1
        tx:         -1
        total:      -1
    per_share:
      period:       5m
      warning:
        rx:         -1
        tx:         -1
        total:      -1
      limit:
        rx:         -1
        tx:         -1
        total:      -1
```

## The Global Controls

The `enforcing` boolean will globally enable or disable limits for the controller.

The `cycle` value controls how frequently the limits system will look for limited resources to re-enable.

## Resource Counts

The `environments` and `shares` values control the number of environments and shares that are allowed per-account. Any limit value can be set to `-1`, which means _unlimited_.