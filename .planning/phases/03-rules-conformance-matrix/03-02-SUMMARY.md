---
phase: 03-rules-conformance-matrix
plan: "02"
subsystem: testing
tags: [go, testing, rules, conformance, model-validation]
requires:
  - phase: 03-rules-conformance-matrix
    provides: checked-in rule matrix and stable panic-substring helper contract
provides:
  - Dedicated `TestModelRulesConformance` coverage for every `define_panic` matrix row
  - Define-time enforcement for state-behavior ownership, top-level transition endpoint pairing, and nested history placement
affects: [03-rules-conformance-matrix, verification, release-gate]
tech-stack:
  added: []
  patterns: [rule-id conformance subtests, stable panic-fragment assertions, direct-owner model validation]
key-files:
  created:
    - hsm_rules_model_conformance_test.go
  modified:
    - hsm.go
    - hsm_test.go
key-decisions:
  - "The model conformance suite follows the checked-in `define_panic` matrix rows one-to-one through `TestModelRulesConformance/HSMxx/...` targets."
  - "Legacy fixtures were updated to conform to HSM20 instead of preserving the old source-only top-level transition loophole."
patterns-established:
  - "Define-time contract coverage now lives in a dedicated root suite separate from general feature tests."
  - "Constructor ownership rules that promise direct placement enforcement should validate against the immediate stack owner, not any ancestor state."
requirements-completed: [VER-04]
duration: 3min
completed: 2026-03-13
---

# Phase 3 Plan 2: Model Rules Conformance Summary

**Dedicated define-time conformance suite for every `define_panic` matrix row, backed by tightened builder validation for ownership, history placement, and top-level transition shape**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-13T18:09:43Z
- **Completed:** 2026-03-13T18:12:32Z
- **Tasks:** 1
- **Files modified:** 3

## Accomplishments

- Added `hsm_rules_model_conformance_test.go` with a root `TestModelRulesConformance` suite and rule-ID subtests for every matrix row classified as `define_panic`.
- Tightened `hsm.go` so `Entry`, `Exit`, `Activity`, and `Defer` reject transition nesting, top-level transitions require both endpoints or neither, and history pseudostates cannot be declared at the model root.
- Aligned legacy complex-model fixtures with HSM20 so the full suite stays green under the stricter definition-time contract.

## Task Commits

Each task was committed atomically:

1. **Task 1: Build the define-time rules conformance suite (RED)** - `dd1eda5` (test)
2. **Task 1: Build the define-time rules conformance suite (GREEN)** - `cdb550e` (fix)

**Plan metadata:** Pending final docs commit at summary creation time.

## Files Created/Modified

- `hsm_rules_model_conformance_test.go` - Dedicated root model conformance suite covering every `define_panic` rule ID with stable panic-fragment assertions.
- `hsm.go` - Tightens define-time ownership and structure validation so the matrix-backed negative cases panic where promised.
- `hsm_test.go` - Moves legacy `/s` internal-transition fixtures inside state-owned definitions to match the enforced HSM20 contract.

## Decisions Made

- Kept the suite rooted in the checked-in matrix instead of inferring coverage from scattered legacy validations, so `go test -run 'TestModelRulesConformance/HSMxx'` now points directly at contract evidence.
- Fixed the builder to match the matrix where the existing API still leaked invalid model shapes, instead of weakening the new conformance expectations to match old loopholes.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Enforced missing define-time ownership and placement rules**
- **Found during:** Task 1 (Build the define-time rules conformance suite)
- **Issue:** `Entry`, `Exit`, `Activity`, and `Defer` could still bind through an ancestor state from inside `Transition(...)`, top-level source-only transitions slipped past HSM20, and top-level history pseudostates were accepted despite the matrix classifying them as define-time failures.
- **Fix:** Added direct-owner validation for state behaviors, tracked explicit transition endpoints for top-level validation, and rejected root-level `ShallowHistory`/`DeepHistory`.
- **Files modified:** `hsm.go`
- **Verification:** `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance'` and `GOCACHE="$PWD/.cache/go-build" go test ./...`
- **Committed in:** `cdb550e`

**2. [Rule 3 - Blocking] Realigned legacy complex-model fixtures with HSM20**
- **Found during:** Task 1 (Build the define-time rules conformance suite)
- **Issue:** Existing complex-model fixtures relied on a now-invalid top-level source-only transition shape, which broke the full suite once HSM20 was enforced.
- **Fix:** Moved the internal `/s` transition definitions into state `s` in the legacy test models so behavior stayed the same while the model shape matched the rule contract.
- **Files modified:** `hsm_test.go`
- **Verification:** `GOCACHE="$PWD/.cache/go-build" go test ./...`
- **Committed in:** `cdb550e`

**3. [Rule 3 - Blocking] Redirected the Go build cache for verification**
- **Found during:** Task 1 (Build the define-time rules conformance suite)
- **Issue:** `go test` could not write to the default cache path in this environment.
- **Fix:** Re-ran verification with `GOCACHE="$PWD/.cache/go-build"`.
- **Files modified:** None
- **Verification:** `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance'` and `GOCACHE="$PWD/.cache/go-build" go test ./...`
- **Committed in:** None (execution-environment workaround only)

---

**Total deviations:** 3 auto-fixed (1 bug, 2 blocking)
**Impact on plan:** The extra fixes were required to make the matrix-backed suite honest and keep the existing verification corpus consistent with the newly enforced rule contract.

## Issues Encountered

- The first HSM42 attempt used an initial pseudostate source and tripped the separate initial-trigger restriction first, so the final conformance case was rewritten to use a choice pseudostate source instead.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 03 can build runtime and exemplar conformance against the remaining `runtime_semantic` and `exemplar` rows without redefining the model-suite contract.
- Plan 04 can reconcile the matrix and canonical gate with direct evidence already present for every `define_panic` rule row.

## Self-Check: PASSED

- FOUND: `.planning/phases/03-rules-conformance-matrix/03-02-SUMMARY.md`
- FOUND: `dd1eda5`
- FOUND: `cdb550e`

---
*Phase: 03-rules-conformance-matrix*
*Completed: 2026-03-13*
