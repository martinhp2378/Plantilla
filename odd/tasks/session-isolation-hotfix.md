# Session isolation hotfix

## Goal
Prevent cross-machine session identity takeover before starting the broader authorization decoupling.

## Scope
- Patch the shared `sesiones` library so requests without a valid UUID cookie never recover another anonymous session by browser fingerprint.
- Preserve valid-cookie session recovery, rotation, expiry, Redis/PostgreSQL persistence, flash messages, and logout behavior.
- Release a new compatible library version and update PlantillaInstitucional.
- Invalidate pre-fix active sessions as part of deployment guidance.

## Tasks
- [x] Reproduce the two-client identity takeover in the shared library.
- [x] Remove unsafe fingerprint fallback and add compatibility regressions.
- [x] Verify and release the patched library version.
- [x] Update PlantillaInstitucional and verify login/session compatibility.
- [x] Document mandatory existing-session invalidation and deployment order.

## Evidence
- Root cause: cookie-less requests reuse an anonymous session through a weak fingerprint composed of User-Agent, Accept-Language, and Accept-Encoding.
- `HabilitarFingerprinting` was ignored by the vulnerable middleware and was not an effective mitigation.
- Shared library release: `v0.4.5` at `c1a4df610907b7ada523549b9116ae188910086b`, based directly on `v0.4.3` without dependency upgrades.
- PlantillaInstitucional integration: focused authentication tests and full Go suite pass.
- Rollout contract: rotate `COOKIE_NOMBRE`, invalidate all pre-fix server-side sessions, then require every user to authenticate again.
