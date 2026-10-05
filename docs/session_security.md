# Deploy the session-isolation hotfix safely

Version `v0.4.5` of `github.com/issemym-repos/sesiones` prevents clients without a valid session cookie from recovering an anonymous session through a shared browser fingerprint. Deployment must also invalidate sessions created before the fix.

## Required rollout

1. Back up the configured session store and enter an authorized maintenance window.
2. Deploy the application using `sesiones v0.4.5` while traffic remains closed.
3. Rotate `COOKIE_NOMBRE` to a namespace never used by the affected deployment, such as `plantilla_v2`.
4. Invalidate all existing server-side sessions in the configured PostgreSQL or Redis store using the approved operator procedure for that environment.
5. Reopen traffic and confirm that every user must authenticate again.
6. Verify with two separate clients that signing in on one client never changes the identity observed by the other.

## Why both actions are mandatory

The library fix prevents creation of new shared anonymous sessions. It cannot revoke a UUID cookie that two clients already received before deployment. Rotating the cookie name makes old browser cookies unreachable; clearing the server-side store revokes the corresponding session records.

| Store | Required outcome |
|---|---|
| PostgreSQL | Every pre-fix session row is invalidated using the approved database maintenance procedure. |
| Redis | Every pre-fix session key and index is invalidated using the approved cache maintenance procedure. |

Do not improvise destructive production commands. Session invalidation requires the environment owner, a verified backup when applicable, and evidence that only session data is affected.

## Rollback rule

A rollback must not restore the vulnerable cookie namespace or a pre-fix session store snapshot. If application rollback is unavoidable, keep the rotated cookie name, invalidate sessions again, and block traffic until an approved safe version is running.

## Verification checklist

- [ ] The deployed module resolves to `github.com/issemym-repos/sesiones v0.4.5` or newer.
- [ ] `COOKIE_NOMBRE` differs from the pre-fix deployment value.
- [ ] Pre-fix PostgreSQL or Redis sessions are no longer accepted.
- [ ] Two cookie-less clients with identical browser headers receive different session UUIDs.
- [ ] A valid cookie still restores only its owning client's session.
- [ ] All users were required to authenticate again after deployment.
