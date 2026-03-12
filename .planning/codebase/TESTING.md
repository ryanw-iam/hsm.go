# hsm.go Testing

## Overview

Testing in `hsm.go` is split between the main runtime package and the two helper modules. The suite mixes correctness tests, race-enabled CI runs, and benchmarks.

## Main Package Tests

- Primary test file: `hsm_test.go`
- Benchmark file: `hsm_bench_test.go`
- The main test surface appears integration-oriented, covering package behavior through public APIs rather than deep internal unit slices

## Helper Module Tests

### `kind`

- Test file: `kind/kind_test.go`
- Coverage appears minimal relative to the role of the module

### `muid`

- Test file: `muid/muid_test.go`
- Benchmark file: `muid/muid_bench_test.go`
- `muid` appears to have stronger dedicated test and performance coverage than `kind`

## CI

- Primary current CI file: `.github/workflows/unit-tests.yml`
- GitHub Actions runs:
- `go test -v ./...`
- `go test -race -short ./...`
- CI currently sets up Go `1.22`

## Legacy CI

- `.travis.yml` still exists and references:
- Go `1.16` and `1.17`
- `goveralls`
- `gocov`
- `cover`

This legacy config is materially older than the current module/tool version pins.

## Test Strategy Signals

- The project relies on standard Go test tooling rather than custom harnesses
- Race detection is part of the active GitHub Actions flow
- Benchmarks are treated as first-class coverage for both the core runtime and ID helper module

## Testing Risks

- There is version drift between declared Go versions in `go.mod` and CI versions in `.github/workflows/unit-tests.yml` and `.travis.yml`
- Module test depth is uneven across `hsm`, `kind`, and `muid`
- The main implementation is highly centralized in `hsm.go`, so narrow unit isolation may be harder than integration-style testing
