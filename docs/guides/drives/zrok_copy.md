# The zrok Drives CLI

The zrok Drives CLI tools allow for simple, ergonomic management and synchronization of local and remote file objects transparently.

## Sharing a Drive

Virtual drives are shared through the `zrok` CLI using the `--backend-mode drive` flag with the `zrok share` command, using either the `public` or `private` sharing modes:

```
$ mkdir /tmp/junk
$ zrok share private --headless --backend-mode drive /tmp/junk
[   0.124]    INFO sdk-golang/ziti.(*listenerManager).createSessionWithBackoff: {session token=[cf640aac-2706-49ae-9cc9-9a497d67d9c5]} new service session
[   0.145]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok access private wkcfb58vj51l
```

The command shown above creates an ephemeral `zrok` drive share pointed at the local `/tmp/junk` folder.

Notice that the share token allocated by `zrok` is `wkcfb58vj51l`. We'll use that share token to identify our virtual drive in the following operations.

## Working with the Drive Share

First, let's copy a file into our virtual drive using the `zrok copy` command:

```
$ zrok copy LICENSE zrok://wkcfb58vj51l
[   0.119]    INFO zrok/drives/sync.OneWay: => /LICENSE
copy complete!
```

We used the URL scheme `zrok://<shareToken>` to refer to the private virtual drive we allocated above using the `zrok share private` command. Use `zrok://` URLs with the drives CLI tools to refer to contents of private virtual drives.

Next, let's get a directory listing of the virtual drive:

```
$ zrok ls zrok://wkcfb58vj51l
┌──────┬─────────┬─────────┬───────────────────────────────┐
│ TYPE │ NAME    │ SIZE    │ MODIFIED                      │
├──────┼─────────┼─────────┼───────────────────────────────┤
│      │ LICENSE │ 11.3 kB │ 2024-01-19 12:16:46 -0500 EST │
└──────┴─────────┴─────────┴───────────────────────────────┘
```

We can make directories on the virtual drive:

```
$ zrok mkdir zrok://wkcfb58vj51l/stuff
$ zrok ls zrok://wkcfb58vj51l
┌──────┬─────────┬─────────┬───────────────────────────────┐
│ TYPE │ NAME    │ SIZE    │ MODIFIED                      │
├──────┼─────────┼─────────┼───────────────────────────────┤
│      │ LICENSE │ 11.3 kB │ 2024-01-19 12:16:46 -0500 EST │
│ DIR  │ stuff   │         │                               │
└──────┴─────────┴─────────┴───────────────────────────────┘
```

We can copy the contents of a local directory into the new directory on the virtual drive:

```
$ ls -l util/
total 20
-rw-rw-r-- 1 michael michael 329 Jul 21 13:17 email.go
-rw-rw-r-- 1 michael michael 456 Jul 21 13:17 headers.go
-rw-rw-r-- 1 michael michael 609 Jul 21 13:17 proxy.go
-rw-rw-r-- 1 michael michael 361 Jul 21 13:17 size.go
-rw-rw-r-- 1 michael michael 423 Jan  2 11:57 uniqueName.go
$ zrok copy util/ zrok://wkcfb58vj51l/stuff
[   0.123]    INFO zrok/drives/sync.OneWay: => /email.go
[   0.194]    INFO zrok/drives/sync.OneWay: => /headers.go
[   0.267]    INFO zrok/drives/sync.OneWay: => /proxy.go
[   0.337]    INFO zrok/drives/sync.OneWay: => /size.go
[   0.408]    INFO zrok/drives/sync.OneWay: => /uniqueName.go
copy complete!
$ zrok ls zrok://wkcfb58vj51l/stuff
┌──────┬───────────────┬───────┬───────────────────────────────┐
│ TYPE │ NAME          │ SIZE  │ MODIFIED                      │
├──────┼───────────────┼───────┼───────────────────────────────┤
│      │ email.go      │ 329 B │ 2024-01-19 12:26:45 -0500 EST │
│      │ headers.go    │ 456 B │ 2024-01-19 12:26:45 -0500 EST │
│      │ proxy.go      │ 609 B │ 2024-01-19 12:26:45 -0500 EST │
│      │ size.go       │ 361 B │ 2024-01-19 12:26:45 -0500 EST │
│      │ uniqueName.go │ 423 B │ 2024-01-19 12:26:45 -0500 EST │
└──────┴───────────────┴───────┴───────────────────────────────┘
```

And we can remove files and directories from the virtual drive:

```
$ zrok rm zrok://wkcfb58vj51l/LICENSE
michael@fourtyfour Fri Jan 19 12:29:12 ~/Repos/nf/zrok 
$ zrok ls zrok://wkcfb58vj51l
┌──────┬───────┬──────┬──────────┐
│ TYPE │ NAME  │ SIZE │ MODIFIED │
├──────┼───────┼──────┼──────────┤
│ DIR  │ stuff │      │          │
└──────┴───────┴──────┴──────────┘
michael@fourtyfour Fri Jan 19 12:29:14 ~/Repos/nf/zrok 
$ zrok rm zrok://wkcfb58vj51l/stuff
michael@fourtyfour Fri Jan 19 12:29:20 ~/Repos/nf/zrok 
$ zrok ls zrok://wkcfb58vj51l
┌──────┬──────┬──────┬──────────┐
│ TYPE │ NAME │ SIZE │ MODIFIED │
├──────┼──────┼──────┼──────────┤
└──────┴──────┴──────┴──────────┘
```

