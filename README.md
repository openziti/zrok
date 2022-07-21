# zrok

`zrok` is a utility for quickly proving public access to dark, private applications. 

![zrok overview](docs/images/zrok_overview.png)

`zrok` is designed with the following properties:

## Easiest Possible On-ramp

`zrok` is the fastest, simplest path for exposing dark, private applications onto the public internet using Ziti.

### Simple Registration

Registering for access to `zrok` should provide the user with a single identity token, which can be used from any shell environment to quickly enable access to private applications.

Enabling `zrok` in a shell should be as simple as executing something like:

```
$ zrok enable <token>
```

### Single-Executable Deployment

A registered user should only need a single executable (`zrok`), along with their identity, to enable `zrok` capabilities in any shell environment.

### URLs that Don't Change

The smallest, simplest `zrok` implementation could be capable of providing URLs that don't change. The competition does not offer this capability without a subscription.

## Expand into Ziti

The `zrok` implementation should (ideally) be such that `zrok` usage patterns can co-exist with larger, more featureful Ziti implementations. Ideally, a developer who started with `zrok` should have patterns that allow them to incrementally expand their usage.

## Multiple Isolated Tenants

A single `zrok` implementation should support multiple isolated tenants coexisting on the same deployment (and underlying Ziti network) in a secure manner.

## Self-hosting Capable

The `zrok` implementation should support self-hosting, such that existing Ziti users can easily add `zrok` capabilities to their existing networks.
