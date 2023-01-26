# Introducing zrok

I'm fortunate that I've had the opportunity to work on many interesting projects throughout my career. I was one of the original developers who broke ground on the [OpenZiti](https://github.com/openziti/ziti) project back in 2017. Most of my work on OpenZiti centered on the [fabric](https://github.com/openziti/fabric), data and control plane design, and designing abstractions that would support a lot of what became the "edge" layers. It's been quite exciting to watch OpenZiti blossom and grow into what it's becoming. 

For the last six months, I've had the opportunity to re-approach the world of zero-trust and next-generation networking from the other side of the stack. Instead of working in the lowest layers of protocols and abstractions, I'm working from the perspective of end users and enabling an amazing end-user experience. I'm excited to introduce you to a new set of tools designed to empower end users at the network edge to seamlessly and transparently share resources. Imagine network sharing that is equally secure and transparent.

This new project is called... `zrok`.

`zrok` focuses on streamlining sharing for both developers and end users alike. `zrok` takes inspiration from several other offerings that focus on streamlining developer endpoint sharing. Starting from that recipe, `zrok` adds powerful capabilities that are made possible by building on the foundation provided by OpenZiti. 

Here are some of the things that make `zrok` different...

## Private Sharing

Most of the offerings in this space allow you to easily create "tunnels" that allow outbound-only access to local HTTP resources without punching any holes in a firewall. These tools make these kinds of tunnels effortless to create; with a single command, you've got a public URL that you can share to allow access to your endpoint.

`zrok` expands on this model by supporting something that we're calling "private sharing". You'll share your resources using a single command, but your resources will be privately shared on an OpenZiti network, where they can be securely accessed with a single `zrok` command by other users.

In this model, no user ever has to enable any inbound access from untrusted users. All network access is handled through a secure, zero-trust overlay network. And to make it even simpler, `zrok` handles all of the control plane management of the overlay network, deeply simplifying the experience. This secure sharing model remains the single-command affair that users have come to expect.

## Files; Repositories; Video... Decentralized

Most of the other offerings in this space are focused on sharing low-level network resources. These tools are often used by developers or operations staff to allow access to a private HTTP endpoint or to facilitate a callback to a private endpoint through a webhook. It's considered table stakes for these tools to do this in a _frictionless_ way.

`zrok` also provides a frictionless experience for sharing these kinds of network resources. However, we're taking it a step further... `zrok` will also make this kind of frictionless, decentralized sharing possible for files, software repositories, video streams, and other kinds of resources we haven't even thought of yet.

Combine this kind of resource sharing with our private sharing model, and you've got the recipe for very powerful decentralized services. Imagine using `zrok` as a decentralized, distributed replacement for large centralized file-sharing platforms. Or use it as a replacement for large, centralized video streaming platforms.

We're still just getting started on building out these aspects of `zrok`. But as of this writing, `zrok` already provides built-in single-command file sharing. Combine this with private sharing and you can see this powerful model in action today.

## zrok.io; Production zrok

[NetFoundry](https://netfoundry.io) is offering [zrok.io](https://zrok.io), a managed `zrok` service instance you can use to try out `zrok` and even run small production workloads. This service is currently in limited beta and is available through an invitation process. Visit [zrok.io](https://zrok.io) for details about requesting an invite.

Once `zrok` and `zrok.io` are out of beta, we'll be opening it up to the public.

`zrok.io` runs on top of the open-source version of `zrok`. We're building out a production environment to make sure we can properly operationalize it, but it's the same code you can run in your own environments.

## Open-Source; Self-Host

`zrok` is committed to being open-source. You've got everything you need to host your own `zrok` instance on top of your own private OpenZiti network. We've even streamlined this process, and we're including a simple [guide](https://github.com/openziti/zrok/blob/main/docs/v0.3_self_hosting_guide.md) to getting this running in minutes, including the OpenZiti portions.

You can [access](https://github.com/openziti/zrok) the open-source version of `zrok` today.

## A Start

I'm really excited about sharing `zrok` with you. As of this writing, we're at `v0.3.0`, and there is still a ton of work to do to get `zrok` to where I know it can go. `zrok` is open-source, and we're going to be developing it in public, just like the rest of the OpenZiti products (check out the [OpenZiti GitHub](https://github.com/openziti)).

Starting with `v0.4`, I'm planning on producing a set of regularly-released "development notebooks", documenting the development process and giving you a look at the work we're doing with `zrok`. I'm also planning on producing a set of videos that work through some of what's involved in building your own tiny version of `zrok` on top of OpenZiti; these will be a great introduction to building a _Ziti Native Application_ from the ground up. These videos will also be a comprehensive look at how `zrok` works.

We'd love your participation in the `zrok` project! You can find us on GitHub at [https://github.com/openziti/zrok](https://github.com/openziti/zrok).