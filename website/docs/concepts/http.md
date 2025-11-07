---
sidebar_position: 22
---

# Sharing HTTP Servers

`zrok` can share HTTP and HTTPS resources natively. If you have an existing web server that you want to share with other users, you can use the `zrok share` command using the `--backend-mode proxy` flag.

The `--backend-mode proxy` is the default backend mode, so if you do not specify a `--backend-mode` you will get the `proxy` mode by default.

If you have a web server running on `localhost` that you want to expose to other users using `zrok`, you can execute a command like the following:

```
$ zrok share public localhost:8080
```
When you execute this command, you'll get a `zrok` bridge like the following:

```
╭───────────────────────────────────────────────────────────────╮╭────────────────╮
│               http://cht7gj4g5pjf.share.zrok.io               ││[PUBLIC] [PROXY]│
╰───────────────────────────────────────────────────────────────╯╰────────────────╯
╭─────────────────────────────────────────────────────────────────────────────────╮
│                                                                                 │
│                                                                                 │
│                                                                                 │
│                                                                                 │
╰─────────────────────────────────────────────────────────────────────────────────╯
```

The URL shown at the top of the bridge shows the address where you can access your `public` share.

Hit `CTRL-C` or `q` in the bridge to exit it and delete the `public` share.
