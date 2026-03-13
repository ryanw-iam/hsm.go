# Research: Verification Features for hsm.go

## Table Stakes

### Core Correctness

- Full public API tests across `hsm`, `kind`, and `muid`
- Deterministic coverage of success and failure branches
- Regression tests for any known bug or panic path
- CI enforcement for test pass/fail and race detector status

### Concurrency And Timing

- Race-enabled test lanes
- Deterministic tests for dispatch ordering, cancellation, timers, and run-to-completion behavior
- Explicit coverage of scheduler-sensitive paths and concurrent behaviors

### Coverage And Evidence

- Coverage reports tracked in CI
- Branch/path gap accounting documented outside raw percentage alone
- Reproducible evidence for failure-path coverage

### Security And Supply Chain

- Dependency scanning with `govulncheck`
- Release gates that fail on unresolved high-severity dependency issues

## Differentiators

- Rule conformance suite derived from `rules.md`
- Path matrix showing which branches/panic paths/error paths are covered
- Fuzz targets for state-path resolution, event dispatch edge cases, identifier parsing, and malformed inputs
- Fault-injection harnesses for panic recovery, context cancellation, invalid model definitions, and runtime edge conditions
- Benchmark regression thresholds for latency-sensitive or allocation-sensitive paths

## Anti-Features

- Chasing 100% statement coverage while leaving concurrency or recovery behavior unverified
- Flaky timing tests that “usually pass” and erode trust in CI
- Huge monolithic golden tests that hide which path actually failed
- Unbounded fuzzing in standard CI where deterministic regression value is low

## Complexity Notes

- Public API coverage: Medium
- Failure and panic path coverage: High
- Deterministic concurrency/timing harnesses: High
- Rule conformance suite: Medium
- Fuzzing and seed corpus curation: Medium
- Coverage aggregation and path accounting: High
- Benchmark regression gating: Medium

## Dependency Notes

- Deterministic harnesses should exist before aggressive race/fuzz expansion
- Coverage accounting should come after or alongside test expansion, not before
- Rule conformance requirements depend on a stable interpretation of `rules.md`
- Benchmark gating should come after correctness and flake control
