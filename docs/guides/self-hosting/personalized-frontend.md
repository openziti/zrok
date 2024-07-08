---
title: Personalized Frontend
sidebar_label: Personalized Frontend
sidebar_position: 19
---

This guide describes an approach for self-hosting _only_ the components required to manage a public frontend, complete with TLS and customized DNS, for one or many zrok private shares.

This approach gives you complete control over the way that your shares are accessed publicly, and can be self-hosted on an extremely minimal VPS instance or through a container hosting service.

We're going to explore this approach using a minimal VPS through this guide.

## Overview

The approach looks like this:

![personalized-frontend-1](../../images/personalized-frontend-1.png)