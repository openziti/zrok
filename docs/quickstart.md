# zrok quickstart

## ziti quickstart

```
$ source /dev/stdin <<< "$(wget -qO- https://raw.githubusercontent.com/openziti/ziti/release-next/quickstart/docker/image/ziti-cli-functions.sh)"; expressInstall
```

## configure frontend identity

```
$ ziti edge create identity device -o ~/.zrok/frontend.jwt frontend
New identity proxy created with id: -zbBF8eVb-
Enrollment expires at 2022-08-10T18:46:16.641Z
```

```
$ ziti edge enroll -j ~/.zrok/frontend.jwt -o ~/.zrok/identities/frontend.json
INFO    generating 4096 bit RSA key                  
INFO    enrolled successfully. identity file written to: proxy.json
```

```
$ ziti edge create erp frontend --edge-router-roles "#all" --identity-roles @frontend
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