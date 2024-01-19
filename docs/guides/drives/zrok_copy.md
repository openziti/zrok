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

