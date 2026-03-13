---
phase: 04-adversarial-regression-matrix
plan: "03"
subsystem: testing
tags: [go, testing, allocs, fuzz, benchmarks, verification]
requires:
  - phase: 04-01
    provides: deterministic adversarial runtime coverage and hostile helper-path assertions
  - phase: 04-02
    provides: checked-in fuzz targets and deterministic seed corpora for root and muid properties
provides:
  - Stable allocation regression gates for root dispatch and set flows
  - Stable allocation regression gates for default muid make and make-string paths
  - Canonical verifier coverage for adversarial suites, fuzz-seed replay, alloc regression, and benchmark smoke
affects: [release-gating, verification, 05-remediation-and-hardening]
tech-stack:
  added: []
  patterns:
    - testing.AllocsPerRun ceilings for release-blocking performance regressions
    - explicit package-scoped verifier steps for deterministic adversarial evidence
key-files:
  created:
    - hsm_regression_allocs_test.go
    - muid/muid_regression_allocs_test.go
  modified:
    - scripts/verify-workspace.sh
key-decisions:
  - Keep root hot-path regression contracts at measured allocation ceilings of 5 allocs/run for Dispatch and 9 allocs/run for Set rather than introducing wall-clock thresholds.
  - Keep benchmark smoke deterministic with fixed 20-iteration benchmem runs on the existing root and muid benchmark names.
patterns-established:
  - Allocation regressions should be enforced with narrow public-path contracts instead of broad microbenchmark baselines.
  - The canonical verifier should call adversarial, fuzz-seed, alloc, and benchmark evidence explicitly rather than relying on incidental package-wide test discovery.
requirements-completed: [VER-05]
duration: 3 min
completed: 2026-03-13
---

# Phase 4 Plan 03: Canonical Adversarial Gate Summary

**Stable allocation ceilings for root dispatch/set and default muid generation wired into the canonical verifier with explicit adversarial, fuzz-seed, and benchmark-smoke steps**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-13T21:12:30Z
- **Completed:** 2026-03-13T21:15:49Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added CI-safe allocation regression tests for root `Dispatch` and `Set` helper paths using `testing.AllocsPerRun`.
- Added default `muid.Make()` and `muid.MakeString()` allocation regression coverage with stable zero and one-allocation ceilings.
- Extended `scripts/verify-workspace.sh` so the canonical release gate now runs adversarial suites, checked-in fuzz-seed replay, allocation regressions, fixed-iteration benchmark smoke, and the existing race-short workspace pass.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add stable allocation regression checks for hot public paths** - `a368620` (test), `7e81bc9` (feat)
2. **Task 2: Wire the canonical verifier to run the bounded adversarial matrix** - `832df2d` (feat)

## Files Created/Modified

- `hsm_regression_allocs_test.go` - release-blocking allocation ceilings for root dispatch and set helper flows
- `muid/muid_regression_allocs_test.go` - release-blocking allocation ceilings for default muid generation paths
- `scripts/verify-workspace.sh` - single canonical verifier with explicit adversarial, fuzz-seed, alloc, benchmark-smoke, and race-short steps

## Decisions Made

- Measured and locked the root helper hot paths at `Dispatch <= 5 allocs/run` and `Set <= 9 allocs/run` instead of guessing stricter contracts that would become flaky blockers.
- Kept the benchmark gate as advisory smoke with `-benchmem -benchtime=20x` on existing benchmark names rather than introducing time-based pass/fail thresholds.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Concurrent `git add` operations triggered a transient `.git/index.lock`; resolved by serializing Git staging for the rest of the plan.
- `scripts/verify-workspace.sh` required `git add -f` because the repository ignore rules reject the normal add path for `scripts/`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 4 now closes on a canonical verifier that surfaces deterministic hostile-path, fuzz-seed, allocation, and benchmark-smoke evidence in one place.
- Phase 5 can treat any future adversarial or regression failures as remediation work instead of more verification expansion.

## Self-Check: PASSED

- Found `.planning/phases/04-adversarial-regression-matrix/04-03-SUMMARY.md`.
- Verified task commits `a368620`, `7e81bc9`, and `832df2d` in `git log --oneline --all`.

---
*Phase: 04-adversarial-regression-matrix*
*Completed: 2026-03-13*
