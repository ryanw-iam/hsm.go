# hsm.go Stack

## Overview

`hsm.go` is a Go workspace centered on a hierarchical state machine runtime and DSL. The main package lives at the repository root of this subtree, with two local helper modules for kind encoding and ID generation.

## Core Toolchain

- Language: Go
- Go version in `go.mod`: `1.25.3`
- Go version in `go.work`: `1.25.3`
- Tool pin in `.tool-versions`: `golang 1.25.3`
- Module root: `go.mod`
- Workspace file: `go.work`

## Modules

- Main module: `github.com/stateforward/hsm.go` in `go.mod`
- Helper module: `github.com/stateforward/hsm.go/kind` in `kind/go.mod`
- Helper module: `github.com/stateforward/hsm.go/muid` in `muid/go.mod`

## Local Wiring

- The main module requires `kind` and `muid` as versioned modules.
- `go.mod` replaces both with local paths:
- `github.com/stateforward/hsm.go/kind => ./kind`
- `github.com/stateforward/hsm.go/muid => ./muid`
- `go.work` includes `.`, `./kind`, and `./muid`

## Main Package

- Primary implementation file: `hsm.go`
- Public docs and usage examples: `README.md`
- Benchmarks: `hsm_bench_test.go`
- Main behavior tests: `hsm_test.go`

## Internal Libraries

### `kind`

- Path: `kind/kind.go`
- Purpose: bit-packed type/kind hierarchy helpers
- Docs: `kind/README.md`
- Tests: `kind/kind_test.go`

### `muid`

- Path: `muid/muid.go`
- Purpose: monotonically unique ID generator
- Docs: `muid/README.md`
- Tests: `muid/muid_test.go`, `muid/muid_bench_test.go`

## Dependencies

- Main module runtime deps are only the two local helper modules.
- `muid` pulls external libraries through `muid/go.mod` and `muid/go.sum`.
- External ID-related libraries documented in the module state include `nanoid`, `google/uuid`, and `oklog/ulid`.

## CI And Release Tooling

- GitHub Actions workflows live under `.github/workflows`
- Release config lives in `.goreleaser.yaml`
- Legacy Travis config exists in `.travis.yml`

## Runtime Characteristics

- The main implementation uses standard library packages such as `context`, `sync`, `sync/atomic`, `time`, `reflect`, and `log/slog`.
- The package provides a declarative HSM DSL plus a runtime with event dispatch, attribute updates, operation calls, observers, and grouping helpers.
