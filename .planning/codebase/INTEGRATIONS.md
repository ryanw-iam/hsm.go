# hsm.go Integrations

## Overview

`hsm.go` is primarily a library package rather than an application. Its integrations are mostly local module links, CI/release automation, and ID-generation dependencies in the `muid` helper module.

## Local Module Integrations

- `go.mod` depends on local modules `kind` and `muid`
- `go.work` binds the three modules into one workspace
- `go.mod` uses `replace` directives to point to `./kind` and `./muid`

## External Package Integrations

The main `hsm` package has no direct third-party runtime dependencies in `go.mod`. External dependency usage is concentrated in `muid`.

### `muid` module

- Module file: `muid/go.mod`
- External libraries include:
- `github.com/matoous/go-nanoid/v2`
- `github.com/google/uuid`
- `github.com/oklog/ulid/v2`

These libraries support alternate unique-ID generation strategies or related encoding/comparison logic around MUID generation.

## GitHub Integrations

- CI workflows are defined in `.github/workflows/unit-tests.yml`
- Release-related workflows exist in `.github/workflows/auto-release.yml`, `.github/workflows/release.yml`, and `.github/workflows/version-check.yml`
- `.goreleaser.yaml` targets GitHub releases for `stateforward/hsm-go`

## Toolchain Integrations

- `.tool-versions` integrates with asdf-style local tool management
- `go test` and `go test -race` are used in CI
- `.travis.yml` still references `goveralls`, `gocov`, and `cover`

## Runtime-Level Integrations

The main package integrates heavily with the Go standard library:

- `context` for lifecycle and cancellation
- `sync` and `sync/atomic` for runtime coordination
- `time` for timing/event scheduling
- `log/slog` for runtime logging
- `reflect` for operation dispatch and attribute comparison paths

## Absent Integrations

No evidence was found in this subtree for:

- SQL or NoSQL databases
- HTTP clients or servers
- cloud SDKs
- auth providers
- webhooks
- message brokers

## Integration Risk Notes

- CI metadata is split across modern GitHub Actions and legacy Travis config.
- The main module is dependency-light, but `muid` broadens the external surface area.
- Because the helper modules are local workspace members, changes in `kind` or `muid` directly affect the root package without a publish boundary during local development.
