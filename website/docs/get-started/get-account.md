---
sidebar_label: "1. Get an account token"
sidebar_position: 2
---

# Step 1: Get an account token

In this step, you'll create a zrok account and get the account token you need to enable your environment.

## myzrok.io (hosted)

1. Go to [myzrok.io](https://myzrok.io) and sign up for a free account.
2. After signing in, open the [API console](https://api-v2.zrok.io/). Your interface will look like this:

    ![zrok API console, empty](../images/zrok-getting-started-button.png)

3. Click the green **CLICK HERE TO GET STARTED!** button. The getting started wizard opens:

    ![zrok getting started wizard](../images/zrok-getting-started-modal.png)

4. Locate your account token under step 2 of the wizard. It looks like a short alphanumeric string,
   for example `7g3K6gVKikWb`.
5. Copy and save your account token. Treat it like a password—it authenticates your device to the zrok service.

## Self-hosted instance

If you've deployed your own zrok instance, an administrator creates your account with:

```bash
zrok2 admin create account <username> <password>
```

The command outputs your account token. Copy it for use in Step 3.

<div style={{marginBottom: '2rem'}} />
