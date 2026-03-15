---
phase: 4
slug: adversarial-regression-matrix
status: passed
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-13
---

# Phase 4 — Validation Record

> Executed validation record for the shipped Phase 4 adversarial-regression scope, reconciled during the Phase 8 backfill.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `GOCACHE="$PWD/.cache/go-build" go test ./... ./muid/... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|FuzzHSMEventScriptProperties|FuzzMUIDGeneratorProperties|Test.*RegressionAllocs' -count=1` |
| **Full suite command** | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run the smallest relevant adversarial or regression command with `GOCACHE="$PWD/.cache/go-build"`.
- **After every plan wave:** Run `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Validation Architecture

- Helper/runtime guard coverage stays anchored to `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeStopTimeoutDispatchesErrorEventDeterministically|TestRuntimeRulesConformance' -count=1`.
- Runtime adversarial coverage is isolated behind `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial' -count=1`.
- Root and `muid` property evidence replay through ordinary `go test` with checked-in corpora via `FuzzHSMEventScriptProperties` and `FuzzMUIDGeneratorProperties`.
- Allocation regression evidence stays explicit and bounded through package-scoped `Test.*RegressionAllocs` reruns.
- The canonical workspace gate remains `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`, so adversarial, fuzz-seed, allocation, benchmark-smoke, and race-short evidence all terminate in the same release entrypoint.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | VER-05 | unit-helper | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeStopTimeoutDispatchesErrorEventDeterministically|TestRuntimeRulesConformance' -count=1` | ✅ | ✅ green |
| 04-01-02 | 01 | 1 | VER-05 | unit-adversarial | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial' -count=1` | ✅ | ✅ green |
| 04-02-01 | 02 | 1 | VER-05 | fuzz-seed-root | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'FuzzHSMEventScriptProperties' -count=1` | ✅ | ✅ green |
| 04-02-02 | 02 | 1 | VER-05 | fuzz-seed-muid | `GOCACHE="$PWD/.cache/go-build" go test ./muid/... -run 'FuzzMUIDGeneratorProperties' -count=1` | ✅ | ✅ green |
| 04-03-01 | 03 | 2 | VER-05 | alloc-regression | `bash -lc 'GOCACHE="$PWD/.cache/go-build" go test ./... -run '\''Test.*RegressionAllocs'\'' -count=1 && GOCACHE="$PWD/.cache/go-build" go test ./muid/... -run '\''Test.*RegressionAllocs'\'' -count=1'` | ✅ | ✅ green |
| 04-03-02 | 03 | 2 | VER-05 | canonical-gate | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` | ✅ | ✅ green |

*Status: ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `.planning/phases/04-adversarial-regression-matrix/04-01-PLAN.md` — adversarial runtime and helper-contract plan
- [ ] `.planning/phases/04-adversarial-regression-matrix/04-02-PLAN.md` — bounded property and seed-corpus plan
- [ ] `.planning/phases/04-adversarial-regression-matrix/04-03-PLAN.md` — allocation regression and canonical verifier plan

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have automated verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all missing references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** complete
