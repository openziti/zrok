---
sidebar_position: 10
---
# Shares - Reserved

By default a `public` or `private` share is allocated a _share token_ when you run the `zrok share` command. The `zrok share` command is the bridge between your local environment and the users you are sharing with. When you terminate the `zrok share`, the bridge is eliminated and the _share token_ is deleted. If you run `zrok share` again, you will be allocated a brand new _share token_.

You can use a `reserved` share to persist your _share token_ across multiple runs of the `zrok share` bridge. When you use a `reserved` share, the share token will not be deleted between multiple runs of `zrok share`.

To use a `reserved` share, you will first run the `zrok reserve` command to create the reserved share (see `zrok reserve --help` for details). Once you've created your `reserved` share, you will use the `zrok share reserved` command (see `--help` for details) to run the bridge for the shared resource.

This pattern works for both `public` and `private` shares, and for all resource types supported by `zrok`.

To delete your `reserved` share use the `zrok release` command.
