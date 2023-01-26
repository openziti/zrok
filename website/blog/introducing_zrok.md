# Introducing zrok

I'm fortunate that I've had the opportunity to work on many interesting projects throughout my career. I was one of the original developers who broke ground on the OpenZiti project back in 2017. I had the opportunity to work deep in the core design for OpenZiti, and a lot of the core ideas came from work I did directly.

For the last six months, I've had the opportunity to come at the world of Zero Trust and next generation networking from the other end. I've been working on a set of user-focused tools that aim to streamline sharing by making the network both secure, and invisible.

This new project is called... `zrok`.

`zrok` focuses on streamlining sharing for both developers and end users alike. `zrok` takes inspiration from a number of other offerings that focus on streamlining developer endpoint sharing. Starting from that recipe, `zrok` adds a number of powerful capabilities that are made possible by building on the foundation provided by OpenZiti. 

Here are some of the things that make `zrok` different...

## Private Sharing

Most of the offerings in this space allow you to easily create "tunnels" that allow outbound-only access to local HTTP resources without punching any holes in a firewall. These tools make these kinds of tunnels effortless to create; a single command and you've got a public URL that you can share to allow access to your endpoint.

`zrok` expands on this model by supporting something that we're calling "private sharing". You'll share your resources using a single command, but your resources will be privately shared on an OpenZiti network, where they can be securely accessed with a single `zrok` command by other users.

In this model, nobody ever has to enable any inbound access from untrusted users. All network access is handled through a secure, zero trust overlay network. And to make it even simpler, `zrok` handles all of the control plane management of the overlay network. This secure sharing model remains the single-command affair that users have come to expect.

## Files; Repositories; Video... Decentralized

Most of the other offerings in this space have focused on sharing network resources. These tools are often used by developers to allow local access to a private HTTP endpoint, or to facilitate a callback to a private endpoint through a webhook. It's considered table stakes for these tools to do this in a way that is _frictionless_.

`zrok` also provides a frictionless experience for sharing these kinds of network resources. However, we're taking it a step further, though... `zrok` will also make this kind of frictionless, decentralized sharing possible for files, software repositories, video streams, and a number of other kinds of resources we haven't even thought of yet.

Combine this kind of resource sharing with our private sharing model, and you've got the recipe for a number of very powerful decentralized services. Imagine using `zrok` as a decentralized, distributed replacement for large centralized file sharing platforms. Use it as a replacement for large, centralized video streaming platforms.

We're still just getting started on building out these aspects of `zrok`. But as of this writing, `zrok` already provides built-in single-command file sharing. Combine that with private sharing and you can see this powerful model in action right now.

## zrok.io

NetFoundry is offering `zrok.io`, a managed service instance you can use to try out `zrok` and even run small production workloads. This service is currently in limited beta and is available through an invitation process until we're out of beta. Visit [zrok.io](https://zrok.io) for details about requesting an invite.

`zrok.io` runs on top of the open source version of `zrok`. We've built out some scaffolding to make sure we can properly operationalize it, but it's the same code you can run in your own environments.

## Open Source; Self-Host

`zrok` is open source. You've got everything you need to host your own `zrok` instance on top of your own private OpenZiti network. We've even streamlined this process, and we're including a simple [guide](https://github.com/openziti/zrok/blob/main/docs/v0.3_self_hosting_guide.md) to getting this running in minutes, including the OpenZiti portions.

## A Start

I'm really excited about sharing `zrok` with you. As of this writing, we're at `v0.3.0`, and there is still a ton of work to do to get `zrok` to where I know it can go. `zrok` is open source, and we're going to be developing it in public, just like the rest of the OpenZiti products.

We'd love your participation! You can find us on Github at [https://github.com/openziti/zrok](https://github.com/openziti/zrok).