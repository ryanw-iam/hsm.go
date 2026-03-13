---
phase: 02-deterministic-runtime-semantics
plan: "03"
subsystem: testing
tags: [go, hsm, lifecycle, determinism, race]
requires:
  - phase: 02-deterministic-runtime-semantics
    provides: deterministic timer and concurrency harnesses from plans 01 and 02
provides:
  - Focused deterministic coverage for stop, restart, activity cancellation, timeout dispatch, and lifecycle waiter behavior
  - Narrow runtime fixes for stop-driven exit observation and stop-queued lifecycle event processing
affects: [02-deterministic-runtime-semantics, runtime-verification, release-gate]
tech-stack:
  added: []
  patterns: [focused lifecycle suites, waiter-based lifecycle assertions, manual clock-triggered timeout verification]
key-files:
  created: [hsm_runtime_lifecycle_test.go]
  modified: [hsm.go]
key-decisions:
  - "Lifecycle waiters should observe public state paths during stop and restart, not internal concurrent-behavior names."
  - "Stop must hand off any lifecycle events queued while it holds the processing lock so timeout-driven errors remain observable under race."
patterns-established:
  - "Lifecycle-sensitive runtime behavior belongs in a dedicated root suite that uses explicit waiters and clock injection instead of incidental sleeps."
  - "Race-gate fixes in Phase 2 stay confined to the lifecycle synchronization path surfaced by deterministic tests."
requirements-completed: [VER-03]
duration: 7min
completed: 2026-03-13
---

# Phase 2 Plan 3: Deterministic Runtime Semantics Summary

**Deterministic lifecycle coverage for stop, restart, cancellation, timeout-driven errors, and race-clean stop synchronization**

## Performance

- **Duration:** 7 min
- **Started:** 2026-03-13T03:50:00Z
- **Completed:** 2026-03-13T03:57:01Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Added a focused lifecycle runtime suite that covers stop cancellation, restart re-entry with initial data, termination timeout dispatch, and post-stop waiter boundaries without wall-clock sleeps.
- Fixed `AfterExecuted` and `AfterExit` lifecycle observation gaps so stop-driven activity completion and exits are visible through the public waiter APIs.
- Fixed a stop-path synchronization bug that could strand timeout-driven lifecycle events under the canonical `-race -short` gate.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add deterministic lifecycle and termination-timeout coverage** - `d924379` (fix)
2. **Task 2: Prove the deterministic runtime suite under the canonical race-enabled gate** - `cd16f63` (fix)

## Files Created/Modified
- `hsm_runtime_lifecycle_test.go` - Focused deterministic coverage for stop, restart, timeout, and lifecycle waiter semantics.
- `hsm.go` - Narrow lifecycle runtime fixes for stop-driven exit/executed waiters and stop-queued event handoff.

## Decisions Made
- Kept lifecycle verification in a dedicated root suite so stop and restart behavior can be asserted with explicit waiter channels and injected clock triggers.
- Limited runtime changes to lifecycle observation and stop synchronization paths surfaced by the new deterministic tests rather than broad queue or helper refactors.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Closed public lifecycle waiters on stop-driven activity completion and exit**
- **Found during:** Task 1 (Add deterministic lifecycle and termination-timeout coverage)
- **Issue:** `AfterExecuted` waiters on state paths never closed for direct state activities, and `AfterExit` waiters were not closed when `Stop` exited states directly.
- **Fix:** Closed executed waiters for direct state-owner activities and closed exit waiters from the stop loop.
- **Files modified:** hsm.go
- **Verification:** `GOCACHE="$PWD/.cache/go-build" go test ./...`
- **Committed in:** d924379

**2. [Rule 1 - Bug] Resumed processing for lifecycle events queued while Stop held the processing lock**
- **Found during:** Task 2 (Prove the deterministic runtime suite under the canonical race-enabled gate)
- **Issue:** A termination-timeout `ErrorEvent` dispatched during `Stop` could be enqueued before the lock released and then remain unprocessed under `-race`.
- **Fix:** After stop releases the processing lock, it now kicks off processing if lifecycle events were queued during shutdown.
- **Files modified:** hsm.go
- **Verification:** `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Committed in:** cd16f63

**3. [Rule 3 - Blocking] Redirected Go build cache for local verification**
- **Found during:** Task 1 and Task 2 verification
- **Issue:** The default Go build cache path was unavailable in the execution environment.
- **Fix:** Ran verification with `GOCACHE` pointed at `.cache/go-build` in the repository.
- **Files modified:** None
- **Verification:** `GOCACHE="$PWD/.cache/go-build" go test ./...` and `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Committed in:** None

---

**Total deviations:** 3 auto-fixed (2 bug, 1 blocking)
**Impact on plan:** All deviations were required to make the lifecycle contract deterministic and race-clean. No scope creep beyond Phase 2 runtime semantics.

## Issues Encountered
The race gate exposed a real stop-path synchronization bug rather than a flaky test, which required a narrow runtime fix before the canonical verification entrypoint would pass consistently.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
Phase 2 now has deterministic runtime coverage for timer, concurrency, multi-instance, and lifecycle-sensitive semantics, all validated by the canonical workspace gate including `-race -short`.
The next phase can build on these focused root suites without revisiting sleep-based lifecycle assertions.

## Self-Check: PASSED
- FOUND: `.planning/phases/02-deterministic-runtime-semantics/02-03-SUMMARY.md`
- FOUND: `d924379`
- FOUND: `cd16f63`
