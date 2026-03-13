# hsm.go Verification Hardening

## What This Is

This project initializes planning for the existing `hsm.go` library as a brownfield codebase. The immediate goal is to raise it to a production-grade verification standard for use inside an enterprise event-driven agentic IGA platform by testing every reachable code path in the `hsm.go` workspace and fixing anything that fails under that bar.

## Core Value

`hsm.go` must be a trustworthy, release-gated runtime dependency whose behavior is exhaustively verified across normal, edge, failure, concurrency, timing, and recovery paths.

## Requirements

### Validated

- ✓ The library provides a declarative hierarchical state machine DSL and runtime in `hsm.go` — existing
- ✓ The workspace includes reusable `kind` and `muid` helper modules used by the main package — existing
- ✓ The codebase already ships Go tests, benchmarks, CI workflows, and release metadata — existing

### Active

- [ ] Achieve exhaustive path-oriented test coverage for the `hsm.go` workspace, including `hsm`, `kind`, and `muid`
- [ ] Add release-gating verification for public APIs, runtime branches, error paths, panic-recovery paths, and helper-module behavior
- [ ] Add deterministic validation for concurrency, timing, dispatch ordering, and race-sensitive behavior
- [ ] Add conformance coverage for the library rules captured in `rules.md`
- [ ] Add fault-injection, fuzz/property-style, and compatibility/performance regression checks where needed to close verification gaps
- [ ] Fix any correctness, determinism, observability, or maintainability issues uncovered by the new verification bar

### Out of Scope

- New end-user product features for the library API — this milestone is about verification hardening, not expanding scope
- Sibling implementations outside `hsm.go` — this initialization is scoped to the Go workspace only
- Formal external certification artifacts or compliance paperwork — the goal is aerospace-grade rigor in engineering practice, not a certification package unless requested later

## Context

`hsm.go` is already in use as a dependency inside an enterprise-grade event-driven agentic IGA platform, so test quality is directly tied to downstream production risk. The codebase is brownfield, centered on a large `hsm.go` runtime file with helper modules in `kind/` and `muid/`, and already has a generated codebase map under `.planning/codebase/`. Current automation shows version drift between the declared Go toolchain and older CI metadata, which is relevant because the verification effort must become release-gating rather than advisory.

## Constraints

- **Scope**: Current directory only (`hsm.go` workspace) — sibling language ports are explicitly excluded
- **Quality**: Every code path must be tested — the library is a critical dependency in production
- **Behavior**: Concurrency, timing, race, and failure paths must be treated as first-class verification targets — runtime correctness matters more than nominal-path coverage
- **Change Policy**: Production code may be changed where tests expose gaps — verification hardening includes fixing what is found
- **Release**: Verification must become a release gate, not just additional documentation — the output must influence CI and shipping decisions

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Scope initialization to `hsm.go` only | The user explicitly wants the current directory and child directories only | — Pending |
| Treat this as brownfield hardening, not greenfield feature work | The library already exists and is already in production use | — Pending |
| Include all of public APIs, internal branches, panic/error paths, concurrency/timing behavior, and helper modules in scope | The user wants every code path tested and all verification categories release-gated | — Pending |
| Allow production fixes as part of the effort | New verification will surface defects or testability gaps that must be corrected | — Pending |

---
*Last updated: 2026-03-13 after initialization*
