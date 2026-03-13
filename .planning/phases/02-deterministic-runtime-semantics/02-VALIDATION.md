---
phase: 02
slug: deterministic-runtime-semantics
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-13
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `bash scripts/verify-workspace.sh` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `bash scripts/verify-workspace.sh`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 20 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | VER-03 | harness/timers | `go test ./...` | ✅ | ⬜ pending |
| 02-01-02 | 01 | 1 | VER-03 | timers/waiters | `go test ./...` | ✅ | ⬜ pending |
| 02-02-01 | 02 | 2 | VER-03 | concurrency/order | `go test ./...` | ✅ | ⬜ pending |
| 02-02-02 | 02 | 2 | VER-03 | multi-instance targeting | `go test ./...` | ✅ | ⬜ pending |
| 02-03-01 | 03 | 3 | VER-03 | lifecycle/timeout | `go test ./...` | ✅ | ⬜ pending |
| 02-03-02 | 03 | 3 | VER-03 | race/full-gate | `bash scripts/verify-workspace.sh` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `.planning/phases/02-deterministic-runtime-semantics/02-01-PLAN.md` — deterministic timer and waiter harness plan
- [x] `.planning/phases/02-deterministic-runtime-semantics/02-02-PLAN.md` — dispatch ordering and multi-instance semantics plan
- [x] `.planning/phases/02-deterministic-runtime-semantics/02-03-PLAN.md` — lifecycle timeout and race-hardening plan

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 20s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
