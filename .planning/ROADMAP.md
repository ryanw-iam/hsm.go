# Roadmap: hsm.go Verification Hardening

## Overview

This roadmap raises the existing `hsm.go` workspace to a release-gated verification standard by first establishing exhaustive baseline coverage across `hsm`, `kind`, and `muid`, then proving deterministic runtime behavior, then codifying `rules.md` as executable conformance checks, then adding adversarial regression coverage, and finally remediating every issue exposed by the new bar.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Coverage Baseline And Release Gates** - Establish exhaustive path coverage and make core verification mandatory for shipping. Completed 2026-03-13.
- [x] **Phase 2: Deterministic Runtime Semantics** - Prove concurrency, timing, ordering, and race-sensitive behavior deterministically. Completed 2026-03-13.
- [x] **Phase 3: Rules Conformance Matrix** - Turn `rules.md` into executable conformance checks for valid and invalid models. (completed 2026-03-13)
- [x] **Phase 4: Adversarial Regression Matrix** - Add fault injection, fuzz/property checks, and compatibility/performance regression coverage. Completed 2026-03-13.
- [ ] **Phase 5: Hardening Remediation Closure** - Fix issues exposed by the new verification bar and close the release gate cleanly.

## Phase Details

### Phase 1: Coverage Baseline And Release Gates
**Goal**: The workspace has exhaustive baseline path coverage and a release-gated verification entry point across `hsm`, `kind`, and `muid`.
**Depends on**: Nothing (first phase)
**Requirements**: VER-01, VER-02
**Success Criteria** (what must be TRUE):
  1. Public APIs in `hsm`, `kind`, and `muid` are exercised by repeatable tests that cover normal, branch, error, and panic-recovery behavior.
  2. A clean checkout can run the required verification suite and get a definitive pass/fail result that blocks release on failures.
  3. Helper-module behavior and critical runtime branches are included in the same gating evidence rather than treated as advisory side tests.
  4. Verification output makes path coverage gaps and failing release criteria obvious enough to drive follow-up work.
**Plans**: 3 plans

Plans:
- [x] `01-01-PLAN.md` - Restructure root `hsm` coverage into explicit runtime and observer path suites.
- [x] `01-02-PLAN.md` - Bring `kind` and `muid` to the baseline bar and close the exposed `muid` gate failure.
- [x] `01-03-PLAN.md` - Create the canonical verification command, path matrix evidence, and release-blocking workflow wiring.

### Phase 2: Deterministic Runtime Semantics
**Goal**: Runtime behavior remains deterministic under concurrency, timing, dispatch ordering, and race-sensitive conditions.
**Depends on**: Phase 1
**Requirements**: VER-03
**Success Criteria** (what must be TRUE):
  1. Concurrent dispatch and state transitions can be exercised repeatedly with stable, asserted outcomes.
  2. Timer-driven, cancellation, and ordering-sensitive behaviors can be verified without flaky sleeps or nondeterministic expectations.
  3. Race-enabled verification passes on the runtime scenarios that matter for production use.
  4. Post-dispatch assertions observe settled machine state through deterministic synchronization points instead of incidental timing.
**Plans**: 3 plans

Plans:
- [x] `02-01-PLAN.md` - Build deterministic timer and waiter harnesses, then move timer-driven runtime assertions into focused suites.
- [x] `02-02-PLAN.md` - Add deterministic concurrent dispatch, queue-ordering, and multi-instance targeting coverage.
- [x] `02-03-PLAN.md` - Lock lifecycle timeout and restart semantics behind deterministic tests and close the race-enabled gate.

### Phase 3: Rules Conformance Matrix
**Goal**: The modeling contract in `rules.md` is enforced by executable conformance coverage.
**Depends on**: Phase 2
**Requirements**: VER-04
**Success Criteria** (what must be TRUE):
  1. The documented rule set is represented by executable tests that demonstrate compliant usage patterns passing.
  2. Invalid model definitions or runtime usages that violate documented rules fail in clear, repeatable ways.
  3. Rule conformance evidence is maintained in-repo so regressions against the documented contract are caught before release.
**Plans**: 4 plans

Plans:
- [x] `03-01-PLAN.md` - Publish the checked-in rules matrix and shared conformance helper contract.
- [x] `03-02-PLAN.md` - Build the dedicated define/build-time conformance suite for invalid models.
- [x] `03-03-PLAN.md` - Build the deterministic runtime/exemplar conformance suite for semantic and exemplar rules.
- [x] `03-04-PLAN.md` - Reconcile matrix evidence to the implemented suites and prove the full canonical gate.

### Phase 4: Adversarial Regression Matrix
**Goal**: The verification suite covers hostile, malformed, and regression-prone conditions beyond nominal functional tests.
**Depends on**: Phase 3
**Requirements**: VER-05
**Success Criteria** (what must be TRUE):
  1. Fault-injected runtime and model failures can be triggered deliberately and produce asserted outcomes instead of ad hoc debugging.
  2. Fuzz or property-style checks preserve core invariants for targeted input and state-path surfaces.
  3. Compatibility or performance regressions that matter to downstream consumers are surfaced by automated checks before release.
  4. Added adversarial coverage improves confidence without introducing flaky or unbounded standard CI behavior.
**Plans**: 3 plans

Plans:
- [x] `04-01-PLAN.md` - Add deterministic fault-injection coverage for hostile runtime and public-helper failures.
- [x] `04-02-PLAN.md` - Add bounded root and `muid` property checks with checked-in fuzz seeds.
- [x] `04-03-PLAN.md` - Add allocation and benchmark-smoke regression evidence, then wire the canonical adversarial gate.

### Phase 5: Hardening Remediation Closure
**Goal**: Issues exposed by the verification matrix are fixed and locked in as permanent regression coverage.
**Depends on**: Phase 4
**Requirements**: VER-06
**Success Criteria** (what must be TRUE):
  1. Previously failing correctness, determinism, observability, or maintainability scenarios now pass with permanent regression protection.
  2. Diagnostics and test structure are clear enough that a failing verification result points directly to the broken behavior family.
  3. No known v1 verification-blocking defects discovered in earlier phases remain open when the release gate is run end to end.
**Plans**: 3 plans

Plans:
- [ ] `05-01-PLAN.md` - Align `Stop(nil)` with the hardened hostile-helper contract and lock the seam behind focused regression coverage.
- [ ] `05-02-PLAN.md` - Align `rules.md`, exported helper comments, and `README.md` around the hardened public synchronization contract.
- [ ] `05-03-PLAN.md` - Make the canonical gate phase-neutral, track contract artifacts cleanly, remove stale legacy CI, and prove the final release gate.

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Coverage Baseline And Release Gates | 3/3 | Complete | 2026-03-13 |
| 2. Deterministic Runtime Semantics | 3/3 | Complete | 2026-03-13 |
| 3. Rules Conformance Matrix | 4/4 | Complete   | 2026-03-13 |
| 4. Adversarial Regression Matrix | 3/3 | Complete | 2026-03-13 |
| 5. Hardening Remediation Closure | 0/3 | Planned | - |
