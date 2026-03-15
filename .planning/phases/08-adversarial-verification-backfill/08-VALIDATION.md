---
phase: 08
slug: adversarial-verification-backfill
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-14
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

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

- **After every task commit:** Run the smallest relevant adversarial backfill command plus any affected artifact check.
- **After every plan wave:** Run `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 08-01-01 | 01 | 1 | VER-05 | adversarial re-verification | `bash -lc 'GOCACHE="$PWD/.cache/go-build" go test ./... ./muid/... -run '\''TestRuntimeAdversarial|TestPublicHelperAdversarial|FuzzHSMEventScriptProperties|FuzzMUIDGeneratorProperties|Test.*RegressionAllocs'\'' -count=1 && GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh'` | ✅ | ⬜ pending |
| 08-02-01 | 02 | 2 | VER-05 | verification artifact | `bash -lc 'REPORT=.planning/phases/04-adversarial-regression-matrix/04-VERIFICATION.md && test -f "$REPORT" && rg -n "^# Phase 4: Adversarial Regression Matrix Verification Report$" "$REPORT" && rg -n "^\\*\\*Status:\\*\\* passed$" "$REPORT" && rg -n "^### Observable Truths$" "$REPORT" && rg -n "^### Required Artifacts$" "$REPORT" && rg -n "^### Key Link Verification$" "$REPORT" && rg -n "^## Requirements Coverage$" "$REPORT" && rg -n "VER-05" "$REPORT" && rg -n "hsm_runtime_adversarial_test\\.go|hsm_properties_fuzz_test\\.go|muid/muid_fuzz_test\\.go|hsm_regression_allocs_test\\.go|muid/muid_regression_allocs_test\\.go" "$REPORT" && rg -n "scripts/verify-workspace\\.sh|BenchmarkNestedStates_NoEntryExitActivity|BenchmarkMUIDGeneration" "$REPORT"'` | ❌ | ⬜ pending |
| 08-02-02 | 02 | 2 | VER-05 | validation reconciliation | `bash -lc 'VALIDATION=.planning/phases/04-adversarial-regression-matrix/04-VALIDATION.md && ! rg -n "^status: draft|⬜ pending|Approval: pending" "$VALIDATION" && rg -n "04-01-01|04-01-02|04-02-01|04-02-02|04-03-01|04-03-02" "$VALIDATION" && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" go test ./... -run '\''TestRuntimeStopTimeoutDispatchesErrorEventDeterministically\\|TestRuntimeRulesConformance'\''" "$VALIDATION" && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" go test ./... -run '\''TestRuntimeAdversarial\\|TestPublicHelperAdversarial'\'' -count=1" "$VALIDATION" && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" go test ./... -run '\''FuzzHSMEventScriptProperties'\'' -count=1" "$VALIDATION" && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" go test ./muid/... -run '\''FuzzMUIDGeneratorProperties'\'' -count=1" "$VALIDATION" && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" bash scripts/verify-workspace\\.sh" "$VALIDATION"'` | ✅ | ⬜ pending |
| 08-03-01 | 03 | 3 | VER-05 | traceability closure | `bash -lc 'AUDIT=.planning/v1.0-MILESTONE-AUDIT.md && rg -n "^- \\[x\\] \\*\\*VER-05\\*\\*" .planning/REQUIREMENTS.md && rg -n "^\\| VER-05 \\| Phase 8 \\| Complete \\|" .planning/REQUIREMENTS.md && rg -n "04-VERIFICATION\\.md" "$AUDIT" && ! rg -n "VER-05.*orphaned|04-VERIFICATION\\.md does not exist" "$AUDIT"'` | ✅ | ⬜ pending |

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
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
