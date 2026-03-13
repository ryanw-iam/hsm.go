---
phase: 4
slug: adversarial-regression-matrix
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-13
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` |
| **Config file** | `go.mod` |
| **Quick run command** | `GOCACHE=$(pwd)/.cache/go-build go test ./... ./muid/... -run 'Test.*Adversarial|Test.*Regression|Fuzz'` |
| **Full suite command** | `GOCACHE=$(pwd)/.cache/go-build bash scripts/verify-workspace.sh` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `GOCACHE=$(pwd)/.cache/go-build go test ./... ./muid/... -run 'Test.*Adversarial|Test.*Regression|Fuzz'`
- **After every plan wave:** Run `GOCACHE=$(pwd)/.cache/go-build bash scripts/verify-workspace.sh`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | VER-05 | unit-helper | `GOCACHE=$(pwd)/.cache/go-build go test ./... -run 'TestRuntimeStopTimeoutDispatchesErrorEventDeterministically|TestRuntimeRulesConformance'` | ✅ existing | ⬜ pending |
| 04-01-02 | 01 | 1 | VER-05 | unit-adversarial | `GOCACHE=$(pwd)/.cache/go-build go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial'` | ❌ planned | ⬜ pending |
| 04-02-01 | 02 | 1 | VER-05 | fuzz-seed-root | `GOCACHE=$(pwd)/.cache/go-build go test ./... -run 'FuzzHSMEventScriptProperties'` | ❌ planned | ⬜ pending |
| 04-02-02 | 02 | 1 | VER-05 | fuzz-seed-muid | `GOCACHE=$(pwd)/.cache/go-build go test ./muid/... -run 'FuzzMUIDGeneratorProperties'` | ❌ planned | ⬜ pending |
| 04-03-01 | 03 | 2 | VER-05 | alloc-regression | `GOCACHE=$(pwd)/.cache/go-build go test ./... ./muid/... -run 'Test.*RegressionAllocs'` | ❌ planned | ⬜ pending |
| 04-03-02 | 03 | 2 | VER-05 | canonical-gate | `GOCACHE=$(pwd)/.cache/go-build bash scripts/verify-workspace.sh` | ✅ existing | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave Structure

- **Wave 1:** 04-01 and 04-02 run in parallel. 04-01 owns helper/runtime fault-injection work. 04-02 stays limited to stable event-script and `muid` property coverage so it does not depend on 04-01 remediation.
- **Wave 2:** 04-03 depends on both Wave 1 plans and wires the canonical gate to execute allocation regression checks explicitly before benchmark smoke.

## Wave 0 Requirements

- [x] Existing deterministic harness, waiters, clock injection, and canonical verifier already exist from Phases 1-3.
- [x] No extra Wave 0 scaffolding is required before Phase 4 execution.

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

**Approval:** pending
