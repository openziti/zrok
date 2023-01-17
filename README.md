![zrok](docs/images/zrok.png)

`zrok` is a next-generation sharing platform built on top of [Ziti](https://docs.openziti.io), a programmable zero-trust network overlay. `zrok` is a _Ziti Native Application_.

`zrok` facilitates sharing resources publicly and privately, exposing them to an audience you can easily control.

Like other competitors in this space, `zrok` allows users to create ephemeral reverse proxies ("tunnels") for `http` resources. Additionally:

* `zrok` allows users to _privately_ share resources with other `zrok` users; in _private_ usage scenarios, your private resources are not exposed to any public endpoints; all communication is securely and privately transported between `zrok` environments
* `zrok` allows sharing other types of resources; rather than just proxying `http` endpoints, `zrok` allows users to easily and rapidly share files and web content
* `zrok` is ready to be extended to easily support many kinds of decentralized resource sharing; `zrok` provides a framework that makes this kind of peer-to-peer resource sharing simple and secure

## Frictionless

You can be up and sharing using the `zrok.io` service in minutes. Here is a synopsis of what's involved.

### First-time Setup

* Download the binary for your platform [here](https://zrok.io/downloads)
* `zrok invite` to create an account with the service
* `zrok enable` to enable your shell environment for sharing with the service

### And then... sharing...

* `zrok share` to share resources immediately, simply and securely

See the [Concepts and Getting Started Guide](docs/v0.3_getting_started/getting_started.md) for a full overview.

## Self-Hosting

`zrok` is designed to scale up to support extremely large service instances. `zrok.io` is run by NetFoundry using the same code base that is available to self-hosted environments.

`zrok` is also designed to scale down to support extremely small deployments. Run an OpenZiti network with `zrok` layered on top of it on a Raspberry Pi!

The single `zrok` binary contains everything you need to operate `zrok` environments and also host your own service instances. Just add a Ziti network and you're up and running.

See the [v0.3 Quick Start](docs/v0.3_quickstart.md) for details on getting your own `zrok` service instance running. This builds on top of the [Ziti Quick Start](https://docs.openziti.io/docs/learn/quickstarts/network/) to have you running a `zrok` service instance in minutes.
