---
phase: 01
slug: coverage-baseline-and-release-gates
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-13
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./... ./kind/... ./muid/...` |
| **Full suite command** | `go test -race -short ./... ./kind/... ./muid/...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... ./kind/... ./muid/...`
- **After every plan wave:** Run `go test -race -short ./... ./kind/... ./muid/...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 20 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | VER-01 | unit/integration | `go test ./... ./kind/... ./muid/...` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 1 | VER-01 | unit/edge | `go test ./... ./kind/... ./muid/...` | ❌ W0 | ⬜ pending |
| 01-03-01 | 03 | 2 | VER-02 | integration/race | `go test -race -short ./... ./kind/... ./muid/...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `.planning/phases/01-coverage-baseline-and-release-gates/01-01-PLAN.md` — root runtime coverage inventory and restructuring plan
- [ ] `.planning/phases/01-coverage-baseline-and-release-gates/01-02-PLAN.md` — helper-module baseline closure plan
- [ ] `.planning/phases/01-coverage-baseline-and-release-gates/01-03-PLAN.md` — release gate and evidence automation plan

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
