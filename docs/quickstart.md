# zrok quickstart

## ziti quickstart

```
$ source /dev/stdin <<< "$(wget -qO- https://raw.githubusercontent.com/openziti/ziti/release-next/quickstart/docker/image/ziti-cli-functions.sh)"; expressInstall
```

## configure frontend identity

```
$ ziti edge create identity service frontend -o ~/.zrok/identities/frontend.jwt
New identity proxy created with id: -zbBF8eVb-
Enrollment expires at 2022-08-10T18:46:16.641Z
```

```
$ ziti edge enroll -j ~/.zrok/identities/frontend.jwt -o ~/.zrok/identities/frontend.json
INFO    generating 4096 bit RSA key                  
INFO    enrolled successfully. identity file written to: proxy.json
```

```
$ ziti edge create erp frontend --edge-router-roles "#all" --identity-roles "@frontend"
New edge router policy frontend created with id: aOxvtWIanFIAwuU51lF9SU
```

## configure controller identity
```
$ ziti edge create identity service ctrl -o ~/.zrok/identities/ctrl.jwt 
New identity ctrl created with id: e8c3tQo3SR
Enrollment expires at 2022-10-14T19:59:01.908Z
```

```
$ ziti edge enroll -j ~/.zrok/identities/ctrl.jwt -o ~/.zrok/identities/ctrl.json
INFO    generating 4096 bit RSA key                  
INFO    enrolled successfully. identity file written to: /home/michael/.zrok/identities/ctrl.json 
```

```
$ ziti edge create erp ctrl --edge-router-roles "#all" --identity-roles "@ctrl"
New edge router policy ctrl created with id: 7OxvtWIanFIAwuU51lF9SU
```

## create metrics service
```
$ ziti edge create service metrics
New service metrics created with id: 56y5AFsKmSsIWLsmFNBeJz
```

### create service edge router policy for metrics service
```
$ ziti edge create serp ctrl-public --service-roles "@metrics" --edge-router-roles "#all"
```

### allow the controller to bind the metrics service
```
$ ziti edge create sp ctrl-bind Bind --identity-roles "@ctrl" --service-roles "@metrics"
New service policy ctrl-bind created with id: 3SXgFftSgBnenjgdBENOGR
```

### allow frontends to dial the metrics service
```
$ ziti edge create sp ctrl-dial Dial --identity-roles "@frontend" --service-roles "@metrics"
New service policy ctrl-dial created with id: 6pCe9uGj8oB2JXlWb44x2u
```

## start zrok resources

adjust `ctrl.yml` (or create a copy for your environment).

```
$ zrok ctrl etc/ctrl.yml
```

```
$ zrok proxy ~/.zrok/proxy.json
```

## create zrok account

```
$ zrok create account 
New Username: michael@quigley.com
New Password: 
Confirm Password: 
[   3.122]    INFO main.glob..func1: api token: 9ae56d39a6e96d65a45518b5ea1637a0677581a33ba44bbc3c103f6351ec478fb8185e97a993382ed2daa26720d40b052824dbce5ef38874c82893f33e445b06
```

## enable zrok for your shell

```
$ zrok enable 9ae56d39a6e96d65a45518b5ea1637a0677581a33ba44bbc3c103f6351ec478fb8185e97a993382ed2daa26720d40b052824dbce5ef38874c82893f33e445b06
[   0.691]    INFO main.enable: enabled, identity = 'ARjEc8eVA-'
```

## tunnel

```
$ zrok http <endpoint>
```