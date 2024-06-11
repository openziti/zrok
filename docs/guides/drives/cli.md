# The Drives CLI

The zrok drives CLI tools allow for simple, ergonomic management and synchronization of local and remote files.

## Sharing a Drive

Virtual drives are shared through the `zrok` CLI using the `--backend-mode drive` flag through the `zrok share` command, using either the `public` or `private` sharing modes. We'll use the `private` sharing mode for this example:

```
$ mkdir /tmp/junk
$ zrok share private --headless --backend-mode drive /tmp/junk
[   0.124]    INFO sdk-golang/ziti.(*listenerManager).createSessionWithBackoff: {session token=[cf640aac-2706-49ae-9cc9-9a497d67d9c5]} new service session
[   0.145]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok access private wkcfb58vj51l
```

The command shown above creates an ephemeral, `private` drive share pointed at the local `/tmp/junk` folder.

Notice that the share token allocated by `zrok` is `wkcfb58vj51l`. We'll use that share token to identify our virtual drive in the following operations.

## Working with a Private Drive Share

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
$ zrok ls zrok://wkcfb58vj51l
┌──────┬───────┬──────┬──────────┐
│ TYPE │ NAME  │ SIZE │ MODIFIED │
├──────┼───────┼──────┼──────────┤
│ DIR  │ stuff │      │          │
└──────┴───────┴──────┴──────────┘
$ zrok rm zrok://wkcfb58vj51l/stuff
$ zrok ls zrok://wkcfb58vj51l
┌──────┬──────┬──────┬──────────┐
│ TYPE │ NAME │ SIZE │ MODIFIED │
├──────┼──────┼──────┼──────────┤
└──────┴──────┴──────┴──────────┘
```

## Working with Public Shares

Public shares work very similarly to private shares, they just use a different URL scheme:

```
$ zrok share public --headless --backend-mode drive /tmp/junk
[   0.708]    INFO sdk-golang/ziti.(*listenerManager).createSessionWithBackoff: {session token=[05e0f48b-242b-4fd9-8edb-259488535c47]} new service session
[   0.878]    INFO main.(*sharePublicCommand).run: access your zrok share at the following endpoints:
 https://6kiww4bn7iok.share.zrok.io
```

The same commands, with a different URL scheme work with the `zrok` drives CLI:

```
$ zrok copy util/ https://6kiww4bn7iok.share.zrok.io
[   0.268]    INFO zrok/drives/sync.OneWay: => /email.go
[   0.406]    INFO zrok/drives/sync.OneWay: => /headers.go
[   0.530]    INFO zrok/drives/sync.OneWay: => /proxy.go
[   0.655]    INFO zrok/drives/sync.OneWay: => /size.go
[   0.714]    INFO zrok/drives/sync.OneWay: => /uniqueName.go
copy complete!
michael@fourtyfour Fri Jan 19 12:42:52 ~/Repos/nf/zrok 
$ zrok ls https://6kiww4bn7iok.share.zrok.io
┌──────┬───────────────┬───────┬───────────────────────────────┐
│ TYPE │ NAME          │ SIZE  │ MODIFIED                      │
├──────┼───────────────┼───────┼───────────────────────────────┤
│      │ email.go      │ 329 B │ 2023-07-21 13:17:56 -0400 EDT │
│      │ headers.go    │ 456 B │ 2023-07-21 13:17:56 -0400 EDT │
│      │ proxy.go      │ 609 B │ 2023-07-21 13:17:56 -0400 EDT │
│      │ size.go       │ 361 B │ 2023-07-21 13:17:56 -0400 EDT │
│      │ uniqueName.go │ 423 B │ 2024-01-02 11:57:14 -0500 EST │
└──────┴───────────────┴───────┴───────────────────────────────┘
```

For basic authentication provided by public shares, the `zrok` drives CLI offers the `--basic-auth` flag, which accepts a `<username>:<password>` parameter to specify the authentication for the public virtual drive (if it's required).

Alternatively, the authentication can be set using the `ZROK_DRIVES_BASIC_AUTH` environment variable:

```
$ export ZROK_DRIVES_BASIC_AUTH=username:password
```

## One-way Synchronization

The `zrok copy` command includes a `--sync` flag, which only copies files detected as _modified_. `zrok` considers a file with the same modification timestamp and size to be the same. Of course, this is not a strong guarantee that the files are equivalent. Future `zrok` drives versions will provide a cryptographically strong mechanism (a-la `rsync` and friends) to guarantee that files and trees of files are synchronized.

For now, the `--sync` flag provides a convenience mechanism to allow resuming copies of large file trees and provide a reasonable guarantee that the trees are in sync.

Let's take a look at `zrok copy --sync` in action:

```
$ zrok copy --sync docs/ https://glmv049c62p7.share.zrok.io
[   0.636]    INFO zrok/drives/sync.OneWay: => /_attic/
[   0.760]    INFO zrok/drives/sync.OneWay: => /_attic/network/
[   0.816]    INFO zrok/drives/sync.OneWay: => /_attic/network/_category_.json
[   0.928]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/
[   0.987]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/ziti-ctrl.service
[   1.048]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/ziti-ctrl.yml
[   1.107]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/ziti-router0.service
[   1.167]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/ziti-router0.yml
[   1.218]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/zrok-access-public.service
[   1.273]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/zrok-ctrl.service
[   1.328]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/zrok-ctrl.yml
[   1.382]    INFO zrok/drives/sync.OneWay: => /_attic/network/prod/zrok.io-network-skeleton.md
[   1.447]    INFO zrok/drives/sync.OneWay: => /_attic/overview.md
[   1.572]    INFO zrok/drives/sync.OneWay: => /_attic/sharing/
[   1.622]    INFO zrok/drives/sync.OneWay: => /_attic/sharing/_category_.json
[   1.673]    INFO zrok/drives/sync.OneWay: => /_attic/sharing/reserved_services.md
[   1.737]    INFO zrok/drives/sync.OneWay: => /_attic/sharing/sharing_modes.md
[   1.793]    INFO zrok/drives/sync.OneWay: => /_attic/v0.2_account_requests.md
[   1.902]    INFO zrok/drives/sync.OneWay: => /_attic/v0.4_limits.md
...
[   9.691]    INFO zrok/drives/sync.OneWay: => /images/zrok_web_ui_empty_shares.png
[   9.812]    INFO zrok/drives/sync.OneWay: => /images/zrok_web_ui_new_environment.png
[   9.870]    INFO zrok/drives/sync.OneWay: => /images/zrok_zoom_to_fit.png
copy complete!
```

Because the target drive was empty, `zrok copy --sync` copied the entire contents of the local `docs/` tree into the virtual drive. However, if we run that command again, we get:

```
$ zrok copy --sync docs/ https://glmv049c62p7.share.zrok.io
copy complete!
```

The virtual drive contents are already in sync with the local filesystem tree, so there is nothing for it to copy.

Let's alter the contents of the drive and run the `--sync` again:

```
$ zrok rm https://glmv049c62p7.share.zrok.io/images
$ zrok copy --sync docs/ https://glmv049c62p7.share.zrok.io
[   0.364]    INFO zrok/drives/sync.OneWay: => /images/
[   0.456]    INFO zrok/drives/sync.OneWay: => /images/zrok.png
[   0.795]    INFO zrok/drives/sync.OneWay: => /images/zrok_cover.png
[   0.866]    INFO zrok/drives/sync.OneWay: => /images/zrok_deployment.drawio
...
[   2.254]    INFO zrok/drives/sync.OneWay: => /images/zrok_web_ui_empty_shares.png
[   2.340]    INFO zrok/drives/sync.OneWay: => /images/zrok_web_ui_new_environment.png
[   2.391]    INFO zrok/drives/sync.OneWay: => /images/zrok_zoom_to_fit.png
copy complete!
```

Because we removed the `images/` tree from the virtual drive, `zrok copy --sync` detected this and copied the local `images/` tree back onto the virtual drive.

## Drive-to-Drive Copies and Synchronization

The `zrok copy` CLI can operate on pairs of virtual drives remotely, without ever having to store files locally. This allow for drive-to-drive copies and synchronization.

Here are a couple of examples:

```
$ zrok copy --sync https://glmv049c62p7.share.zrok.io https://glmv049c62p7.share.zrok.io
copy complete!
```

Specifying the same URL for both the source and the target of a `--sync` operation should always result in nothing being copied... they are the same drive with the same state.

We can copy files between two virtual drives with a single command:

```
$ zrok copy --sync https://glmv049c62p7.share.zrok.io zrok://hsml272j3xzf
[   1.396]    INFO zrok/drives/sync.OneWay: => /_attic/
[   2.083]    INFO zrok/drives/sync.OneWay: => /_attic/overview.md
[   2.704]    INFO zrok/drives/sync.OneWay: => /_attic/sharing/
...
[ 118.240]    INFO zrok/drives/sync.OneWay: => /images/zrok_web_console_empty.png
[ 118.920]    INFO zrok/drives/sync.OneWay: => /images/zrok_enable_modal.png
[ 119.589]    INFO zrok/drives/sync.OneWay: => /images/zrok_cover.png
[ 120.214]    INFO zrok/drives/sync.OneWay: => /getting-started.mdx
copy complete!
$ zrok copy --sync https://glmv049c62p7.share.zrok.io zrok://hsml272j3xzf
copy complete!
```

## Copying from Drives to the Local Filesystem

In the current version of the drives CLI, `zrok copy` always assumes the destination is a directory. There is currently no way to do:

```
$ zrok copy somefile someotherfile
```

What you'll end up with on the local filesystem is:

```
somefile
someotherfile/somefile
```

It's in the backlog to support file destinations in a future release of `zrok`. So, when using `zrok copy`, always take note of the destination.

`zrok copy` supports a default destination of `file://.`, so you can do single parameter `zrok copy` commands like this:

```
$ zrok ls https://azc47r3cwjds.share.zrok.io
┌──────┬─────────┬─────────┬───────────────────────────────┐
│ TYPE │ NAME    │ SIZE    │ MODIFIED                      │
├──────┼─────────┼─────────┼───────────────────────────────┤
│      │ LICENSE │ 11.3 kB │ 2023-07-21 13:17:56 -0400 EDT │
└──────┴─────────┴─────────┴───────────────────────────────┘
$ zrok copy https://azc47r3cwjds.share.zrok.io/LICENSE
[   0.260]    INFO zrok/drives/sync.OneWay: => /LICENSE
copy complete!
$ ls -l
total 12
-rw-rw-r-- 1 michael michael 11346 Jan 19 13:29 LICENSE
```

You can also specify a local folder as the destination for your copy:

```
$ zrok copy https://azc47r3cwjds.share.zrok.io/LICENSE /tmp/inbox
[   0.221]    INFO zrok/drives/sync.OneWay: => /LICENSE
copy complete! 
$ l /tmp/inbox
total 12
-rw-rw-r-- 1 michael michael 11346 Jan 19 13:30 LICENSE
```

## Unique Names and Reserved Shares

Private reserved shares with unque names can be particularly useful with the drives CLI:

```
$ zrok reserve private -b drive --unique-name mydrive /tmp/junk
[   0.315]    INFO main.(*reserveCommand).run: your reserved share token is 'mydrive'
$ zrok share reserved --headless mydrive
[   0.289]    INFO main.(*shareReservedCommand).run: sharing target: '/tmp/junk'
[   0.289]    INFO main.(*shareReservedCommand).run: using existing backend target: /tmp/junk
[   0.767]    INFO sdk-golang/ziti.(*listenerManager).createSessionWithBackoff: {session token=[d519a436-9fb5-4207-afd5-7cbc28fb779a]} new service session
[   0.927]    INFO main.(*shareReservedCommand).run: use this command to access your zrok share: 'zrok access private mydrive'
```

This makes working with `zrok://` URLs particularly convenient:

```
$ zrok ls zrok://mydrive
┌──────┬─────────┬─────────┬───────────────────────────────┐
│ TYPE │ NAME    │ SIZE    │ MODIFIED                      │
├──────┼─────────┼─────────┼───────────────────────────────┤
│      │ LICENSE │ 11.3 kB │ 2023-07-21 13:17:56 -0400 EDT │
└──────┴─────────┴─────────┴───────────────────────────────┘
```

## Future Enhancements

Coming in a future release of `zrok` drives are features like:

* two-way synchronization between multiple hosts... allowing for shared "dropbox-like" usage scenarios between multiple environments
* better ergonomics for single-file destinations
