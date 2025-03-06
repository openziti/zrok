# zrok Pastebin

This example shows the use of the zrok SDK spinning up a simple pastebin command.

## Self-hosting Setup :wrench:

You don't need this section if you're using hosted zrok from NetFoundry (https://api-v1.zrok.io/).

Refer to the [setup guide](../../../docs/guides/self-hosting/self_hosting_guide.md) for details on setting up your zrok
environment if you're self-hosting zrok.

### Install Python Requirements

The zrok SDK requires Python 3.10 or later.

If you haven't already installed them, you'll need the dependent libraries used in the examples.

```bash
pip install -r ./sdk/python/examples/pastebin/requirements.txt
```

## Running the Example :arrow_forward:

This example contains a `copyto` server portion and `pastefrom` client portion.

### copyto

The server portion expects to get the data you want to send via stdin. It can be invoked by:

```bash
echo "this is a cool test" | python pastebin.py copyto
```

You should see some helpful info printed out to your terminal:

```bash
access your pastebin using 'pastebin.py pastefrom vp0xgmknvisu'
```

The last token in that line is your share token. We'll use that in the pastefrom command to access our data.

### pastefrom

The `pastefrom` client expects the share token as an argument.
If we envoke it using the same token as above:

```bash
python pastebin.py pastefrom vp0xgmknvisu
```

we see the data we had piped into the `copyto` server:

```text
this is a cool test
```
