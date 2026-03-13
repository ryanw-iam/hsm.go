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
| 04-01-01 | 01 | 1 | VER-05 | unit | `GOCACHE=$(pwd)/.cache/go-build go test ./... -run 'Test.*Adversarial'` | ❌ W0 | ⬜ pending |
| 04-02-01 | 02 | 2 | VER-05 | fuzz-seed | `GOCACHE=$(pwd)/.cache/go-build go test ./... ./muid/... -run 'Fuzz'` | ❌ W0 | ⬜ pending |
| 04-03-01 | 03 | 3 | VER-05 | allocs+bench-smoke | `GOCACHE=$(pwd)/.cache/go-build go test -run '^$' -bench '^(BenchmarkNestedStates_NoEntryExitActivity|BenchmarkMUIDGeneration)$' -benchtime=20x -benchmem ./ ./muid` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Existing infrastructure covers all phase requirements.

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
