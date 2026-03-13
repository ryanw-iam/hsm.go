---
phase: 04-adversarial-regression-matrix
plan: "01"
subsystem: testing
tags: [go, testing, hsm, adversarial, runtime]
requires:
  - phase: 02-deterministic-runtime-semantics
    provides: waiter-based synchronization and injected clock harnesses for deterministic runtime faults
  - phase: 03-rules-conformance-matrix
    provides: focused root runtime suites that isolate behavioral contracts by failure family
provides:
  - Deterministic adversarial runtime coverage for concurrent panic recovery, process panic recovery, and terminate-timeout faults
  - Hostile public-helper regression coverage for nil-instance and typed-nil-context dispatch, set, call, and broadcast helpers
  - Narrow helper/runtime hardening that preserves missing-context contracts instead of panicking on typed-nil contexts
affects: [phase-04-plan-02, phase-04-plan-03, phase-05-hardening-remediation-closure, verification]
tech-stack:
  added: []
  patterns: [error-event recorder helper, typed-nil-context guardrails, deterministic adversarial waiter assertions]
key-files:
  created: [hsm_runtime_adversarial_test.go]
  modified: [hsm_test_helpers_test.go, hsm.go]
key-decisions:
  - "Typed-nil helper contexts should collapse to the existing missing-context contract instead of panicking in context lookups."
  - "Adversarial runtime evidence should reuse ErrorEvent transitions, waiter channels, and the deterministic clock harness rather than introducing a second fault runner."
patterns-established:
  - "Hostile helper tests assert immediate closed-waiter behavior for nil or missing context paths."
  - "Process-level panic recovery is exercised through guard panics, while concurrent panic recovery is exercised through state activity panics."
requirements-completed: [VER-05]
duration: 4 min
completed: 2026-03-13
---

# Phase 4 Plan 1: Adversarial Fault-Injection Summary

**Deterministic adversarial runtime coverage for panic recovery, terminate-timeout faults, and typed-nil-context helper handling in the root package**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-13T21:04:16Z
- **Completed:** 2026-03-13T21:08:19Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Extended the shared root test helper surface with reusable ErrorEvent capture and immediate waiter-closure assertions for hostile helper paths.
- Added a focused adversarial runtime suite that covers concurrent activity panics, process-loop guard panics, terminate-timeout injection, and hostile public helper contracts.
- Hardened the runtime’s context lookup seams so typed-nil helper contexts now return closed waiters or `ErrMissingHSM` instead of panicking.

## Task Commits

Each task was committed atomically:

1. **Task 1: Extend the existing deterministic helper surface for adversarial assertions** - `e1c6166` (feat)
2. **Task 2: Add the hostile runtime and public-helper adversarial suite (RED)** - `01faefa` (test)
3. **Task 2: Add the hostile runtime and public-helper adversarial suite (GREEN)** - `03b75e7` (feat)

**Plan metadata:** Pending final docs commit at summary creation time.

## Files Created/Modified

- `hsm_test_helpers_test.go` - Adds reusable ErrorEvent recording and immediate waiter-closure assertions for bounded hostile-path tests.
- `hsm_runtime_adversarial_test.go` - Adds deterministic adversarial coverage for concurrent panic recovery, process panic recovery, terminate timeout, and hostile helper inputs.
- `hsm.go` - Guards public helper context lookup seams so typed-nil contexts preserve missing-context behavior instead of panicking.

## Decisions Made

- Treated typed-nil `context.Context` values as missing helper context so `Dispatch`, `Set`, `Call`, `DispatchAll`, and `DispatchTo` stay aligned with existing nil or missing-context contracts.
- Kept timeout adversarial assertions on the established Phase 2 pattern of observing `ErrorEvent` effects after `Stop`, instead of forcing a post-shutdown state transition that the runtime does not guarantee.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed typed-nil helper context panics in public helper paths**
- **Found during:** Task 2 (Add the hostile runtime and public-helper adversarial suite)
- **Issue:** A typed-nil `context.Context` caused `FromContext` and `DispatchTo` helper paths to dereference nil and panic instead of preserving the existing missing-context contract.
- **Fix:** Added nil guards to `FromContext`, `InstancesFromContext`, and `DispatchTo`, which restores closed-waiter or `ErrMissingHSM` behavior for hostile helper calls.
- **Files modified:** hsm.go
- **Verification:** `GOCACHE=$(pwd)/.cache/go-build go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial'`
- **Committed in:** `03b75e7`

**2. [Rule 3 - Blocking] Redirected the Go build cache for local verification**
- **Found during:** Task 1 and Task 2 verification
- **Issue:** The default Go build cache path was unavailable in the execution environment.
- **Fix:** Ran verification with `GOCACHE=$(pwd)/.cache/go-build`.
- **Files modified:** None
- **Verification:** `GOCACHE=$(pwd)/.cache/go-build go test ./... -run 'TestRuntimeStopTimeoutDispatchesErrorEventDeterministically|TestRuntimeRulesConformance'` and `GOCACHE=$(pwd)/.cache/go-build go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial'`
- **Committed in:** None

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking)
**Impact on plan:** Both deviations were narrow and necessary to make the adversarial suite truthful and runnable. No alternate harness or broader runtime redesign was introduced.

## Issues Encountered

- The initial RED fixtures needed to mirror the runtime’s actual transition-shape and stop-timeout contracts before they isolated the intended helper panic; the final suite now reflects those established semantics directly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 4 now has deterministic hostile runtime evidence in the root package, so the property and regression-gate plans can build on explicit adversarial contracts rather than exploratory debugging.
- The public helper surface now behaves predictably for typed-nil contexts, removing a brittle edge case that would have destabilized later property-style coverage.

## Self-Check: PASSED

- FOUND: `.planning/phases/04-adversarial-regression-matrix/04-01-SUMMARY.md`
- FOUND: `e1c6166`
- FOUND: `01faefa`
- FOUND: `03b75e7`
