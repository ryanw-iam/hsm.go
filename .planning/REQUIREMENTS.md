# Requirements: hsm.go Verification Hardening

## Scope

Brownfield verification hardening for the `hsm.go` workspace only. These v1 requirements are derived from the active scope in `.planning/PROJECT.md` and apply to `hsm.go`, `kind/`, `muid/`, and supporting verification/release automation inside this repository.

## v1 Requirements

### Coverage And Release Gates

- [x] **VER-01**: Achieve exhaustive path-oriented test coverage for the `hsm.go` workspace, including `hsm`, `kind`, and `muid`.
- [x] **VER-02**: Add release-gating verification for public APIs, runtime branches, error paths, panic-recovery paths, and helper-module behavior.

### Runtime Determinism

- [x] **VER-03**: Add deterministic validation for concurrency, timing, dispatch ordering, and race-sensitive behavior.

### Rules Conformance

- [x] **VER-04**: Add conformance coverage for the library rules captured in `rules.md`.

### Adversarial Regression

- [ ] **VER-05**: Add fault-injection, fuzz/property-style, and compatibility/performance regression checks where needed to close verification gaps.

### Remediation

- [x] **VER-06**: Fix any correctness, determinism, observability, or maintainability issues uncovered by the new verification bar.

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| VER-01 | Phase 1 | Complete |
| VER-02 | Phase 1 | Complete |
| VER-03 | Phase 6 | Complete |
| VER-04 | Phase 7 | Complete |
| VER-05 | Phase 8 | Pending |
| VER-06 | Phase 5 | Complete |

`VER-03` stays mapped to Phase 6 in this table because the closure work was the Phase 6 backfill of the missing Phase 2 verification artifact at `.planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md`, not a new runtime implementation change.
