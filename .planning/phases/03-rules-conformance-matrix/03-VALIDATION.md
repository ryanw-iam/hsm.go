---
phase: 03
slug: rules-conformance-matrix
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-13
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `bash scripts/verify-workspace.sh` |
| **Estimated runtime** | ~20 seconds |

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
| 03-01-01 | 01 | 1 | VER-04 | matrix/classification | `go test ./...` | ❌ W0 | ⬜ pending |
| 03-02-01 | 02 | 1 | VER-04 | define-panic conformance | `go test ./...` | ❌ W0 | ⬜ pending |
| 03-03-01 | 03 | 2 | VER-04 | runtime/exemplar conformance | `go test ./...` | ❌ W0 | ⬜ pending |
| 03-03-02 | 03 | 2 | VER-04 | release-gated rules closeout | `bash scripts/verify-workspace.sh` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `.planning/phases/03-rules-conformance-matrix/03-01-PLAN.md` — rule matrix and classification inventory plan
- [ ] `.planning/phases/03-rules-conformance-matrix/03-02-PLAN.md` — define/build-time conformance suite plan
- [ ] `.planning/phases/03-rules-conformance-matrix/03-03-PLAN.md` — runtime/exemplar conformance and full-gate plan

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
