---
phase: 5
slug: hardening-remediation-closure
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-13
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` |
| **Config file** | `go.mod` |
| **Quick run command** | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|Test.*RegressionAllocs'` |
| **Full suite command** | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|Test.*RegressionAllocs'`
- **After every plan wave:** Run `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | VER-06 | unit | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial'` | ✅ | ⬜ pending |
| 05-02-01 | 02 | 2 | VER-06 | unit+docs | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeRulesConformance|TestRuntimeAdversarial'` | ✅ | ⬜ pending |
| 05-03-01 | 03 | 3 | VER-06 | smoke+hygiene | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Release-clean tracked contract artifacts and stale CI cleanup | VER-06 | Git tracking and file-presence hygiene are repository-state checks, not runtime behavior | Run `git ls-files --error-unmatch rules.md`, confirm phase-neutral wording in `scripts/verify-workspace.sh`, and confirm `.travis.yml` is removed or intentionally archived. |

---

## Validation Sign-Off

- [x] All tasks have automated verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all missing references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
