---
sidebar_position: 22
sidebar_label: Permission Modes
---

# Permission Modes

Shares created in zrok `v0.4.26` and newer now include a choice of _permission mode_. 

Shares created with zrok `v0.4.25` and older were created using what is now called the _open permission mode_. Whether _public_ or _private_, these shares can be accessed by any user of the zrok service instance, as long as they know the _share token_ of the share. Effectively shares with the _open permission mode_ are accessible by any user of the zrok service instance.

zrok now supports a _closed permission mode_, which allows for more fine-grained control over which zrok users are allowed to privately access your shares using `zrok access private`.

zrok defaults to continuing to create shares with the _open permission mode_. This will likely change in a future release. We're leaving the default behavior in place to allow users a period of time to get comfortable with the new permission modes.

## Creating a Share with Closed Permission Mode

Adding the `--closed` flag to the `zrok share` or `zrok reserve` commands will create shares using the _closed permission mode_:

```
$ zrok share private --headless --closed -b web .
[   0.066]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok access private 0vzwzodf0c7g
```

By default any environment owned by the account that created the share is _allowed_ to access the new share. But a user trying to access the share from an environment owned by a different account will enounter the following error message:

```
$ zrok access private 0vzwzodf0c7g
[ERROR]: unable to access ([POST /access][401] accessUnauthorized)
```

The `zrok share` and `zrok reserve` commands now include an `--access-grant` flag, which allows you to specify additional zrok accounts that are allowed to access your shares:

```
$ zrok share private --headless --closed --access-grant anotheruser@test.com -b web .
[   0.062]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok access private y6h4at5xvn6o
```

And now `anotheruser@test.com` will be allowed to access the share:

```
$ zrok access private --headless y6h4at5xvn6o
[   0.049]    INFO main.(*accessPrivateCommand).run: allocated frontend 'VyvrJihAOEHD'
[   0.051]    INFO main.(*accessPrivateCommand).run: access the zrok share at the following endpoint: http://127.0.0.1:9191
```

## Adding and Removing Access Grants for Existing Shares

If you've created a share (either reserved or ephemeral) and you forgot to include an access grant, or want to remove an access grant that was mistakenly added, you can use the `zrok modify share` command to make the adjustments:

Create a share:

```
$ zrok share private --headless --closed -b web .
[   0.064]    INFO main.(*sharePrivateCommand).run: allow other to access your share with the following command:
zrok access private s4czjylwk7wa
```

In another shell in the same environment you can execute:

```
$ zrok modify share s4czjylwk7wa --add-access-grant anotheruser@test.com
updated
```

And to remove the grant:

```
$ zrok modify share s4czjylwk7wa --remove-access-grant anotheruser@test.com
updated
```

## Limitations

As of `v0.4.26` there is currently no way to _list_ the current access grants. This will be addressed shortly in a subsequent update.