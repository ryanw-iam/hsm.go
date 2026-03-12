# hsm.go Concerns

## High Priority

- The main runtime is heavily concentrated in `hsm.go`. That creates a large blast radius for changes and makes focused refactors harder.
- There is visible version drift in automation:
- `go.mod`, `go.work`, and `.tool-versions` pin Go `1.25.3`
- `.github/workflows/unit-tests.yml` runs Go `1.22`
- `.travis.yml` still references Go `1.16` and `1.17`

## Runtime Concerns

- `hsm.go` uses reflection in runtime-sensitive paths, including operation invocation and attribute comparison behavior
- The runtime also contains panic and recover handling in event-processing and concurrent-behavior paths
- That combination can complicate observability and performance debugging under load

## Testing Concerns

- `kind/kind_test.go` appears very light compared with the importance of the kind hierarchy to the rest of the package
- The main runtime package is well represented by large tests, but the monolithic source file means subtle regressions may still be hard to localize quickly

## Workflow Concerns

- `rules.md` is untracked according to `git status --short`
- If `rules.md` is intended to be part of the package contract, leaving it untracked risks convention drift between local workflow and committed source

## Structural Tradeoffs

- The subtree is clean at the directory level, but the main package is physically monolithic
- `kind` and `muid` are separated cleanly, yet the central `hsm` package still bundles DSL, model, runtime, helpers, and observability into one file

## Follow-Up Candidates

- Align CI/runtime version declarations across `go.mod`, `.tool-versions`, GitHub Actions, and any remaining Travis metadata
- Decide whether `rules.md` should be committed or explicitly treated as local-only workflow guidance
- Add or strengthen tests for `kind`
- Consider whether some runtime concerns in `hsm.go` should be split into smaller internal files without changing the public API
