# Controller OpenZiti management session

The controller shares one OpenZiti management client per API endpoint and username for the life of the process, including bandwidth-limit enforcement and relaxation. It fetches the controller CA bundle and authenticates when that client is first created. A failed initial build is not cached; the next call tries again.

When a management request receives HTTP 401, the transport refreshes the session and replays the request once, including its body. Concurrent requests check the session generation under a lock so they can share a refresh. The CA bundle and management transport remain in use across refreshes. Failed re-authentication or a second 401 returns the operation's typed unauthorized error to its caller; there is no retry loop.

The process-wide `zrok.ziti.authentications` expvar counter counts successful authentications. Read it at `/debug/vars` on the configured `admin.profile_endpoint` listener. Each successful authentication also writes an INFO log with the running count and reason (`initial` or `expired`). After startup the count should remain stable across ordinary share creates and deletes, increasing only when an expired session needs replacement.

Credentials are retained from the first successful build for an endpoint and username. Restart the controller after changing credentials. Separate admin command processes have their own session.
