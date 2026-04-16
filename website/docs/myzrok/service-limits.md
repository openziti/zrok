---
title: Service limits
---

NetFoundry's public zrok instance implements limits based on pricing tier, as well as rate limits to protect the service for all users.

## Limits on shares, environments, or bandwidth

The number of shares, environments, or allowed bandwidth is based on your myzrok subscription.
These limits are defined on the [zrok pricing](https://zrok.io/pricing/) page.
Bandwidth limits are based on a rolling 24-hour window.

:::warning
If you exceed the daily bandwidth of your plan, any running shares are disabled and the zrok API prevents new shares
from being created until bandwidth falls below the 24-hour limit.
:::

## Rate limits for public shares

Public shares are subject to API rate limiting by IP address and by individual share token, to protect the zrok service
so that one user doesn't negatively impact others.

| Scope | Requests per 300 seconds | Average |
|-------|--------------------------|---------|
| Per IP address | 2,000 | 6.66 req/sec |
| Per share (any number of IPs) | 7,500 | 25 req/sec |

The rate limiter allows bursts up to the limit within a shorter timespan. Once exceeded, new requests are blocked until
the rate falls below the 300-second window threshold.
