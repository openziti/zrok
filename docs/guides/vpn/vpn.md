---
sidebar_label: VPN
---

# zrok VPN Guide

zrok VPN backend allows for simple host-to-host VPN setup.

## Operating System Requirements

zrok VPN requires elevated privileges to manage network devices.

### Windows

On Windows, you must run zrok VPN commands as an administrator and install Wintun by placing `wintun.dll` ([download link](https://www.wintun.net/)) in the same directory as the `zrok.exe` executable.

### Linux

On Linux, the simplest way to grant the necessary privileges is to run zrok VPN commands as root. You can enable a separate environment for root by also running `zrok enable` as the root user, or you can prefix the commands like `sudo -E` to allow zrok running as root to use the zrok environment owned by the current user. The minimum privilege is runing zrok VPN commands and the `ip` command with the `NET_ADMIN` kernel capability. The `zrok-share.service` unit has a commented example to grant `NET_ADMIN` as an Ambient Capability.

### macOS

On macOS, you must run zrok VPN commands as root. You can prefix the zrok command with `sudo -E` to allow zrok running as root to use the zrok environment owned by the current user.

## Start the VPN Server

VPN is shared through the `vpn` backend of `zrok` command.

```bash
eugene@hermes $ sudo -E zrok share private --headless --backend-mode vpn
[   0.542]    INFO sdk-golang/ziti.(*listenerManager).createSessionWithBackoff: {session token=[589d443c-f59d-4fc8-8c48-76609b7fb402]} new service session
[   0.705]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok access private 3rq7torslq3n
[   0.705]    INFO zrok/endpoints/vpn.(*Backend).Run: started
```

![VPN share](./vpn-share.png)

`sudo` or equivalent invocation is required because VPN mode needs to create a virtual network device (`tun`)
`-E` option allows `zrok` to find your zrok configuration files (in your `$HOME/.zrok`)

By default `vpn` backend uses subnet `10.122.0.0/16` and assigns `10.122.0.1` to the host that stared VPN share.

Example output from `ifconfig`:

```text
tun0: flags=4305<UP,POINTOPOINT,RUNNING,NOARP,MULTICAST>  mtu 16384
        inet 10.122.0.1  netmask 255.255.0.0  destination 10.122.0.1
        inet6 fe80::705f:24e4:dcfc:a6b2  prefixlen 64  scopeid 0x20<link>
        inet6 fd00:7a72:6f6b::1  prefixlen 64  scopeid 0x0<global>
        unspec 00-00-00-00-00-00-00-00-00-00-00-00-00-00-00-00  txqueuelen 500  (UNSPEC)
        RX packets 0  bytes 0 (0.0 B)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 27  bytes 3236 (3.2 KB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```

Default IP/subnet setting can be overridden by adding `<target>` parameter:

```bash
sudo -E zrok share private --headless --backend-mode vpn 192.168.42.12/24
```

## Reserve a VPN Share Token

As with all backend modes, you can reserve a share token for a VPN share.

```bash
eugene@hermes $ zrok reserve private --backend-mode vpn
[   0.297]    INFO main.(*reserveCommand).run: your reserved share token is 'k77y2cl7jmjl'

eugene@hermes $ sudo -E zrok share reserved k77y2cl7jmjl --headless
[   0.211]    INFO main.(*shareReservedCommand).run: sharing target: '10.122.0.1/16'
[   0.211]    INFO main.(*shareReservedCommand).run: using existing backend target: 10.122.0.1/16
[   0.463]    INFO sdk-golang/ziti.(*listenerManager).createSessionWithBackoff: {session token=[22c5708d-e2f2-41aa-a507-454055f8bfcc]} new service session
[   0.641]    INFO main.(*shareReservedCommand).run: use this command to access your zrok share: 'zrok access private k77y2cl7jmjl'
[
```

## Access the VPN Share

```bash
eugene@calculon % sudo -E zrok access private --headless k77y2cl7jmjl
[   0.201]    INFO main.(*accessPrivateCommand).run: allocated frontend '50B5hloP1s1X'
[   0.662]    INFO main.(*accessPrivateCommand).run: access the zrok share at the following endpoint: VPN:
[   0.662]    INFO main.(*accessPrivateCommand).run: 10.122.0.1 -> CONNECTED Welcome to zrok VPN
[   0.662]    INFO zrok/endpoints/vpn.(*Frontend).Run: connected:Welcome to zrok VPN
```

zrok creates a virtual network device, i.e., a "tun" interface, when you run `zrok access`.

Example output from `ifconfig` run on a VPN client device:

```bash
utun5: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> mtu 1500
        inet 10.122.0.3 --> 10.122.0.1 netmask 0xff000000
        inet6 fe80::ce08:faff:fe8a:7b25%utun5 prefixlen 64 scopeid 0x14
        nd6 options=201<PERFORMNUD,DAD>
```

At this point a VPN tunnel is active between your server and client.
In the example above server is `hermes(10.122.0.1)` and client is `calculon(10.122.0.3)`.
All devices in the VPN can access one another by IP address.

```bash
eugene@calculon ~ % ssh eugene@10.122.0.1
Welcome to Ubuntu 23.10 (GNU/Linux 6.5.0-27-generic x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/pro

0 updates can be applied immediately.

Last login: Tue Apr 16 09:27:13 2024 from 127.0.0.1

eugene@hermes:~$ who am i
eugene   pts/8        2024-04-16 10:04 (10.122.0.3)

eugene@hermes:~$
```

You can also make a reverse(server-to-client) connection:

```bash
eugene@hermes:~$ ssh 10.122.0.3
Last login: Tue Apr 16 09:57:28 2024

eugene@calculon ~ % who am i
eugene           ttys008      Apr 16 10:06 (10.122.0.1)
```
