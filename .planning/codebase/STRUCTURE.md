# hsm.go Structure

## Root Files

- `hsm.go`: main implementation
- `README.md`: package documentation and usage
- `rules.md`: domain-specific usage rules for model authors
- `version.go`: package version constant
- `go.mod`, `go.sum`: main module metadata
- `go.work`: workspace metadata
- `.tool-versions`: local tool pin
- `.goreleaser.yaml`: release config
- `.travis.yml`: legacy CI config

## Tests And Benchmarks

- `hsm_test.go`: main package tests
- `hsm_bench_test.go`: benchmarks for the main package

## Helper Modules

### `kind/`

- `kind/go.mod`: module metadata
- `kind/kind.go`: implementation
- `kind/kind_test.go`: tests
- `kind/version.go`: version constant
- `kind/README.md`: package docs

### `muid/`

- `muid/go.mod`, `muid/go.sum`: module metadata
- `muid/muid.go`: implementation
- `muid/muid_test.go`: tests
- `muid/muid_bench_test.go`: benchmarks
- `muid/version.go`: version constant
- `muid/README.md`: package docs

## Automation And Repo Metadata

- `.github/workflows/unit-tests.yml`
- `.github/workflows/auto-release.yml`
- `.github/workflows/release.yml`
- `.github/workflows/version-check.yml`
- `.gitignore`
- `.vscode/settings.json`

## Structural Notes

- This subtree is a standalone git repository rooted at `hsm.go/.git`
- The main runtime package is file-heavy but directory-light
- Secondary functionality is separated only where it forms a reusable module boundary
- The directory structure favors shallow navigation over deep package nesting

## Naming Patterns

- Test files follow standard Go `_test.go` naming
- Version constants are split into `version.go` files per module
- Supporting modules use short, purpose-specific names: `kind` and `muid`

## Practical Entry Points

- Start with `README.md` for API usage
- Read `rules.md` for authoring constraints
- Inspect `hsm.go` for runtime and DSL details
- Use `hsm_test.go` and `hsm_bench_test.go` to understand expected behavior and hot paths
