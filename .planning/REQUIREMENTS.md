# Requirements: hsm.go Verification Hardening

## Overview

This requirements set defines the verification bar for the existing `hsm.go` workspace. The objective is not feature expansion. The objective is to make the `hsm`, `kind`, and `muid` packages release-gated through exhaustive, deterministic, and maintainable verification of all meaningful runtime paths.

## v1 Requirements

### Harness Foundation

| ID | Requirement | Why It Matters | Traceability |
|----|-------------|----------------|--------------|
| HARN-01 | The workspace must provide reusable deterministic test harnesses for state transitions, observer flows, emitted events, and action execution. | Exhaustive verification is not sustainable without shared fixtures and assertions. | Phase 1 |
| HARN-02 | Timing-sensitive and order-sensitive paths must be testable without brittle sleep-based tests. | Concurrency and timer behavior must be validated without introducing flakiness. | Phase 1 |
| HARN-03 | The verification layer must support branch forcing or equivalent fault injection where natural reachability is insufficient. | Error, panic, and recovery paths still need direct assertions. | Phase 1 |

### Core Runtime Coverage

| ID | Requirement | Why It Matters | Traceability |
|----|-------------|----------------|--------------|
| CORE-01 | Public `hsm` package APIs and declarative model construction paths must have exhaustive success-path and edge-path tests. | Downstream systems depend on the public contract, not just internal implementation details. | Phase 2 |
| CORE-02 | Internal runtime branches covering dispatch, transitions, actions, guards, subscriptions, and emitted outputs must have explicit behavioral coverage. | Every meaningful execution path must be observable and locked down. | Phase 2 |
| CORE-03 | Existing and newly discovered correctness defects must receive regression tests before closure. | The verification effort must permanently capture defects it finds. | Phase 2 |

### Failure And Recovery

| ID | Requirement | Why It Matters | Traceability |
|----|-------------|----------------|--------------|
| FAIL-01 | Invalid model definitions, invalid runtime inputs, and illegal state transitions must surface asserted failures or documented behavior. | High-assurance consumers need predictable failure semantics. | Phase 3 |
| FAIL-02 | Panic and recovery paths must be exercised and asserted where the runtime intentionally recovers or propagates. | Safety claims are incomplete if panic behavior is unverified. | Phase 3 |
| FAIL-03 | Shutdown, cancellation, and partially failed execution flows must be verified for cleanup and observability guarantees. | Production systems fail in partial ways, not just idealized ones. | Phase 3 |

### Rule Conformance

| ID | Requirement | Why It Matters | Traceability |
|----|-------------|----------------|--------------|
| RULE-01 | Each enforceable rule in `rules.md` must have explicit conformance coverage or a documented rationale if it is non-testable. | The documented contract must match actual runtime behavior. | Phase 3 |
| RULE-02 | Rule violation cases must assert the expected rejection, failure mode, or diagnostic behavior. | A conformance suite is incomplete if only valid cases are tested. | Phase 3 |

### Concurrency And Timing

| ID | Requirement | Why It Matters | Traceability |
|----|-------------|----------------|--------------|
| CONC-01 | Concurrent dispatch, asynchronous behavior execution, and race-prone shared paths must run under race-enabled test suites. | The library is used in an event-driven production system. | Phase 4 |
| CONC-02 | Ordering, timing, and scheduling-sensitive semantics must be validated with deterministic or bounded-repeat tests. | Concurrency assertions are only useful if the tests are trustworthy. | Phase 4 |
| CONC-03 | Stress-style verification must exercise realistic concurrent workloads and surface nondeterministic failures early. | High-assurance confidence requires more than single-threaded examples. | Phase 4 |

### Helper Modules

| ID | Requirement | Why It Matters | Traceability |
|----|-------------|----------------|--------------|
| HELP-01 | The `kind` module must have exhaustive coverage for encoding, hierarchy, comparisons, and edge cases. | `kind` is foundational to state classification behavior. | Phase 5 |
| HELP-02 | The `muid` module must have exhaustive coverage for ID generation, parsing, formatting, ordering, and boundary conditions. | `muid` contributes identity behavior used outside nominal paths. | Phase 5 |
| HELP-03 | Helper-module regressions, fuzz targets, and benchmarks must exist where risk justifies them. | Verification must include both correctness and misuse resilience. | Phase 5 |

### Evidence And Release Gates

| ID | Requirement | Why It Matters | Traceability |
|----|-------------|----------------|--------------|
| EVID-01 | Coverage evidence must account for package-level and runtime-driven paths rather than relying on a single headline percentage. | Coverage numbers alone can mask missing behavioral paths. | Phase 6 |
| EVID-02 | Race, fuzz, benchmark, compatibility, and vulnerability checks must be wired into CI as release gates where they materially reduce risk. | Verification must affect shipping decisions, not sit as optional local work. | Phase 6 |
| EVID-03 | The project must produce a maintainable verification matrix that makes uncovered gaps and release readiness obvious. | Enterprise consumers need a durable signal, not one-off test additions. | Phase 6 |

## v2 Candidates

| ID | Candidate | Why Deferred |
|----|-----------|--------------|
| FORM-01 | Model checking or formally specified transition invariants for critical HSM behaviors. | Higher cost than the initial verification hardening milestone. |
| CERT-01 | External certification-oriented evidence pack and audit artifacts. | Not requested for this milestone; engineering rigor comes first. |

## Out Of Scope

- Adding new public library features unrelated to verification hardening
- Modifying sibling implementations outside the `hsm.go` workspace
- Producing formal compliance paperwork or certification submissions

## Acceptance Standard

The milestone is only complete when all v1 requirements are mapped to executed phases, all material code paths in `hsm`, `kind`, and `muid` are covered by deterministic verification, and the resulting test and CI matrix is strong enough to gate releases for production use.
