![zrok](docs/images/zrok.png)

# zrok

`zrok` is a next-generation sharing platform built on top of [Ziti][openziti], a programmable zero-trust network overlay. `zrok` is a _Ziti Native Application_.

`zrok` facilitates sharing resources publicly and privately, exposing them to an audience you can easily control.

As of version `v0.3.0`, `zrok` provides users the ability to publicly proxy local `http`/`https` endpoints (similar to other players in this space). Additionally, `zrok` provides the ability to:

* _privately_ share resources with other `zrok` users; in _private_ usage scenarios, your private resources are not exposed to any public endpoints, and all communication is securely and privately transported between `zrok` clients
* use `web` sharing; easily share files with others using a single `zrok` command

## Self-Hostable

`zrok` is fully self-hostable.
