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

## Resource Limits

The `environments` and `shares` values control the number of environments and shares that are allowed per-account. Any limit value can be set to `-1`, which means _unlimited_.

## Bandwidth Limits

The `bandwidth` section is designed to provide a configurable system for controlling the amount of data transfer that can be performed by users of the `zrok` service instance. The bandwidth limits are configurable for each share, environment, and account.

`per_account`, `per_environment`, and `per_share` are all configured the same way:

The `period` specifies the time window for the bandwidth limit. See the documentation for [`time.Duration.ParseDuration`](https://pkg.go.dev/time#ParseDuration) for details about the format used for these durations. If the `period` is set to 5 minutes, then the limits implementation will monitor the send and receive traffic for the resource (share, environment, or account) for the last 5 minutes, and if the amount of data is greater than either the `warning` or the `limit` threshold, action will be taken.

The `rx` value is the number of bytes _received_ by the resource. The `tx` value is the number of bytes _transmitted_ by the resource. And `total` is the combined `rx`+`tx` value.

If the traffic quantity is greater than the `warning` threshold, the user will receive an email notification letting them know that their data transfer size is rising and will eventually be limited (the email details the limit threshold).

If the traffic quantity is greater than the `limit` threshold, the resources will be limited until the traffic in the window (the last 5 minutes in our example) falls back below the `limit` threshold.

### Limit Actions

When a resource is limited, the actions taken differ depending on what kind of resource is being limited.

When a share is limited, the dial service policies for that share are removed. No other action is taken. This means that public frontends will simply return a `404` as if the share is no longer there. Private frontends will also return `404` errors. When the limit is relaxed, the dial policies are put back in place and the share will continue operating normally.

When an environment is limited, all of the shares in that environment become limited, and the user is not able to create new shares in that environment. When the limit is relaxed, all of the share limits are relaxed and the user is again able to add shares to the environment.

When an account is limited, all of the environments in that account become limited (limiting all of the shares), and the user is not able to create new environments or shares. When the limit is relaxed, all of the environments and shares will return to normal operation.

## Unlimited Accounts

The `accounts` table in the database includes a `limitless` column. When this column is set to `true` the account is not subject to any of the limits in the system.