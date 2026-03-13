---
phase: 04-adversarial-regression-matrix
plan: "02"
subsystem: testing
tags: [go, testing, fuzz, property, hsm, muid]
requires:
  - phase: 02-deterministic-runtime-semantics
    provides: waiter-based deterministic runtime assertions and settled-state observation patterns
  - phase: 03-rules-conformance-matrix
    provides: stable public-contract coverage patterns that avoid scheduler-order assertions
provides:
  - Root bounded fuzz target for deterministic HSM event-script invariants with checked-in hostile seeds
  - MUID generator property target for defaults, masking, monotonicity, overflow, and round-trip invariants
  - Checked-in seed corpora that replay under ordinary go test without enabling open-ended fuzzing
affects: [phase-04-canonical-gate, verification, release-gate]
tech-stack:
  added: []
  patterns: [seed-backed fuzz targets as deterministic regression carriers, reducer-based runtime property assertions]
key-files:
  created:
    - hsm_properties_fuzz_test.go
    - testdata/fuzz/FuzzHSMEventScriptProperties/seed-basic-script
    - testdata/fuzz/FuzzHSMEventScriptProperties/seed-nil-and-empty
    - muid/muid_fuzz_test.go
    - muid/testdata/fuzz/FuzzMUIDGeneratorProperties/seed-default-config
    - muid/testdata/fuzz/FuzzMUIDGeneratorProperties/seed-counter-boundary
  modified: []
key-decisions:
  - "Keep both property targets narrow by skipping oversized scripts or invalid bit layouts so ordinary go test replays remain deterministic."
  - "Model the root HSM property as a fixed public event script reducer instead of generating arbitrary models or asserting scheduler-sensitive details."
  - "Exercise MUID boundary behavior by forcing generator state to regressed-clock and counter-overflow conditions inside the test rather than adding open-ended fuzz loops."
patterns-established:
  - "Checked-in fuzz corpora should carry durable adversarial regressions while default verification stays on ordinary go test."
  - "Property targets should assert public invariants after each replay step, not exploratory throughput or internal remediation seams."
requirements-completed: [VER-05]
duration: 2 min
completed: 2026-03-13
---

# Phase 4 Plan 2: Bounded Property Seeds Summary

**Deterministic HSM and MUID property targets with checked-in fuzz seed corpora for settled-state and generator invariants**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-13T21:04:02Z
- **Completed:** 2026-03-13T21:06:06Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Added `FuzzHSMEventScriptProperties` with a reducer-backed root model that asserts dispatch/process waiter completion, settled state, and single terminal entry across deterministic hostile scripts.
- Added `FuzzMUIDGeneratorProperties` covering config defaulting, machine-ID masking, monotonic ID emission, regressed-clock handling, counter overflow, and base32 round-trips.
- Checked in root and `muid` seed corpora so ordinary `go test` replays adversarial cases without enabling unbounded `-fuzz` runs in the blocking path.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add a bounded root property target with checked-in hostile event seeds** - `3cf377b` (test, RED), `ccc9f81` (feat, GREEN)
2. **Task 2: Add MUID generator property targets and durable seed corpus** - `c29b562` (test, RED), `c3836aa` (feat, GREEN)

**Plan metadata:** Pending final docs commit at summary creation time.

## Files Created/Modified
- `hsm_properties_fuzz_test.go` - Root fuzz target that replays tiny event scripts against a fixed public HSM model and asserts deterministic invariants after every dispatch.
- `testdata/fuzz/FuzzHSMEventScriptProperties/seed-basic-script` - Normal root script corpus entry for `begin|finish`.
- `testdata/fuzz/FuzzHSMEventScriptProperties/seed-nil-and-empty` - Malformed-but-supported root script corpus entry with empty and duplicate event names.
- `muid/muid_fuzz_test.go` - Bounded `muid` property target for defaulting, masking, monotonicity, overflow, and string round-trips.
- `muid/testdata/fuzz/FuzzMUIDGeneratorProperties/seed-default-config` - Default-like `muid` configuration corpus entry.
- `muid/testdata/fuzz/FuzzMUIDGeneratorProperties/seed-counter-boundary` - Tight counter-width `muid` configuration corpus entry for overflow-path replay.

## Decisions Made
- Used a fixed reducer and public state path assertions for the root property target so failures reproduce from a single script string and never depend on internal scheduler order.
- Kept `muid` hostile coverage inside one bounded fuzz target by normalizing configs, skipping invalid bit layouts, and forcing boundary states explicitly after generator construction.
- Left deeper fuzzing as an opt-in workflow such as `GOCACHE=$(pwd)/.cache/go-build go test ./... ./muid/... -fuzz 'FuzzHSMEventScriptProperties|FuzzMUIDGeneratorProperties' -fuzztime=10s` instead of placing it on the default blocking path.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Plan 04-03 can wire the canonical adversarial gate to replay these checked-in seeds through ordinary `go test` without needing any extra harness.
- Any deeper exploratory fuzzing can stay opt-in with a bounded `-fuzztime`, preserving deterministic release verification.

## Self-Check: PASSED

- FOUND: `.planning/phases/04-adversarial-regression-matrix/04-02-SUMMARY.md`
- FOUND: `3cf377b`
- FOUND: `ccc9f81`
- FOUND: `c29b562`
- FOUND: `c3836aa`

---
*Phase: 04-adversarial-regression-matrix*
*Completed: 2026-03-13*
