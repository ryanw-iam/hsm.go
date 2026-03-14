---
phase: 06
slug: deterministic-runtime-verification-backfill
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-14
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(AfterUsesDeterministicTimer|AfterNegativeDurationDoesNotSchedule|AtUsesDeterministicTimer|EveryUsesManualTriggers|WhenWaitsForExplicitSignal|ConfigClockOverridesDefaultClock|DefaultClockFallbackUsesDeterministicTimer|ConcurrentDispatchSerializesProcessing|DeferredEventsReplayAfterTransition|CompletionEventsPreemptQueuedEvents|DispatchAllReachesEveryStartedInstance|DispatchToTargetsMatchingInstancesOnly|StopCancelsActivityAndClosesAfterExecuted|RestartReentersInitialStateWithData|StopTimeoutDispatchesErrorEventDeterministically)$' -count=1` |
| **Full suite command** | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task commit:** Run the smallest relevant targeted deterministic-runtime `go test` command plus any affected artifact check.
- **After every plan wave:** Run `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 20 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | VER-03 | evidence inventory | `bash -lc "test -f .planning/phases/02-deterministic-runtime-semantics/02-01-SUMMARY.md && test -f .planning/phases/02-deterministic-runtime-semantics/02-02-SUMMARY.md && test -f .planning/phases/02-deterministic-runtime-semantics/02-03-SUMMARY.md && ! test -f .planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md"` | ✅ | ⬜ pending |
| 06-01-02 | 01 | 1 | VER-03 | deterministic runtime suites | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(AfterUsesDeterministicTimer|AfterNegativeDurationDoesNotSchedule|AtUsesDeterministicTimer|EveryUsesManualTriggers|WhenWaitsForExplicitSignal|ConfigClockOverridesDefaultClock|DefaultClockFallbackUsesDeterministicTimer|ConcurrentDispatchSerializesProcessing|DeferredEventsReplayAfterTransition|CompletionEventsPreemptQueuedEvents|DispatchAllReachesEveryStartedInstance|DispatchToTargetsMatchingInstancesOnly|StopCancelsActivityAndClosesAfterExecuted|RestartReentersInitialStateWithData|StopTimeoutDispatchesErrorEventDeterministically)$' -count=1` | ✅ | ⬜ pending |
| 06-02-01 | 02 | 2 | VER-03 | verification artifact | `bash -lc "test -f .planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md && rg -n 'VER-03|Status: passed|Observable Truths|Requirements Coverage' .planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md"` | ❌ W0 | ⬜ pending |
| 06-02-02 | 02 | 2 | VER-03 | validation reconciliation | `bash -lc \"! rg -n '^status: draft|⬜ pending|Approval: pending' .planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md\"` | ✅ | ⬜ pending |
| 06-03-01 | 03 | 3 | VER-03 | traceability closure | `bash -lc \"rg -n 'VER-03' .planning/REQUIREMENTS.md .planning/ROADMAP.md .planning/v1.0-MILESTONE-AUDIT.md .planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md\"` | ✅ | ⬜ pending |
| 06-03-02 | 03 | 3 | VER-03 | canonical gate | `GOCACHE=\"$PWD/.cache/go-build\" bash scripts/verify-workspace.sh` | ✅ | ⬜ pending |

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
