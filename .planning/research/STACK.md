# Research: Verification Stack for hsm.go

## Context

This is brownfield verification hardening for an existing Go library workspace. The goal is not product feature expansion; it is to establish a high-assurance verification stack for `hsm.go`, `kind`, and `muid`, with release-gating confidence across public APIs, failure paths, concurrency, timing, and rule conformance.

## Recommended Core Stack

### Toolchain Baseline

- **Go toolchain**: align all verification work to the repo’s declared version in `go.mod`, `go.work`, and `.tool-versions`
- **Primary runner**: `go test`
- **Race detector**: `go test -race`
- **Fuzzing**: native Go fuzzing via `testing.F` and `go test -fuzz=...`
- **Unit/package coverage**: `go test -coverprofile=...`
- **Integration/runtime coverage**: `go build -cover`, `GOCOVERDIR`, and `go tool covdata`
- **Benchmarking**: `go test -bench=...`
- **Dependency security**: `govulncheck`

## Why This Stack

- It stays inside the supported Go toolchain rather than introducing a large third-party verification surface
- Native fuzzing is coverage-guided and first-class in Go
- The race detector is the standard mechanism for surfacing real concurrency bugs under exercised workloads
- Integration coverage support allows runtime paths outside normal unit tests to be measured
- `govulncheck` closes a separate but relevant release-gating risk: vulnerable dependencies

## Recommended Supporting Artifacts

- Coverage merge/report scripts under a dedicated verification directory
- CI jobs for:
- package tests
- race tests
- fuzz smoke/regression runs
- integration coverage runs
- benchmark regression checks
- vulnerability scans
- Golden regression fixtures for known edge cases and failures
- `testdata/fuzz/` corpora for durable fuzz regressions

## What Not To Use As The Foundation

- Do not rely on statement coverage alone as a proxy for exhaustive verification
- Do not treat benchmarks as performance evidence unless they are versioned, repeatable, and gated
- Do not depend primarily on external fuzzing frameworks when native Go fuzzing already fits the language/runtime
- Do not make race detection optional for release paths if concurrency correctness is part of the trust promise

## Confidence

- **Go-native test/race/fuzz/coverage tooling**: High confidence
- **Integration coverage using `go build -cover` and `GOCOVERDIR`**: High confidence
- **Third-party specialized harnesses**: Medium confidence, only where native tooling proves insufficient

## Implications For hsm.go

- Most effort should go into deterministic harness design and coverage accounting, not into choosing fancy tools
- The main risk is test architecture and discipline, not lack of tooling
- The roadmap should probably introduce verification in layers: deterministic harnesses first, then exhaustive suites and gates
