# Research: Verification Pitfalls for hsm.go

## Pitfall 1: Confusing Coverage Percentage With Trust

- **Warning signs**: high statement coverage but low failure-path or concurrency confidence
- **Prevention**: track path categories explicitly, not just a single percentage
- **Phase fit**: early verification architecture and requirements phases

## Pitfall 2: Flaky Timing Tests

- **Warning signs**: tests rely on sleeps, wall-clock timing, or scheduler luck
- **Prevention**: build deterministic harnesses and isolate timing assumptions
- **Phase fit**: harness foundation phase

## Pitfall 3: Race Detector Without Real Workloads

- **Warning signs**: `-race` is green only on tiny happy-path tests
- **Prevention**: run race-enabled suites against realistic concurrent paths and stress scenarios
- **Phase fit**: concurrency verification phase

## Pitfall 4: Fuzzing Without Reproducible Regression Value

- **Warning signs**: fuzzing burns CI time without curated corpora or resulting regression tests
- **Prevention**: keep seed corpora, minimize failures, and convert findings into stable test assets
- **Phase fit**: fuzz/property phase

## Pitfall 5: Untestable Design Gaps Are Left In Place

- **Warning signs**: hidden nondeterminism, giant helper-free tests, or inability to force critical branches
- **Prevention**: allow production refactors when the verification bar exposes testability limits
- **Phase fit**: code-hardening phases

## Pitfall 6: Helper Modules Are Ignored

- **Warning signs**: main package quality rises while `kind` and `muid` remain thinly tested
- **Prevention**: treat the workspace as one release unit with module-specific evidence
- **Phase fit**: requirements and roadmap coverage mapping

## Pitfall 7: Legacy CI Drift Persists

- **Warning signs**: different Go versions and conflicting automation paths remain active
- **Prevention**: align CI, toolchain, and release gates before trusting results
- **Phase fit**: CI/release hardening phase
