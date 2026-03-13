# Roadmap: hsm.go Verification Hardening

## Overview

This roadmap hardens the existing `hsm.go` workspace into a release-gated dependency for a production event-driven system. The sequence starts by building deterministic verification infrastructure, expands coverage across nominal and adverse runtime behavior, then closes with evidence and CI gates that make the new assurance level durable.

## Phases

**Phase Numbering:**
- Integer phases are planned milestone work.
- Decimal phases are reserved for inserted urgent fixes if verification exposes blocking risk.

- [ ] **Phase 1: Harness And Determinism Foundation** - Build reusable, deterministic verification infrastructure for exhaustive path testing.
- [ ] **Phase 2: Core Runtime Path Coverage** - Exhaustively test public APIs and internal runtime branches in `hsm`.
- [ ] **Phase 3: Failure Recovery And Rules Conformance** - Lock down failure semantics, recovery behavior, and `rules.md` conformance.
- [ ] **Phase 4: Concurrency And Timing Verification** - Validate race-sensitive, ordering-sensitive, and stress paths.
- [ ] **Phase 5: Helper Module Exhaustiveness** - Bring `kind` and `muid` to the same verification bar as the main runtime.
- [ ] **Phase 6: Evidence And Release Gates** - Turn verification into durable release policy with visibility into residual risk.

## Phase Details

### Phase 1: Harness And Determinism Foundation
**Goal**: Create the shared fixtures, assertions, and fault-injection mechanisms needed to test every meaningful code path without flake-prone tests.
**Depends on**: Nothing (first phase)
**Requirements**: [HARN-01, HARN-02, HARN-03]
**Success Criteria** (what must be TRUE):
  1. Contributors can write path-oriented tests against reusable harness helpers instead of duplicating setup logic.
  2. Timing-sensitive and ordering-sensitive paths can be exercised without relying on unbounded sleeps.
  3. Error, panic, and hard-to-reach branches have an agreed mechanism for deterministic verification.
**Plans**: 3 plans

Plans:
- [ ] 01-01: Inventory current test seams and design deterministic harness APIs
- [ ] 01-02: Implement shared fixtures, observers, and scheduling/fault controls
- [ ] 01-03: Prove the harness on representative runtime branches and document usage

### Phase 2: Core Runtime Path Coverage
**Goal**: Exhaustively cover the public contract and internal branch behavior of the `hsm` package.
**Depends on**: Phase 1
**Requirements**: [CORE-01, CORE-02, CORE-03]
**Success Criteria** (what must be TRUE):
  1. Public `hsm` APIs and model-building paths have explicit tests for normal, edge, and regression cases.
  2. Dispatch, transition, guard, action, subscription, and emitted-output branches are behaviorally asserted.
  3. Any defects found while expanding coverage are fixed with permanent regression tests.
**Plans**: 3 plans

Plans:
- [ ] 02-01: Expand API and model-construction coverage across nominal and edge paths
- [ ] 02-02: Cover internal runtime branches and previously unobserved execution paths
- [ ] 02-03: Fix surfaced defects and lock them with regression tests

### Phase 3: Failure Recovery And Rules Conformance
**Goal**: Verify that invalid inputs, recovery behavior, and documented rules behave exactly as claimed.
**Depends on**: Phase 2
**Requirements**: [FAIL-01, FAIL-02, FAIL-03, RULE-01, RULE-02]
**Success Criteria** (what must be TRUE):
  1. Invalid definitions, illegal transitions, and bad runtime inputs fail in asserted and documented ways.
  2. Recovery, shutdown, cancellation, and partial-failure flows preserve expected cleanup and observability semantics.
  3. Every enforceable rule in `rules.md` is covered by conformance tests or explicitly documented as non-testable.
**Plans**: 3 plans

Plans:
- [ ] 03-01: Add invalid-input and illegal-transition verification matrix
- [ ] 03-02: Exercise panic, recovery, shutdown, and partial-failure behavior
- [ ] 03-03: Build and document `rules.md` conformance suite

### Phase 4: Concurrency And Timing Verification
**Goal**: Establish confidence in concurrent and scheduling-sensitive behavior under race-enabled and stress-oriented execution.
**Depends on**: Phase 3
**Requirements**: [CONC-01, CONC-02, CONC-03]
**Success Criteria** (what must be TRUE):
  1. Concurrent dispatch and asynchronous execution paths run in race-enabled suites that are stable enough for CI.
  2. Ordering and timing semantics are validated through deterministic or bounded-repeat checks rather than flaky timing assumptions.
  3. Stress-style scenarios expose nondeterministic failures early enough to block unsafe releases.
**Plans**: 3 plans

Plans:
- [ ] 04-01: Add race-enabled concurrent dispatch and async behavior suites
- [ ] 04-02: Validate ordering and timing semantics with deterministic stress harnesses
- [ ] 04-03: Fix and regression-test any nondeterministic failures found

### Phase 5: Helper Module Exhaustiveness
**Goal**: Bring `kind` and `muid` coverage, regression depth, and misuse resilience up to the same standard as `hsm`.
**Depends on**: Phase 4
**Requirements**: [HELP-01, HELP-02, HELP-03]
**Success Criteria** (what must be TRUE):
  1. `kind` behaviors are exhaustively verified across hierarchy, encoding, and edge cases.
  2. `muid` behaviors are exhaustively verified across generation, parsing, ordering, and boundary conditions.
  3. High-risk helper behaviors are reinforced with regression, fuzz, or benchmark coverage where appropriate.
**Plans**: 3 plans

Plans:
- [ ] 05-01: Audit and expand `kind` verification to exhaustive branch coverage
- [ ] 05-02: Audit and expand `muid` verification to exhaustive branch coverage
- [ ] 05-03: Add helper-focused fuzz, regression, and benchmark checks where justified

### Phase 6: Evidence And Release Gates
**Goal**: Convert the verification work into a durable release-gating system with clear evidence of readiness and remaining gaps.
**Depends on**: Phase 5
**Requirements**: [EVID-01, EVID-02, EVID-03]
**Success Criteria** (what must be TRUE):
  1. Coverage evidence shows which runtime paths, packages, and verification modes are exercised, not just a single summary percentage.
  2. CI gates race, fuzz, compatibility, benchmark, and vulnerability checks where they materially reduce production risk.
  3. Maintainers can tell at a glance whether the workspace is release-ready and what gaps remain.
**Plans**: 3 plans

Plans:
- [ ] 06-01: Define verification evidence model and coverage accounting outputs
- [ ] 06-02: Wire release-gating CI for race, fuzz, compatibility, benchmark, and vuln checks
- [ ] 06-03: Publish readiness matrix and close remaining verification gaps

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Harness And Determinism Foundation | 0/3 | Not started | - |
| 2. Core Runtime Path Coverage | 0/3 | Not started | - |
| 3. Failure Recovery And Rules Conformance | 0/3 | Not started | - |
| 4. Concurrency And Timing Verification | 0/3 | Not started | - |
| 5. Helper Module Exhaustiveness | 0/3 | Not started | - |
| 6. Evidence And Release Gates | 0/3 | Not started | - |
