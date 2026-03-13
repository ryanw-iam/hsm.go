# Phase 4: Adversarial Regression Matrix - Research

**Researched:** 2026-03-13
**Domain:** Go adversarial verification, fuzzing, and CI-safe regression coverage for the `hsm.go` workspace
**Confidence:** HIGH

<user_constraints>
## User Constraints

No phase-specific `04-CONTEXT.md` exists for this phase.

### Locked Decisions
- Phase 4 must satisfy `VER-05`.
- Phase 4 goal: the verification suite covers hostile, malformed, and regression-prone conditions beyond nominal functional tests.
- Success criteria:
  - Fault-injected runtime and model failures can be triggered deliberately and produce asserted outcomes instead of ad hoc debugging.
  - Fuzz or property-style checks preserve core invariants for targeted input and state-path surfaces.
  - Compatibility or performance regressions that matter to downstream consumers are surfaced by automated checks before release.
  - Added adversarial coverage improves confidence without introducing flaky or unbounded standard CI behavior.
- Focus the research on concrete plan structure for:
  - fault-injected runtime and model failures
  - fuzz/property-style checks on core invariants
  - compatibility/performance regression checks that are CI-safe
  - how to avoid flaky or unbounded adversarial coverage in the canonical gate
  - what existing Phase 1-3 helpers and suites can be reused instead of inventing a parallel harness
- The output must include a `## Validation Architecture` section suitable for Nyquist validation scaffolding.

### Claude's Discretion
- Exact file layout and naming for adversarial tests, fuzz targets, and regression checks.
- Exact split between blocking canonical-gate coverage and opt-in deeper fuzz/benchmark commands.
- Exact selection of invariants to codify first for `hsm`, `kind`, and `muid`.

### Deferred Ideas (OUT OF SCOPE)
- Broad defect remediation beyond minimal changes required to make new adversarial checks executable belongs to Phase 5.
- Replacing the canonical verification entrypoint with a separate release harness is out of scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| VER-05 | Add fault-injection, fuzz/property-style, and compatibility/performance regression checks where needed to close verification gaps. | Build Phase 4 around existing deterministic helpers and runtime suites, then add bounded fault-injection tests, seed-backed fuzz/property targets, and CI-safe allocation/benchmark smoke through the canonical verifier. |
</phase_requirements>

## Summary

Phase 4 should be planned as a bounded adversarial layer on top of the Phase 1-3 verification stack, not as an open-ended stress lab. The repository already has the right building blocks: a single canonical gate in [`scripts/verify-workspace.sh`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/scripts/verify-workspace.sh), deterministic runtime helpers in [`hsm_test_helpers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test_helpers_test.go), timer and lifecycle fault seams in [`hsm_runtime_timers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_timers_test.go) and [`hsm_runtime_lifecycle_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_lifecycle_test.go), runtime rule scenarios in [`hsm_rules_runtime_conformance_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_rules_runtime_conformance_test.go), and established benchmark files in [`hsm_bench_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_bench_test.go) and [`muid/muid_bench_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/muid/muid_bench_test.go). The phase should extend those assets rather than create a second adversarial harness.

The runtime also exposes concrete hostile-path seams worth locking down. Concurrent behavior panics are recovered in [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L2945), processing panics are recovered and converted into `ErrorEvent` in [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L3057), termination timeout is already deterministic through injected `Clock.After` in [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L3050), and public helpers such as `Dispatch`, `Set`, `Call`, `DispatchAll`, and `DispatchTo` already define nil/missing-context behavior in [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L3234). Those are higher-value adversarial targets than inventing random model generation from scratch.

For CI safety, the phase should split adversarial work into two lanes. The canonical gate should run deterministic fault-injection tests, regular `go test` execution of fuzz seed corpora, and coarse regression checks such as allocation ceilings or tiny fixed-iteration benchmark smoke. Open-ended fuzzing with `go test -fuzz` and broader benchmark exploration should be codified as explicit bounded commands, but kept out of the default blocking path. This matches the official Go testing guidance: fuzzing is opt-in with `-fuzz` and bounded by `-fuzztime`, while ordinary `go test` runs the fuzz seed corpus as regular regression coverage.

**Primary recommendation:** Plan Phase 4 in three slices: add fault-injection suites around existing recovery seams, add bounded fuzz/property targets with checked-in seed corpora for core invariants, then wire CI-safe allocation and micro-benchmark regression evidence into `scripts/verify-workspace.sh` without introducing a second test harness.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | `go1.25.3` | Testing, fuzzing, benchmarks, and race/short execution | Pinned in [`go.mod`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/go.mod); official stdlib testing stack already covers the phase |
| `testing` (stdlib) | toolchain stdlib | Unit tests, fuzz targets, subtests, benchmarks, `Short`, `AllocsPerRun`, `RunParallel` | Official surface for deterministic adversarial tests and CI-safe regression checks |
| Existing root helpers | current repo | Waiters, panic assertions, deterministic clocks, trace capture | Already proven in Phases 2-3 and aligned with repo patterns |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `cmd/go test` flags | toolchain stdlib | `-fuzz`, `-fuzztime`, `-bench`, `-benchtime`, `-race`, `-short` | To separate bounded canonical checks from deeper opt-in adversarial sweeps |
| `context`, `time`, `sync`, `sync/atomic`, `strconv` | toolchain stdlib | Runtime fault scenarios, deterministic synchronization, invariant assertions | For runtime recovery, public-path regressions, and `muid` properties |
| Existing benchmark files | current repo | Small public-path benchmark smoke and longer advisory perf exploration | Reuse instead of creating a new performance harness |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Stdlib fuzzing/property checks | `rapid`, `gopter`, or other third-party property frameworks | Extra dependency and style drift with little benefit for this repo's targeted invariants |
| Allocation ceilings plus tiny benchmark smoke | Hard wall-clock thresholds in CI | Timing thresholds on shared runners are noisy and become flaky release blockers |
| Extending existing root suites | A separate adversarial runner or bespoke stress harness | Duplicates helpers, splits the release gate, and violates the existing project decision around `scripts/verify-workspace.sh` |

**Installation:**
```bash
# No new dependencies are recommended for Phase 4.
GOCACHE=$(pwd)/.cache/go-build go test ./... ./kind/... ./muid/...
```

## Architecture Patterns

### Recommended Project Structure
```text
./
├── hsm_runtime_adversarial_test.go        # panic recovery, malformed runtime input, fault injection
├── hsm_properties_fuzz_test.go            # bounded HSM helper/runtime-script fuzz targets
├── hsm_regression_allocs_test.go          # stable allocation ceilings for public hot paths
├── hsm_bench_test.go                      # narrowed/modernized benchmark smoke targets
├── hsm_test_helpers_test.go               # extend existing waiters/assertions only if needed
├── muid/
│   ├── muid_fuzz_test.go                  # generator/config/string invariants
│   ├── muid_regression_allocs_test.go     # zero/low-allocation checks
│   └── testdata/fuzz/                     # checked-in failing/interesting seed corpus
└── scripts/verify-workspace.sh            # canonical gate, extended with bounded adversarial checks
```

### Pattern 1: Fault-Inject Existing Recovery Seams
**What:** Add targeted tests that deliberately panic inside concurrent behaviors, synchronous effects, and other already-recovering runtime paths, then assert `ErrorEvent` delivery, waiter completion, and post-failure liveness.
**When to use:** For panic recovery, timeout handling, nil/missing-context behavior, and hostile runtime inputs.
**Example:**
```go
func TestRuntimeAdversarialProcessingPanicBecomesErrorEvent(t *testing.T) {
	errorsSeen := make(chan error, 1)
	workEvent := hsm.Event{Name: "work"}

	model := hsm.Define(
		"ProcessingPanicHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(
				hsm.On(workEvent),
				hsm.Effect(func(context.Context, *THSM, hsm.Event) {
					panic("boom")
				}),
			),
		),
		hsm.Transition(
			hsm.On(hsm.ErrorEvent),
			hsm.Effect(func(_ context.Context, _ *THSM, event hsm.Event) {
				if err, ok := event.Data.(error); ok {
					errorsSeen <- err
				}
			}),
		),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	errorProcessed := hsm.AfterProcess(sm.Context(), sm, hsm.ErrorEvent)
	awaitWaiter(t, "fault dispatch", hsm.Dispatch(sm.Context(), sm, workEvent))
	awaitWaiter(t, "error processing", errorProcessed)
	awaitWaiter(t, "stop after panic", hsm.Stop(context.Background(), sm))

	if err := <-errorsSeen; !strings.Contains(err.Error(), "panic while processing event") {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

Source: repo runtime recovery in [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L2945) and existing deterministic timeout/error pattern in [`hsm_runtime_lifecycle_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_lifecycle_test.go#L92).

### Pattern 2: Use Fuzz Targets As Regression Carriers, Not Canonical-Stress Jobs
**What:** Add small, deterministic fuzz targets with curated seeds. Under normal `go test`, those seeds run as regression tests; deeper exploration happens only under explicit `-fuzz` commands with bounded `-fuzztime`.
**When to use:** For pure or fresh-instance invariants such as `muid` config monotonicity, string round-trips, `Match` wildcard behavior, or tightly-bounded event-script interpreters.
**Example:**
```go
func FuzzGeneratorConfigMonotonicity(f *testing.F) {
	f.Add(uint64(7), 40, 14, int64(1700000000000))
	f.Add(uint64(0xffff), 40, 14, int64(1700000000000))

	f.Fuzz(func(t *testing.T, machineID uint64, timestampBits, machineBits int, epoch int64) {
		if timestampBits <= 0 || machineBits <= 0 || timestampBits+machineBits >= 63 || epoch <= 0 {
			t.Skip()
		}

		g := NewGenerator(Config{
			MachineID:       machineID,
			TimestampBitLen: timestampBits,
			MachineIDBitLen: machineBits,
			Epoch:           epoch,
		}, 0, 0)

		first := g.ID()
		second := g.ID()
		if second <= first {
			t.Fatalf("IDs must increase, got %d then %d", first, second)
		}
	})
}
```

Source: official Go fuzzing docs and current `muid` invariants in [`muid/muid_paths_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/muid/muid_paths_test.go).

### Pattern 3: Keep Canonical CI Bounded With Deterministic Seeds And Fixed Work
**What:** Keep `scripts/verify-workspace.sh` as the single release gate, but only add adversarial work that has bounded runtime: regular tests, fuzz seed-corpus execution, stable allocation checks, and tiny fixed-iteration benchmark smoke.
**When to use:** For everything that must run on every push, PR, and release workflow.
**Example:**
```bash
GOCACHE=$(pwd)/.cache/go-build go test ./... ./kind/... ./muid/...
GOCACHE=$(pwd)/.cache/go-build go test -race -short ./... ./kind/... ./muid/...
GOCACHE=$(pwd)/.cache/go-build go test -run '^$' -bench '^(BenchmarkNestedStates_NoEntryExitActivity|BenchmarkMUIDGeneration)$' -benchtime=20x -benchmem ./ ./muid
```

Source: canonical gate in [`scripts/verify-workspace.sh`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/scripts/verify-workspace.sh), CI workflows in [`.github/workflows/unit-tests.yml`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/.github/workflows/unit-tests.yml) and [`.github/workflows/release.yml`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/.github/workflows/release.yml), plus official `go test` flag docs.

### Pattern 4: Use Allocation Ceilings For Blocking Perf Regressions
**What:** Encode stable allocation expectations in regular tests using `testing.AllocsPerRun`; keep raw ns/op reporting as advisory unless a threshold is shown to be stable on the pinned toolchain.
**When to use:** For public hot paths that already have obvious low-allocation expectations, especially `muid.Make()` and tight root dispatch paths.
**Example:**
```go
func TestMUIDMakeRegressionAllocs(t *testing.T) {
	got := testing.AllocsPerRun(1000, func() {
		_ = muid.Make()
	})
	if got != 0 {
		t.Fatalf("Make() allocs/run = %.2f, want 0", got)
	}
}
```

Source: official `testing` package docs and current `muid` benchmark evidence in [`muid/muid_bench_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/muid/muid_bench_test.go).

### Anti-Patterns to Avoid
- **Open-ended fuzzing in the canonical gate:** `go test -fuzz` runs until stopped unless bounded; keep it out of default CI.
- **A second adversarial harness:** reuse `awaitWaiter`, `assertPanicContains`, deterministic clocks, and existing runtime suites instead of creating parallel helpers.
- **Wall-clock performance thresholds on shared CI runners:** prefer alloc ceilings and tiny fixed-iteration smoke.
- **Stateful fuzz targets that share global mutable state:** create a fresh model/generator per input so failures reproduce.
- **Randomized property loops without durable seeds:** when a hostile input matters, check it into `testdata/fuzz` or a fixed-seed table.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Property-based testing framework | Custom mini-framework or third-party dependency | `testing.F` fuzz targets plus checked-in seeds | Official, dependency-free, and supported by `go test` tooling |
| Panic/timeout harness | New ad hoc synchronization utilities | Existing `awaitWaiter`, `assertWaiterPending`, `assertPanicContains`, and deterministic clock harness | Already aligned with Phases 2-3 |
| Performance regression gate | Timing-threshold parser for benchmark output | `testing.AllocsPerRun` plus tiny `-bench` smoke commands | Allocation ceilings are much more stable in CI |
| Release verification entrypoint | A new `verify-adversarial.sh` used instead of the canonical script | Extend `scripts/verify-workspace.sh` | The project already decided there is one canonical gate |
| Random crash fishing | Infinite loops, sleeps, or unbounded goroutine storms | Deterministic fault models and bounded fuzz inputs | Reproducibility matters more than volume here |

**Key insight:** Adversarial coverage only pays off when a hostile input becomes a durable, bounded regression test. In this repo, that means converting failures into seed corpus files, fixed-seed tables, or stable allocation assertions that run through the same canonical gate as every other verification phase.

## Common Pitfalls

### Pitfall 1: Treating `go test -fuzz` As A Normal CI Command
**What goes wrong:** CI jobs become unbounded or time-variable.
**Why it happens:** Go fuzzing is designed for exploration, not default fixed-duration execution.
**How to avoid:** Keep canonical CI on ordinary `go test`; run deeper fuzzing only through explicit `-fuzz` commands with `-fuzztime`.
**Warning signs:** A proposed gate command includes `-fuzz` but no `-fuzztime`.

### Pitfall 2: Re-testing Timer And Waiter Semantics With New Helpers
**What goes wrong:** Phase 4 duplicates the deterministic work already done in Phases 2-3 and drifts from established semantics.
**Why it happens:** Adversarial work can feel “special” and tempt a separate harness.
**How to avoid:** Reuse `newDeterministicClockHarness`, `awaitWaiter`, and existing runtime rule scenarios as the base for hostile-path tests.
**Warning signs:** New helper names shadow functionality already present in [`hsm_test_helpers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test_helpers_test.go).

### Pitfall 3: Using Timing Thresholds As Blocking Perf Gates
**What goes wrong:** The suite flakes across runners or Go point releases.
**Why it happens:** Benchmark `ns/op` looks like the easiest regression metric.
**How to avoid:** Gate only coarse, stable contracts such as allocation ceilings. Keep benchmark output primarily for visibility unless a threshold is demonstrably stable.
**Warning signs:** Proposed assertions compare exact `ns/op` or percentage slowdown on GitHub Actions.

### Pitfall 4: Fuzzing Shared Global State
**What goes wrong:** Fuzz failures become irreproducible because the target mutates `DefaultClock`, default generators, or other global state across inputs.
**Why it happens:** The repo exposes globals such as `hsm.DefaultClock` and package-level `muid` generators.
**How to avoid:** Prefer pure/helper fuzz targets or construct fresh generators/models per input. Restore globals with `t.Cleanup` when a test must touch them.
**Warning signs:** A fuzz target modifies package globals or depends on call order across corpus entries.

### Pitfall 5: Letting Phase 4 Expand Into Remediation
**What goes wrong:** The plan starts fixing every defect found instead of characterizing and locking it down first.
**Why it happens:** Adversarial tests surface real bugs quickly.
**How to avoid:** Keep Phase 4 centered on exposure and repeatable regression coverage. Route broader behavior changes into Phase 5 unless minimal remediation is required to make the suite runnable.
**Warning signs:** A plan slice is mostly product/runtime refactoring rather than new adversarial evidence.

## Code Examples

Verified patterns from official sources and current repo conventions:

### Bounded Fuzz Command
```bash
GOCACHE=$(pwd)/.cache/go-build go test ./muid -run '^$' -fuzz '^FuzzGeneratorConfigMonotonicity$' -fuzztime=5s
```

Source: official `go test` flag docs and Go fuzzing docs.

### Stable Allocation Regression Check
```go
func TestDispatchRegressionAllocs(t *testing.T) {
	workEvent := hsm.Event{Name: "work"}

	model := hsm.Define(
		"DispatchAllocHSM",
		hsm.State("idle"),
		hsm.Transition(
			hsm.On(workEvent),
			hsm.Source("idle"),
			hsm.Target("idle"),
		),
		hsm.Initial(hsm.Target("idle")),
	)
	sm := hsm.Started(context.Background(), &THSM{}, &model)
	t.Cleanup(func() { awaitWaiter(t, "stop", hsm.Stop(context.Background(), sm)) })

	got := testing.AllocsPerRun(100, func() {
		awaitWaiter(t, "dispatch", hsm.Dispatch(sm.Context(), sm, workEvent))
	})
	if got > 0 {
		t.Fatalf("dispatch allocs/run = %.2f, want 0", got)
	}
}
```

Source: official `testing` package docs plus repo waiter conventions.

### Modern Benchmark Structure
```go
func BenchmarkMUIDGeneration(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = muid.Make()
	}
}
```

Source: official Go benchmark guidance (`b.Loop`) and the repo's existing benchmark surface.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `b.N` loops with manual timer control everywhere | Prefer `b.Loop` for benchmarks | Go 1.24 | Simplifies benchmarks and makes setup/measurement boundaries less error-prone |
| Running exploratory fuzzing directly in CI | Run fuzz seeds under normal `go test`, reserve `-fuzz` for bounded explicit jobs | Go fuzzing since 1.18, current docs still recommend explicit fuzz mode | Keeps canonical CI deterministic while preserving deeper exploration |
| Timing thresholds as perf gates | Allocation ceilings plus tiny fixed-iteration benchmark smoke | Current CI best practice | Reduces shared-runner flake risk |
| Sleep-based hostile-path tests | Waiter- and clock-based deterministic adversarial tests | Already established in Phase 2 | Lets Phase 4 remain adversarial without being nondeterministic |

**Deprecated/outdated:**
- Manual benchmark warmup/reporting patterns should be minimized when the same benchmark can be expressed with `b.Loop` and standard metrics.
- Unbounded fuzzing in push/PR workflows is the wrong default for this phase.

## Open Questions

1. **Which blocking perf contracts are stable enough to enforce?**
   - What we know: `BenchmarkMUIDGeneration` completed locally in about `39 ns/op` with `0 B/op`, and `BenchmarkNestedStates_NoEntryExitActivity` completed locally in about `1902 ns/op` with `312 B/op` using fixed `-benchtime` smoke runs.
   - What's unclear: which root-package allocation ceilings remain stable across the pinned Go toolchain and CI runners.
   - Recommendation: start with obvious low-allocation contracts (`muid.Make`, possibly selected dispatch paths), keep raw timing as advisory until a threshold proves stable.

2. **Should Phase 4 fuzz runtime event scripts or stay on helper/module invariants first?**
   - What we know: the runtime has deterministic helpers, but fresh-machine setup is materially heavier than `muid` or pure helper targets.
   - What's unclear: whether a runtime-script fuzz target adds enough new signal beyond bounded seeded subtests to justify the extra complexity.
   - Recommendation: plan helper/module fuzz targets first, then add one runtime-script fuzz target only if it stays fast, deterministic, and easy to reproduce.

3. **What is the intended contract for `Call` when the operation itself panics?**
   - What we know: concurrent behavior panics and process panics are recovered into `ErrorEvent`, but direct operation invocation appears to propagate panic rather than recover.
   - What's unclear: whether that is intentional API behavior or a latent defect.
   - Recommendation: add a characterization test in Phase 4; if behavior needs to change, route the fix and contract update into Phase 5.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` / `go test` (`go1.25.3`) |
| Config file | [`go.mod`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/go.mod) |
| Quick run command | `GOCACHE=$(pwd)/.cache/go-build go test ./... ./muid/... -run 'Test.*Adversarial|Test.*Regression|Fuzz'` |
| Full suite command | `GOCACHE=$(pwd)/.cache/go-build bash scripts/verify-workspace.sh` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VER-05 | Fault-injected runtime/model failures are deliberate and asserted; bounded fuzz/property invariants run in normal `go test`; CI-safe perf/compat regressions are surfaced before release. | unit + fuzz-seed + benchmark smoke | `GOCACHE=$(pwd)/.cache/go-build go test ./... ./kind/... ./muid/... && GOCACHE=$(pwd)/.cache/go-build go test -run '^$' -bench '^(BenchmarkNestedStates_NoEntryExitActivity|BenchmarkMUIDGeneration)$' -benchtime=20x -benchmem ./ ./muid` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `GOCACHE=$(pwd)/.cache/go-build go test ./... ./muid/... -run 'Test.*Adversarial|Test.*Regression|Fuzz'`
- **Per wave merge:** `GOCACHE=$(pwd)/.cache/go-build bash scripts/verify-workspace.sh`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `hsm_runtime_adversarial_test.go` — recovered panic, malformed input, and post-failure liveness coverage for `VER-05`
- [ ] `hsm_properties_fuzz_test.go` — bounded helper/runtime invariant fuzz seeds for `VER-05`
- [ ] `hsm_regression_allocs_test.go` — stable allocation ceilings for blocking perf regression signal
- [ ] `muid/muid_fuzz_test.go` — generator config, monotonicity, and round-trip fuzz targets for `VER-05`
- [ ] `muid/muid_regression_allocs_test.go` — zero/low-allocation public-path checks
- [ ] `muid/testdata/fuzz/` and any root `testdata/fuzz/` seeds — durable hostile-input corpus for normal `go test`
- [ ] `scripts/verify-workspace.sh` update — bounded benchmark smoke section for Phase 4

## Sources

### Primary (HIGH confidence)
- Official Go fuzzing docs: https://go.dev/doc/security/fuzz/ - checked fuzz target requirements, seed corpus behavior, and bounded fuzz workflow
- Official `testing` package docs (`go1.25.3`): https://pkg.go.dev/testing@go1.25.3 - checked `Short`, `AllocsPerRun`, fuzz/test APIs, benchmark/sub-benchmark APIs, and `RunParallel`
- Official `go test` flag docs (`go1.25.3`): https://pkg.go.dev/cmd/go/internal/test@go1.25.3 - checked `-fuzz`, `-fuzztime`, `-bench`, `-benchtime`, `-race`, and `-short`
- Official Go benchmark guidance: https://go.dev/blog/testing-b-loop - checked current `b.Loop` recommendation
- Repo runtime recovery and public helpers: [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go) - checked panic recovery, timeout behavior, and public helper semantics
- Repo deterministic helpers and runtime suites: [`hsm_test_helpers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test_helpers_test.go), [`hsm_runtime_timers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_timers_test.go), [`hsm_runtime_lifecycle_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_lifecycle_test.go), [`hsm_runtime_concurrency_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_concurrency_test.go), [`hsm_rules_runtime_conformance_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_rules_runtime_conformance_test.go)
- Repo benchmark and invariant files: [`hsm_bench_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_bench_test.go), [`muid/muid_bench_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/muid/muid_bench_test.go), [`muid/muid_paths_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/muid/muid_paths_test.go), [`muid/muid_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/muid/muid_test.go)
- Canonical CI wiring: [`.github/workflows/unit-tests.yml`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/.github/workflows/unit-tests.yml), [`.github/workflows/release.yml`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/.github/workflows/release.yml)

### Secondary (MEDIUM confidence)
- None.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - repo-local verification surface and official Go docs are aligned; no new dependency choice is needed
- Architecture: HIGH - reuse path is explicit in current helpers, suites, workflows, and runtime seams
- Pitfalls: HIGH - official docs and repo history strongly support the bounded-fuzz, deterministic-harness, and alloc-over-timing guidance

**Research date:** 2026-03-13
**Valid until:** 2026-04-12
