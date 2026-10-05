# Durable report retention

## Goal
Complete durable 24-hour report retention with restart-safe, multi-instance cleanup.

## Constraints
- Authorized review-size exception: up to 650 authored changed lines.
- Preserve tenant-private report delivery from commit `abedb80`.
- Use paired migration `000006` and PostgreSQL `SKIP LOCKED` claims.
- Keep external storage deletion outside database transactions.
- No commit, push, or PR without explicit user authorization.

## Tasks
- [x] Inspect and reconcile the interrupted implementation against the lifecycle contract.
- [x] Correct implementation and tests within the authorized edit surfaces and size limit.
- [x] Run focused, PostgreSQL integration, full-suite, diff, and line-budget verification.

## Evidence
- Baseline: `abedb80` (`fix(reports): secure and bound generated files`).
- Focused race tests: PASS.
- Full Go suite: PASS.
- `git diff --check`: PASS.
- Candidate size: 639 diff lines excluding this ODD artifact; authorized maximum is 650.
- PostgreSQL 16 integration: PASS using an authorized disposable container; container, volume, and network were removed afterward.
- LSP: no compiler/type findings; three auxiliary pattern warnings are false positives against pgx loop/finalizer code.
- Commit evidence: this work unit is committed as `feat(reports): add durable report retention`.
