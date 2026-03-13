# Phase 1 Research: Coverage Baseline And Release Gates

## Objective

Determine what Phase 1 must do to establish a real release-gated verification baseline across the full `hsm.go` workspace rather than only the root module.

## Current State

- The repository already has a root test suite with substantial behavioral coverage in `hsm_test.go`.
- `kind/kind_test.go` is effectively empty, so the helper module has almost no meaningful baseline.
- `muid/muid_test.go` is narrow, and `muid/muid_bench_test.go` contains non-benchmark tests with real behavioral assertions.
- Existing CI in `.github/workflows/unit-tests.yml` runs `go test -v ./...` and `go test -race -short ./...`.
- Legacy `.travis.yml` still references much older coverage tooling and Go versions, which is noise for a modern release gate.

## Key Findings

### 1. The current CI gate is not workspace-wide

Running `go test ./...` from the workspace root passes for the root module, but it does not provide a reliable full-workspace signal for nested modules `kind/` and `muid/`.

The stronger commands:

- `go test ./... ./kind/... ./muid/...`
- `go test -race -short ./... ./kind/... ./muid/...`

both exercise the helper modules explicitly and currently fail in `muid`.

### 2. There is already a failing helper-module behavior

The explicit full-workspace commands fail in `muid` with a failing sortability assertion in `muid_bench_test.go`. That means the current release signal is giving a false green because it is not checking the whole workspace the way the user expects.

### 3. Phase 1 needs test inventory structure, not just more assertions

`hsm_test.go` already holds useful trace helpers, panic assertions, and large integration scenarios. The gap is not the total absence of tests. The gap is that behavior families and uncovered branches are not surfaced as a maintainable path inventory, especially across helper modules.

### 4. Helper modules need explicit inclusion in the baseline contract

`kind` and `muid` cannot be treated as follow-up work. The phase context requires all three packages to clear the baseline bar before Phase 1 is complete.

## Implications For Planning

- Plan work must include explicit workspace-wide verification commands, not just root-module `./...`.
- At least one plan must audit and normalize helper-module coverage, starting with exported behavior and edge cases.
- At least one plan must inventory the current `hsm` behavioral coverage into clearer path-oriented groupings so later phases can build on it.
- The phase likely needs separate plan slices for: root runtime coverage structure, helper-module baseline closure, and release/evidence automation.
- Because the explicit full-workspace command already fails, Phase 1 must include remediation of currently exposed failures as part of establishing the baseline gate.

## Recommended Release-Gate Shape

### Quick local feedback

- `go test ./... ./kind/... ./muid/...`

This is the fastest meaningful workspace-wide baseline and should be used after individual task commits where possible.

### Full blocking gate

- `go test ./... ./kind/... ./muid/...`
- `go test -race -short ./... ./kind/... ./muid/...`

This matches the user decision that the gate should stay focused on core correctness in Phase 1 while still covering all workspace packages.

## Evidence Strategy

Phase 1 should produce a maintainable path matrix that:

- groups coverage by behavior family rather than raw percentage
- names known gaps explicitly
- marks whether each gap is covered now, failing now, or deferred to a later roadmap phase
- makes helper-module status visible alongside `hsm` instead of burying it in side tests

## Validation Architecture

Phase 1 validation should use native Go test commands with explicit multi-module coverage:

- Quick command: `go test ./... ./kind/... ./muid/...`
- Full command: `go test -race -short ./... ./kind/... ./muid/...`

Feedback is fast enough for iterative development, and both commands directly align with the Phase 1 release gate decisions. Manual verification should be unnecessary for this phase if the evidence artifact is generated from automated test inventory and command results.

## Risks And Watchouts

- The huge `hsm_test.go` file can absorb more assertions quickly, but continuing that pattern alone will make later coverage accounting harder.
- `muid_bench_test.go` includes functional tests; planning should account for whether those assertions stay there or move into clearer test files.
- Any CI updates must avoid reintroducing root-only green signals that hide failing nested modules.
- The legacy Travis file may confuse future contributors about the real gate and should be considered in release hygiene work if it interferes with clarity.

## Recommendation

Create Phase 1 plans that:

1. Establish explicit workspace-wide test commands and evidence generation.
2. Inventory and expand `hsm` behavior coverage into clearer path-oriented structure.
3. Bring `kind` and `muid` to the baseline bar, including fixing the currently exposed `muid` failure.
