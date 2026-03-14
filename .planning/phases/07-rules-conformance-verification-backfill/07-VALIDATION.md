---
phase: 07
slug: rules-conformance-verification-backfill
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-14
---

# Phase 07 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance' -count=1` |
| **Full suite command** | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task commit:** Run the smallest relevant rules-conformance command plus any affected artifact check.
- **After every plan wave:** Run `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 20 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | VER-04 | rules conformance re-verification | `bash -lc 'GOCACHE="$PWD/.cache/go-build" go test ./... -run '\''TestModelRulesConformance|TestRuntimeRulesConformance'\'' -count=1 && GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh'` | ✅ | ⬜ pending |
| 07-02-01 | 02 | 2 | VER-04 | verification artifact | `bash -lc 'REPORT=.planning/phases/03-rules-conformance-matrix/03-VERIFICATION.md && test -f "$REPORT" && rg -n "^# Phase 3: Rules Conformance Matrix Verification Report$" "$REPORT" && rg -n "^\\*\\*Status:\\*\\* passed$" "$REPORT" && rg -n "^### Observable Truths$" "$REPORT" && rg -n "^### Required Artifacts$" "$REPORT" && rg -n "^### Key Link Verification$" "$REPORT" && rg -n "^### Requirements Coverage$" "$REPORT" && rg -n "VER-04" "$REPORT" && rg -n "03-RULES-MATRIX\\.md|TestModelRulesConformance|TestRuntimeRulesConformance" "$REPORT" && rg -n "scripts/verify-workspace\\.sh|canonical workspace gate" "$REPORT"'` | ❌ | ⬜ pending |
| 07-02-02 | 02 | 2 | VER-04 | validation reconciliation | `bash -lc 'VALIDATION=.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md && ! rg -n "^status: draft|⬜ pending|Approval: pending" "$VALIDATION" && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" go test ./... -run" "$VALIDATION" && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" bash scripts/verify-workspace\\.sh" "$VALIDATION"'` | ✅ | ⬜ pending |
| 07-03-01 | 03 | 3 | VER-04 | traceability closure | `bash -lc 'AUDIT=.planning/v1.0-MILESTONE-AUDIT.md && rg -n "^- \\[x\\] \\*\\*VER-04\\*\\*" .planning/REQUIREMENTS.md && rg -n "^\\| VER-04 \\| Phase 7 \\| Complete \\|" .planning/REQUIREMENTS.md && rg -n "03-VERIFICATION\\.md" "$AUDIT" && ! rg -n "VER-04.*orphaned|03-VERIFICATION\\.md does not exist" "$AUDIT" && rg -n "VER-05" "$AUDIT"'` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

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
