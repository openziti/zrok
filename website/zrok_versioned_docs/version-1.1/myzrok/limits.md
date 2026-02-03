---
title: Limits
---

NetFoundry's public zrok instance implements various limits based on pricing tier,
as well as rate limits in order to protect the service for all users.

### Limits on Shares, Environments, or Bandwidth

The number of shares, enviroments, or allowed bandwidth is based on the limits outlined within your myzrok subscription.
These limits are defined on the [zrok pricing](https://zrok.io/pricing/) page.
Bandwidth limitations are based on a rolling 24 hour window. Note that if you exceed the daily bandwidth of your plan,
any running shares will be disabled, and the zrok API will prevent any new shares from being created until the bandwidth
falls back below the 24 hour limit.

### Rate Limitations For Public Shares
Public shares are subject to API rate limiting, both by IP address, as well as the individual share token.
These limits exist to protect the zrok service so that one user does not negatively impact the experience for others.
The rate limits for public shares are defined below:

#### Per IP Address
2000 requests per 300 seconds (average of 6.66 requests per second)

The rate limiter will allow a burst of requests in a shorter timespan up to 2000 requests, but once the rate limit has been exceeded,
new requests will be blocked until the request rate falls below the limit of the 300 second window.

#### Per Share
7500 requests per 300 seconds from *any number of IP addresses* (average of 25 requests per second)




