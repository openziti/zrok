---
sidebar_position: 25
---

# Sharing TCP and UDP Servers

`zrok` includes support for sharing low-level TCP and UDP network resources using the `tcpTunnel` and `udpTunnel` backend modes.

As of version `v0.4`, `zrok` supports sharing TCP and UDP network resources using `private` sharing.

To share a raw network resource using `zrok`, you'll want to use the `zrok share private` command from your `enable`-d environment, like this:

```
$ zrok share private --backend-mode tcpTunnel 192.168.9.1:22
```

This will result in a share client starting, which looks like this:

```
╭───────────────────────────────────────────────────────────╮╭────────────────────╮
│ access your share with: zrok access private 5adagwfl888k  ││[PRIVATE][TCPTUNNEL]│
╰───────────────────────────────────────────────────────────╯╰────────────────────╯
╭─────────────────────────────────────────────────────────────────────────────────╮
│                                                                                 │
│                                                                                 │
│                                                                                 │
│                                                                                 │
╰─────────────────────────────────────────────────────────────────────────────────╯
```

Then on the system where you want to access your shared resource (an SSH endpoint in this case), you'll need an `enable`-d `zrok` environment. Run the following command (shown in the banner at the top of the `zrok share` client above):

```
$ zrok access private 5adagwfl888k
```

This will start an `access` client on this system:
```
╭─────────────────────────────────────────────────────────────────────────────────╮
│                      tcp://127.0.0.1:9191 -> 5adagwfl888k                       │
╰─────────────────────────────────────────────────────────────────────────────────╯
╭─────────────────────────────────────────────────────────────────────────────────╮
│                                                                                 │
│                                                                                 │
│                                                                                 │
│                                                                                 │
╰─────────────────────────────────────────────────────────────────────────────────╯
```

The `access` client shows the endpoint at the top where the service can be accessed. In this case, you'll want to connect your SSH client to `127.0.0.1:9191`. We'll just use `nc` (netcat) to access the shared TCP port:
```
$ nc 127.0.0.1 9191
SSH-2.0-OpenSSH_9.2 FreeBSD-openssh-portable-9.2.p1,1
```

And both the `share` client and the `access` client show the traffic:

```
╭──────────────────────────────────────────────────────────╮╭─────────────────────╮
│ access your share with: zrok access private 5adagwfl888k ││[PRIVATE] [TCPTUNNEL]│
╰──────────────────────────────────────────────────────────╯╰─────────────────────╯
╭─────────────────────────────────────────────────────────────────────────────────╮
│Friday, 23-Jun-23 15:33:10 EDT ziti-edge-router                                  │
│connId=2147483648, logical=ziti-                                                 │
│sdk[router=tls:ziti-lx:3022] -> ACCEPT 192.168.9.1:22                            │
│                                                                                 │
│                                                                                 │
╰─────────────────────────────────────────────────────────────────────────────────╯
```

```
╭─────────────────────────────────────────────────────────────────────────────────╮
│                       tcp://127.0.0.1:9191 -> 5adagwfl888k                      │
╰─────────────────────────────────────────────────────────────────────────────────╯
╭─────────────────────────────────────────────────────────────────────────────────╮
│Friday, 23-Jun-23 15:33:10 EDT 127.0.0.1:42312 -> ACCEPT 5adagwfl888k            │
│                                                                                 │
│                                                                                 │
│                                                                                 │
╰─────────────────────────────────────────────────────────────────────────────────╯
```

Exit the `access` client to remove the local access to the shared TCP port. Exit the `share` client to disable further accesses to the shared resource.

For UDP network resources just use the `zrok share private --backend-mode udpTunnel` instead of `tcpTunnel`.
