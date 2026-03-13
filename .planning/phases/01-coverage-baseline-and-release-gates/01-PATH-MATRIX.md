# Phase 1 Path Matrix

Phase 1 evidence tracks concrete behavior families instead of a headline coverage percentage. `covered` rows block release today through `scripts/verify-workspace.sh`. `deferred` rows stay visible and name the owning roadmap phase. `failing` rows are reserved for known open blockers; there are none at Phase 1 close.

| Package | Behavior family | Status | Evidence | Notes / owner |
| --- | --- | --- | --- | --- |
| `hsm` | Model path helpers (`Match`, `LCA`, `IsAncestor`) | covered | `hsm_model_paths_test.go` | Explicit helper assertions added in Wave 1 |
| `hsm` | History helper construction paths | covered | `hsm_model_paths_test.go` | Confirms Phase 1 model helper construction does not panic |
| `hsm` | Runtime identity, context, snapshot, restart helpers | covered | `hsm_runtime_paths_test.go` | Focused runtime helper coverage |
| `hsm` | Observer waiters (`AfterDispatch`, `AfterProcess`, `AfterEntry`, `AfterExit`) | covered | `hsm_observer_paths_test.go` | Release gate now exercises observer-family paths explicitly |
| `hsm` | Core integration/runtime transitions, attributes, operations, dispatch APIs | covered | `hsm_test.go` | Existing large integration suite remains part of the blocking gate |
| `hsm` | Timer/concurrency ordering without sleep heuristics | deferred | `hsm_test.go`, Phase 2 | Phase 2 owns deterministic runtime semantics and deeper timing control |
| `hsm` | Rules contract for valid and invalid models from `rules.md` | deferred | `rules.md`, Phase 3 | Phase 3 turns the documented contract into executable conformance tests |
| `hsm` | Fault injection, fuzz/property checks, and regression/perf guards | deferred | Phase 4 | Adversarial hardening belongs to Phase 4 |
| `kind` | ID allocation, ancestry flattening, transitive matching | covered | `kind/kind_test.go` | Explicit helper-module baseline coverage added in Wave 1 |
| `kind` | Broader fuzz/property invariants across many inheritance graphs | deferred | Phase 4 | Adversarial matrix owns scale/fuzz invariants |
| `muid` | Generator defaults, machine-ID masking, clock regression, counter overflow | covered | `muid/muid_paths_test.go` | Directly exercises internal generator path edges |
| `muid` | Public `Make`/`MakeString` monotonicity, uniqueness, and string round-trip | covered | `muid/muid_test.go`, `muid/muid_paths_test.go` | Phase 1 closes the exposed monotonicity/sortability defect |
| `muid` | Benchmark comparisons and allocation tracking | covered | `muid/muid_bench_test.go` | Benchmarks remain non-blocking evidence, behavior assertions moved out |
| `muid` | Fault injection, wraparound stress, and compatibility/performance regressions | deferred | Phase 4 | Reserved for adversarial matrix work |

## Failing

None currently open at Phase 1 close. The earlier `muid` monotonicity failure is resolved by `fix(01-02): restore monotonic default muid generation` (`d2baf42`).

## Deferred Ownership

- **Phase 2**: deterministic runtime semantics, timer ordering, concurrency/race-sensitive behavior
- **Phase 3**: executable `rules.md` conformance matrix for valid and invalid model definitions
- **Phase 4**: adversarial regression matrix, fuzz/property coverage, and compatibility/performance regression guards
