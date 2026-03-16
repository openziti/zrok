---
sidebar_label: Accounts and environments
sidebar_position: 4
---

# Accounts and environments

zrok organizes access around two foundational concepts: your account on a zrok instance, and the environment you
create on each machine you use zrok from.

## Instance and account

zrok is hosted software. You interact with a zrok *instance*, and your account on that instance is identified by a
username and password, which you use to sign in to the web console. Your account also has a *secret token*, which
you use to authenticate from the zrok command line.

You create a new account with NetFoundry's zrok instance by subscribing at [myzrok.io](https://myzrok.io), or in a
self-hosted instance by running `zrok2 invite` or `zrok2 admin create account`.

## Environment

Using your secret token, you use the zrok command line to create an *environment*. An environment corresponds to a
single command-line user on a specific host system.

You create a new environment with `zrok2 enable`. Each machine you use zrok from needs its own environment.
