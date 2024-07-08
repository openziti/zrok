---
title: Personalized Frontend
sidebar_label: Personalized Frontend
sidebar_position: 19
---

This guide describes an approach for self-hosting _only_ the components required to manage a public frontend, complete with TLS and customized DNS, for one or many zrok private shares.

This approach gives you complete control over the way that your shares are accessed publicly, and can be self-hosted on an extremely minimal VPS instance or through a container hosting service.

This approach works for both HTTPS shares, and also for TCP and UDP ports, allowing you to put all of these things onto the public internet, while maintaining strong security for your protected resources.

This guide isn't a detailed _how to_ with specific steps to follow. This is more of a description of the overall concept. You'll want to figure out your own specific steps to implement this style of deployment in your own environment.

## Overview

You've got a handful of private resources that you want to share using zrok. Some of them are HTTP or HTTPS resources. Maybe you also have a couple of raw TCP or UDP ports that you want to expose to the internet.

Let's assume you're using the shared zrok service instance at zrok.io. The deployed solution is going to look like this:

![personalized-frontend-1](../../images/personalized-frontend-1.png)

In our example, we'll be using `zrok share private` to share an endpoint we'll call `A` and a second `zrok share private` to share an endpoint we'll call `B`. These two shares could both be from a single environment on a single system, or they might be distributed anywhere in the world. You could use a single share, or you could use hundreds.

Because we're using `private` zrok shares, they'll need to be accessed using a corresponding `zrok access` private command. The `zrok access private` command creates a "network listener" where the share can be accessed. You can use `zrok access private` to access a `private` zrok share from as many places as you want (up to the limit configuration of the service).

Imagine that both of our shares are HTTP shares. We want our `A` share to be accessed at the URL `a.example.com` and we want our `B` share to be accessed at the URL `b.example.com`. This assumes that we have control over `example.com` such that we can create the necessary DNS entries.

What we'll do is acquire a cheap VPS instance. Even a $5/month (or less!) VPS instance is very likely up to the job, depending on the amount of traffic that we want to send to our shares.

On our VPS instance, we'll run a `zrok access private` for both share `A` and another for share `B`, binding both of those to a port on the `localhost` loopback interface.

With those 2 `access` processes running and bound to loopback ports, we'll install a reverse proxy server like nginx or Caddy, and configure it to support both `a.example.com` and `b.example.com`, directing the traffic for each to the correct loopback port.

Add DNS and TLS to the VPS and you've got a fully-featured public-facing frontend that you fully control, with a minimum amount of components to self host.

zrok in this case becomes the backbone, isolating your private resources from the public internet. There is no access to your protected infrastructure that does not flow through the zrok network.

## TCP and UDP

This approach also works very well for TCP and UDP ports that you might want to publicly expose. Simply add additional `zrok access private` instances, binding those to the correct interface on your system. With the appropriate firewall rules, those ports will be fully exposed to the internet, while your protected resources are still fully private and unreachable from the public internet.

## Privacy

When you use a public frontend (with a simple `zrok share public`) at a hosted zrok instance (like zrok.io), the operators of that service have some amount of visibility into what traffic you're sending to your shares. The load balancers in front of the public frontend maintain logs describing all of the URLs that were accessed, as well as other information (headers, etc.) that contain information about the resource you're sharing.

If you create private shares using `zrok share private` and then run your own `zrok access private` from some other location, the operators of the zrok service instance only know that some amount of data moved between the environment running the `zrok share private` and the `zrok access private`. There is no other information available.