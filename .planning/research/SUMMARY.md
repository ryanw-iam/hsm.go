# Research Summary: hsm.go Verification Hardening

## Stack

The standard and lowest-risk stack for high-assurance verification of `hsm.go` is mostly native Go tooling: `go test`, `-race`, native fuzzing, coverage profiling, integration coverage with `GOCOVERDIR`, benchmarks, and `govulncheck`. The main engineering challenge is deterministic harness design and policy gating, not tool selection.

## Table Stakes

- Exhaustive public API and helper-module tests
- Deterministic failure-path coverage
- Race-enabled CI
- Coverage and regression evidence
- Dependency vulnerability scanning

## Watch Out For

- Coverage percentages creating false confidence
- Flaky timer/concurrency tests
- Race detection on unrealistic workloads
- Ignoring `kind` and `muid`
- Keeping legacy CI drift while claiming release-gating confidence

## Planning Implications

- Start with deterministic harnesses and branch forcing
- Then expand correctness and failure coverage
- Then add rule conformance, race, fuzz, and coverage accounting
- Finish by making the verification matrix release-gating
