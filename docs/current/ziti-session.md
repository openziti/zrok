# Controller OpenZiti management session

The controller shares one OpenZiti management client per API endpoint and username for the life of the process, including bandwidth-limit enforcement and relaxation. It fetches the controller CA bundle and authenticates when that client is first created. A plain `http://` endpoint has no TLS, so the CA fetch is skipped for it; this is what lets controller tests reach a fake Ziti through the production client. A failed initial build is not cached; the next call tries again.

When a management request receives HTTP 401, the transport refreshes the session and replays the request once, including its body. Concurrent requests share one in-flight refresh; no lock is held across the authentication call. The CA bundle and management transport remain in use across refreshes. Failed re-authentication or a second 401 returns the operation's typed unauthorized error to its caller; there is no retry loop.

A failed refresh is remembered for five seconds: requests still holding the failed session during that window get the unauthorized error immediately, and the first request after it makes one fresh attempt. A persistent authentication failure therefore costs one login per window and fails operations quickly instead of serializing them. The first build of the client for an endpoint and username happens outside the cache lock; concurrent callers for the same key wait for that build and share its result, and callers for other keys are not blocked.

The process-wide `zrok.ziti.authentications` expvar counter counts successful authentications. Read it at `/debug/vars` on the configured `admin.profile_endpoint` listener. Each successful authentication also writes an INFO log with the running count and reason (`initial` or `expired`). After startup the count should remain stable across ordinary share creates and deletes, increasing only when an expired session needs replacement.

Credentials are retained from the first successful build for an endpoint and username. Restart the controller after changing credentials. Separate admin command processes have their own session.
