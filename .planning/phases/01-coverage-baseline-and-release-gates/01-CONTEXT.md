# Phase 1: Coverage Baseline And Release Gates - Context

**Gathered:** 2026-03-13
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 1 establishes exhaustive baseline path coverage and a release-gated verification entry point across the `hsm`, `kind`, and `muid` packages. This phase defines the blocking baseline for shipping. It does not yet claim full deterministic concurrency proof, full `rules.md` conformance, or the full adversarial matrix reserved for later phases.

</domain>

<decisions>
## Implementation Decisions

### Release gate shape
- Phase 1 must define one canonical verification command that both maintainers and CI can run.
- The initial blocking gate should focus on core correctness suites across `hsm`, `kind`, and `muid`, rather than forcing every heavier verification mode into Phase 1.
- Failing assertions, broken package suites, and uncovered known panic-path behavior are release blockers in this phase.
- Flaky tests are not allowed in the blocking gate. They must be fixed or excluded until deterministic.

### Evidence format
- Phase 1 evidence should be organized as a path matrix, not a headline coverage number.
- The matrix must name every known gap discovered during Phase 1, rather than hiding missing areas behind aggregate pass/fail output.
- The primary audience is maintainers and release reviewers deciding whether the workspace is safe to ship.
- Deferred gaps assigned to later phases should remain visible in the evidence, but should not fail the Phase 1 gate solely because they are explicitly deferred.

### Workspace-wide baseline
- Phase 1 is not complete unless `hsm`, `kind`, and `muid` all clear the baseline verification bar.
- Helper modules should be covered first through exported behavior, invalid inputs, and obvious edge or boundary cases.
- A materially lagging package blocks Phase 1 completion even if the main `hsm` runtime is stronger.

### Test suite shape
- Phase 1 should introduce clearer path-inventory structure immediately instead of only extending the existing monolithic integration-style suites.
- Existing large tests may remain, but new coverage work should be organized around explicit behavior and path groups so later phases can build on it cleanly.

### Claude's Discretion
- Exact naming and file layout for path inventory suites.
- Exact shape of the canonical verification command and supporting scripts.
- Whether evidence is emitted as markdown, generated text, structured reports, or a combination, as long as the path matrix is clear to maintainers.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `hsm_test.go`: already contains broad integration-style runtime scenarios, trace capture helpers, panic assertions, and behavior sequencing checks that can seed the Phase 1 inventory.
- `hsm_bench_test.go`: provides representative runtime scenarios that may help identify hot or branch-heavy behavior families, even though performance gating is a later phase concern.
- `.github/workflows/unit-tests.yml`: existing CI entry point already runs `go test -v ./...` and `go test -race -short ./...`, so Phase 1 can evolve a release gate from an existing automation path rather than inventing CI from scratch.

### Established Patterns
- The main runtime is concentrated in `hsm.go`, so Phase 1 should expect integration-style and behavior-family coverage to be more practical than highly isolated internal unit slices.
- `kind/kind_test.go` is effectively empty, while `muid/muid_test.go` is narrow and brute-force oriented; package verification depth is currently uneven and must be normalized.
- Panic-based validation and returned-error behavior both exist in the runtime API surface, so the baseline must treat panic behavior as part of the contract rather than as incidental failures.

### Integration Points
- The release gate should integrate with the existing GitHub Actions workflow set under `.github/workflows/`.
- Evidence should map to the public package surfaces and critical runtime branches in `hsm.go`, plus the exported helper-module behavior in `kind/` and `muid/`.
- Later phases will extend this baseline into deterministic runtime semantics, `rules.md` conformance, and adversarial verification, so Phase 1 artifacts should make deferred ownership explicit.

</code_context>

<specifics>
## Specific Ideas

- The desired outcome is enterprise-grade, release-gated verification for a library already used inside an event-driven agentic IGA platform.
- The user wants every meaningful code path tested and explicitly accepts fixing production code when the new tests expose defects.

</specifics>

<deferred>
## Deferred Ideas

- Deterministic concurrency, timing, dispatch ordering, and race-sensitive proof beyond the baseline gate belong to Phase 2.
- Full executable `rules.md` conformance belongs to Phase 3.
- Fault injection, fuzz/property checks, and compatibility or performance regression gates belong to Phase 4.

</deferred>

---

*Phase: 01-coverage-baseline-and-release-gates*
*Context gathered: 2026-03-13*
