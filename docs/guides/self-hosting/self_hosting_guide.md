---
sidebar_position: 40
sidebar_label: Linux VPS
---

# Self-Hosting Guide for Linux

## Walkthrough Video

<iframe width="100%" height="315" src="https://www.youtube.com/embed/870A5dke_u4" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Before you Begin

This will get you up and running with a self-hosted instance of zrok. I'll assume you have the following:

* a Linux server with a public IP
* a wildcard DNS record like `*.zrok.quigley.com` that resolves to the server IP

## OpenZiti Quickstart

The first step is to log in to your Linux server and run the OpenZiti quickstart. This will install a Ziti controller and Ziti router as systemd services.

I specifically used the "Host OpenZiti Anywhere" variant because it provides a public controller. We'll need that to use zrok with multiple devices across different networks.

Keep track of the generated admin password when running the `expressInstall` script. The script will prompt you like this:

```
Do you want to keep the generated admin password 'XO0xHp75uuyeireO2xmmVlK91T7B9fpD'? (Y/n)
```

You'll need that generated password (`XO0xHp75uuyeireO2xmmVlK91T7B9fpD`) when building your `zrok` controller configuration.

BEGIN: [Run the OpenZiti Quickstart](https://docs.openziti.io/docs/learn/quickstarts/network/hosted)

## Install zrok

Download [the latest release](https://github.com/openziti/zrok/releases/latest) from GitHub.

## Configure the Controller

Create a controller configuration file in `etc/ctrl.yml`. The controller does not provide server TLS, but you may front the server with a reverse proxy. This example will expose the non-TLS listener for the controller.

```yaml
#    _____ __ ___ | | __
#   |_  / '__/ _ \| |/ /
#    / /| | | (_) |   <
#   /___|_|  \___/|_|\_\
# controller configuration

v:                  3

admin:
  secrets:
    -               f60b55fa-4dec-4c4a-9244-e3b7d6b9bb13

endpoint:
  host:             0.0.0.0
  port:             18080

store:
  path:             zrok.db
  type:             sqlite3

ziti:
  api_endpoint:     "https://127.0.0.1:1280"
  username:         admin
  password:         "XO0xHp75uuyeireO2xmmVlK91T7B9fpD"

```

The `admin` section defines privileged administrative credentials and must be set in the `ZROK_ADMIN_TOKEN` environment variable in shells where you want to run `zrok admin`.

The `endpoint` section defines where your `zrok` controller will listen. 

The `store` section defines the local `sqlite3` database used by the controller.

The `ziti` section defines how the `zrok` controller should communicate with your OpenZiti installation. When using the OpenZiti quickstart, an administrative password will be generated; the `password` in the `ziti` stanza should reflect this password.

## Environment Variables

The `zrok` binaries are configured to work with the global `zrok.io` service, and default to using `api.zrok.io` as the endpoint for communicating with the service.

To work with a self-hosted `zrok` deployment, you'll need to set the `ZROK_API_ENDPOINT` environment variable to point to the address where your `zrok` controller will be listening, according to `endpoint` in the configuration file above.

In my case, I've set:

```bash
export ZROK_API_ENDPOINT=http://localhost:18080
```

## Bootstrap OpenZiti for zrok

With your OpenZiti network running and your configuration saved to a local file (I refer to mine as `etc/ctrl.yml` in these examples), you're ready to bootstrap the Ziti network.

Use the `zrok admin bootstrap` command to bootstrap like this:

```bash
$ zrok admin bootstrap etc/ctrl.yml 
[   0.002]    INFO main.(*adminBootstrap).run: {
	...
}
[   0.002]    INFO zrok/controller/store.Open: database connected
[   0.006]    INFO zrok/controller/store.(*Store).migrate: applied 0 migrations
[   0.006]    INFO zrok/controller.Bootstrap: connecting to the ziti edge management api
[   0.039]    INFO zrok/controller.Bootstrap: creating identity for controller ziti access
[   0.071]    INFO zrok/controller.Bootstrap: controller identity: jKd8AINSz
[   0.082]    INFO zrok/controller.assertIdentity: asserted identity 'jKd8AINSz'
[   0.085]    INFO zrok/controller.assertErpForIdentity: asserted erps for 'ctrl' (jKd8AINSz)
[   0.085]    INFO zrok/controller.Bootstrap: creating identity for frontend ziti access
[   0.118]    INFO zrok/controller.Bootstrap: frontend identity: sqJRAINSiB
[   0.119]    INFO zrok/controller.assertIdentity: asserted identity 'sqJRAINSiB'
[   0.120]    INFO zrok/controller.assertErpForIdentity: asserted erps for 'frontend' (sqJRAINSiB)
[   0.120] WARNING zrok/controller.Bootstrap: missing public frontend for ziti id 'sqJRAINSiB'; please use 'zrok admin create frontend sqJRAINSiB public https://{token}.your.dns.name' to create a frontend instance
[   0.123]    INFO zrok/controller.assertZrokProxyConfigType: found 'zrok.proxy.v1' config type with id '33CyjNbIepkXHN5VzGDA8L'
[   0.124]    INFO zrok/controller.assertMetricsService: creating 'metrics' service
[   0.126]    INFO zrok/controller.assertMetricsService: asserted 'metrics' service (5RpPZZ7T8bZf1ENjwGiPc3)
[   0.128]    INFO zrok/controller.assertMetricsSerp: creating 'metrics' serp
[   0.130]    INFO zrok/controller.assertMetricsSerp: asserted 'metrics' serp
[   0.134]    INFO zrok/controller.assertCtrlMetricsBind: creating 'ctrl-metrics-bind' service policy
[   0.135]    INFO zrok/controller.assertCtrlMetricsBind: asserted 'ctrl-metrics-bind' service policy
[   0.138]    INFO zrok/controller.assertFrontendMetricsDial: creating 'frontend-metrics-dial' service policy
[   0.140]    INFO zrok/controller.assertFrontendMetricsDial: asserted 'frontend-metrics-dial' service policy
[   0.140]    INFO main.(*adminBootstrap).run: bootstrap complete!
```

The `zrok admin bootstrap` command configures the `zrok` database, the necessary OpenZiti identities, and all of the OpenZiti policies required to run a `zrok` service.

Notice this warning:

```
[   0.120] WARNING zrok/controller.Bootstrap: missing public frontend for ziti id 'sqJRAINSiB'; please use 'zrok admin create frontend sqJRAINSiB public https://{token}.your.dns.name' to create a frontend instance
```

## Run zrok Controller

The `zrok` bootstrap process wants us to create a "public frontend" for our service. `zrok` uses public frontends to allow users to specify where they would like public traffic to ingress from.

The `zrok admin create frontend` command requires a running `zrok` controller, so let's start that up first:

```bash
$ zrok controller etc/ctrl.yml 
[   0.003]    INFO main.(*controllerCommand).run: {
	...
}
[   0.016]    INFO zrok/controller.inspectZiti: inspecting ziti controller configuration
[   0.048]    INFO zrok/controller.findZrokProxyConfigType: found 'zrok.proxy.v1' config type with id '33CyjNbIepkXHN5VzGDA8L'
[   0.048]    INFO zrok/controller/store.Open: database connected
[   0.048]    INFO zrok/controller/store.(*Store).migrate: applied 0 migrations
[   0.049]    INFO zrok/controller.(*metricsAgent).run: starting
[   0.064]    INFO zrok/rest_server_zrok.setupGlobalMiddleware: configuring
[   0.064]    INFO zrok/ui.StaticBuilder: building
[   0.065]    INFO zrok/rest_server_zrok.(*Server).Logf: Serving zrok at http://[::]:18080
[   0.085]    INFO zrok/controller.(*metricsAgent).listen: started
```

## Create zrok Frontend

With our `ZROK_ADMIN_TOKEN` and `ZROK_API_ENDPOINT` environment variables set, we can create our public frontend like this:

```bash
$ zrok admin create frontend sqJRAINSiB public http://{token}.zrok.quigley.com:8080
[   0.037]    INFO main.(*adminCreateFrontendCommand).run: created global public frontend 'WEirJNHVlcW9'
```

The id of the frontend was emitted earlier in by the zrok controller when we ran the bootstrap command. If you don't have that log message the you can find the id again with the `ziti` CLI like this:

```bash
# initialize the Ziti quickstart env
source ~/.ziti/quickstart/$(hostname -s)/$(hostname -s).env
# login as admin
zitiLogin
# list Ziti identities created by the quickstart and bootstrap
ziti edge list identities
```

The id is shown for the "frontend" identity. 

Nice work! The `zrok` controller is fully configured now that you have created the zrok frontend.

## Configure the Public Frontend

Create `etc/http-frontend.yml`. This frontend config file has a `host_match` pattern that represents the DNS zone you're using with this instance of zrok. Incoming HTTP requests with a matching `Host` header will be handled by this frontend. You may also specify the interface address where the frontend will listen for public access requests.

The frontend does not provide server TLS, but you may front the server with a reverse proxy. It is essential the reverse proxy forwards the `Host` header supplied by the viewer. This example will expose the non-TLS listener for the frontend.

```yaml
host_match: zrok.quigley.com
address: 0.0.0.0:8080
```

## Start Public Frontend

In another terminal window, run:

```bash
$ zrok access public etc/http-frontend.yml
[   0.002]    INFO main.(*accessPublicCommand).run: {
	...
}
[   0.002]    INFO zrok/endpoints/public_frontend.newMetricsAgent: loaded 'frontend' identity
```

This process uses the `frontend` identity created during the bootstrap process to provide public access for the `zrok` deployment. It is expected that the configured listener for this `frontend` corresponds to the DNS template specified when creating the public frontend record above.

## Invite Yourself

```bash
$ zrok invite
New Email: user@domain.com
Confirm Email: user@domain.com
invitation sent to 'user@domain.com'!
```

If you look at the console output from your `zrok` controller, you'll see a message like this:

```
[ 238.168]    INFO zrok/controller.(*inviteHandler).Handle: account request for 'user@domain.com' has registration token 'U2Ewt1UCn3ql'
```

You can access your `zrok` controller's registration UI by pointing a web browser at:

```
http://localhost:18080/register/U2Ewt1UCn3ql
```

The UI will ask you to set a password for your new account. Go ahead and do that.

After doing that, I see the following output in my controller console:

```
[ 516.778]    INFO zrok/controller.(*registerHandler).Handle: created account 'user@domain.com' with token 'SuGzRPjVDIcF'
```

Keep track of the token listed above (`SuGzRPjVDIcF`). We'll use this to enable our shell for this `zrok` deployment.

## Enable Your Shell

```bash
$ zrok enable SuGzRPjVDIcF
zrok environment '2AS1WZ3Sz' enabled for 'SuGzRPjVDIcF'
```

Congratulations. You have a working `zrok` environment!
