# Research: Verification Architecture for hsm.go

## Goal

Design a verification system for an existing Go library workspace that can prove coverage across public APIs, internal branches, panic/error paths, concurrency/timing behavior, and helper modules.

## Major Components

### 1. Deterministic Test Harness Layer

- Shared helpers for clock/timer control where possible
- Shared event sequencing helpers
- Shared assertion helpers for state transitions, observer channels, and error outcomes
- Shared helpers for forcing edge-path entry without brittle test duplication

### 2. Behavioral Suite Layer

- Public API tests for nominal behavior
- Failure-path tests for invalid inputs and invalid model definitions
- Rule conformance tests derived from `rules.md`
- Cross-module tests for `kind` and `muid`

### 3. Adversarial Verification Layer

- Race-enabled package runs
- Fuzz targets with durable corpora
- Fault-injection and panic-recovery tests
- Timing and cancellation stress suites

### 4. Evidence Layer

- Coverage profile generation
- Integration coverage aggregation where package tests alone miss runtime paths
- Gap reports for unverified branches or known exclusions
- Benchmark baselines for regression detection

### 5. Release-Gating Layer

- CI workflow matrix
- Pass/fail thresholds and required jobs
- Artifact publication for coverage and regression evidence

## Data Flow

1. Test helpers define deterministic scaffolding
2. Behavioral and adversarial suites execute through `go test`, `-race`, and fuzz/integration flows
3. Coverage and benchmark artifacts are collected and normalized
4. CI evaluates policy gates
5. Failing gates feed back into code fixes or missing-test work

## Suggested Build Order

1. Stabilize deterministic harnesses for timing, ordering, and observer assertions
2. Expand public API and helper-module tests
3. Add explicit failure-path and panic-path coverage
4. Add `rules.md` conformance tests
5. Add race and concurrency stress lanes
6. Add fuzz targets and seed corpora
7. Add coverage aggregation and gap reporting
8. Add benchmark regression gates
9. Harden CI into a required release path

## Why This Build Order

- Deterministic helpers reduce flake before increasing test volume
- Correctness evidence should be in place before advanced gating
- Race/fuzz/stress testing without stable harnesses creates noise instead of confidence
- Coverage and benchmark policy is useful only after the suite meaningfully exercises the runtime
