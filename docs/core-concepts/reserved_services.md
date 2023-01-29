# Reserved Services

With v0.3, `zrok` introduced a concept of "reserving" services. The intention is that the `zrok` control plane will support limits on the number of reserved services (and eventually `frontend`instances) that an account is allowed to utilize. Service reservations could also be time-limited, or possibly even bandwidth-limited (the reservation expires after a bandwidth threshold is crossed).

## Reserved Services Example

With v0.3 `zrok` introduced the `zrok reserve` command:

```
$ zrok reserve private http://localhost:9090
[   0.047]    INFO main.(*reserveCommand).run: your reserved service token is 'x88xujrpk4k3'
[   0.048]    INFO main.(*reserveCommand).run: your reserved service frontend is 'http://x88xujrpk4k3.zrok.quigley.com:8080/'
```

The `reserve` command creates a service reservation that allows a service to become non-ephemeral. The service token `x88xujrpk4k3` is guaranteed to exist between `backend` executions.

Running a `backend` against a service reservation is done like this:

```
$ zrok share reserved x88xujrpk4k3
[   0.005]    INFO main.(*shareReservedCommand).run: sharing target endpoint: 'http://localhost:9090'
[   0.040]    INFO main.(*shareReservedCommand).run: use this command to access your zrok service: 'zrok access private x88xujrpk4k3'
^C
$ zrok share reserved x88xujrpk4k3
[   0.007]    INFO main.(*shareReservedCommand).run: sharing target endpoint: 'http://localhost:9090'
[   0.047]    INFO main.(*shareReservedCommand).run: use this command to access your zrok service: 'zrok access private x88xujrpk4k3'
```

The `share reserved` comand starts a backend process for the service. User-facing and public-facing `frontend` instances are allowed to come and go, just as if the service were ephemeral.

Releasing a reserved service is done with the `zrok release` command:

```
$ zrok release x88xujrpk4k3
[   0.056]    INFO main.(*releaseCommand).run: reserved service 'x88xujrpk4k3' released
```

