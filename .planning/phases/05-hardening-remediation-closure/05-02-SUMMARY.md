---
phase: 05-hardening-remediation-closure
plan: "02"
subsystem: docs
tags: [go, docs, contract, rules, synchronization]
requires:
  - phase: 05-hardening-remediation-closure
    provides: nil-stop helper consistency and the hardened public helper baseline
provides:
  - Tracked `rules.md` contract wording for HSM53 and HSM54
  - Exported helper comments aligned to completion-channel production synchronization
  - README guidance that matches the hardened waiter-helper contract
affects: [phase-05-plan-03, release-gate, public-api-docs]
tech-stack:
  added: []
  patterns: [completion-channel synchronization contract, waiter helpers as deterministic observation tools]
key-files:
  created: [rules.md]
  modified: [hsm.go, README.md]
key-decisions:
  - "Completion channels returned by Dispatch, Set, Restart, Stop, DispatchAll, and DispatchTo remain the only supported production synchronization path."
  - "AfterProcess, AfterDispatch, AfterEntry, AfterExit, and AfterExecuted are documented as test and deterministic observation helpers, not general synchronization primitives."
patterns-established:
  - "Public contract updates must land in rules, exported comments, and README in the same remediation change."
  - "Waiter helpers are documented as observation aids while runtime completion channels define production sequencing."
requirements-completed: [VER-06]
duration: 2 min
completed: 2026-03-14
---

# Phase 5 Plan 2: Contract Alignment Summary

**Tracked rule wording, exported helper comments, and README guidance now all direct production callers to completion channels while limiting waiter helpers to deterministic observation**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-14T03:17:09Z
- **Completed:** 2026-03-14T03:19:13Z
- **Tasks:** 1
- **Files modified:** 3

## Accomplishments

- Aligned `hsm.go` exported comments for `Dispatch`, `Set`, `DispatchAll`, `DispatchTo`, `Restart`, and `Stop` around the hardened completion-channel contract.
- Reframed `AfterProcess`, `AfterDispatch`, `AfterEntry`, `AfterExit`, and `AfterExecuted` in both exported docs and README as tests and deterministic observation only.
- Brought `rules.md` under version control in the same task commit so HSM53 and HSM54 are tracked as the authoritative public contract artifact.

## Task Commits

Each task was committed atomically:

1. **Task 1: Reconcile rule text, exported comments, and README helper guidance** - `e650ea3` (`fix`)

## Files Created/Modified

- `rules.md` - Tracks the authoritative HSM rule contract, including HSM53 and HSM54, in git.
- `hsm.go` - Aligns exported helper comments with the hardened production synchronization contract.
- `README.md` - Mirrors the public helper contract for users reading package-level documentation.

## Decisions Made

- Production synchronization guidance now points only to completion channels returned by the public helper operations that settle runtime work.
- Waiter helpers remain documented and supported for deterministic observation, but explicitly not as general production synchronization primitives.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The plan's verification shell snippet needed safe quoting for literal backticks, but the underlying rule and test checks passed once run without command-substitution expansion.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The hardened synchronization contract now reads consistently across rules, exported docs, and README.
- Phase 05-03 can proceed from a tracked public contract baseline without carrying doc drift forward.

## Self-Check: PASSED

- Found `.planning/phases/05-hardening-remediation-closure/05-02-SUMMARY.md`
- Found commit `e650ea3`

---
*Phase: 05-hardening-remediation-closure*
*Completed: 2026-03-14*
