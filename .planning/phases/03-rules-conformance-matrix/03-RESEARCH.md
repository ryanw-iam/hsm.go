# Phase 3: Rules Conformance Matrix - Research

**Researched:** 2026-03-13
**Domain:** Go executable conformance testing for the `hsm.go` modeling contract
**Confidence:** HIGH

<user_constraints>
## User Constraints

No phase-specific `03-CONTEXT.md` exists for this phase.

### Locked Decisions
- The source of truth for the modeling contract is [`rules.md`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/rules.md).
- Phase 3 must satisfy `VER-04`.
- Phase 3 goal: the modeling contract in `rules.md` is enforced by executable conformance coverage.
- Success criteria:
  - The documented rule set is represented by executable tests that demonstrate compliant usage patterns passing.
  - Invalid model definitions or runtime usages that violate documented rules fail in clear, repeatable ways.
  - Rule conformance evidence is maintained in-repo so regressions against the documented contract are caught before release.
- The output must include a `## Validation Architecture` section suitable for Nyquist validation scaffolding.

### Claude's Discretion
- Exact test file layout and test naming.
- Exact shape of the in-repo conformance evidence artifact.
- Exact grouping of rules into model-build, runtime-semantic, and positive-exemplar suites.

### Deferred Ideas (OUT OF SCOPE)
- Fault injection, fuzz/property-style exploration, and compatibility/performance regression work belong to Phase 4.
- Broad remediation of defects exposed by the higher verification bar belongs to Phase 5, except for minimal fixes required to make `VER-04` executable and green.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| VER-04 | Add conformance coverage for the library rules captured in `rules.md`. | Implement a checked-in rule matrix plus table-driven Go subtests split by enforcement class, then gate them through `go test` and the canonical workspace verifier. |
</phase_requirements>

## Summary

Phase 3 should be planned as a contract-matrix phase, not as another general feature-test expansion. The repository already enforces part of the documented contract directly in [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L595), [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L897), [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L1218), [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L1273), [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L1500), [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L1531), and [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go#L1801). The existing tests already cover attributes, calls, history, `AnyEvent`, completion-event priority, and waiter-based synchronization in scattered feature suites such as [`hsm_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test.go#L681), [`hsm_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test.go#L851), [`hsm_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test.go#L1117), [`hsm_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test.go#L1720), and [`hsm_runtime_concurrency_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_concurrency_test.go#L167). What is missing is a first-class mapping from each `HSMxx` rule to executable evidence.

The key planning decision is to classify every rule before writing tests. Some rules are directly enforceable as model-definition panics. Some are runtime semantics that need deterministic settled-state assertions. Some are usage conventions that the library cannot reliably reject today, so they need positive-only exemplar tests plus an explicit matrix label such as `exemplar` instead of fake negative coverage. Without that classification, the phase will either overpromise on unenforceable rules or bury evidence inside unrelated feature suites.

Current baseline status is good: `go test ./...` and `bash scripts/verify-workspace.sh` both passed in this workspace when rerun with a repo-local `GOCACHE` in the sandbox. That means Phase 3 can start from a green baseline and should focus on organizing and extending conformance evidence rather than first stabilizing the suite.

**Primary recommendation:** Build Phase 3 around a checked-in `03-RULES-MATRIX.md` plus two dedicated Go conformance suites, one for define/build-time rule failures and one for runtime/exemplar semantics, both using rule-ID subtests and existing waiter helpers.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | `go1.25.3` (repo pin in [`go.mod`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/go.mod)) | Compile and run the conformance suites | Already pinned by the repo and used by the existing verification gate |
| `testing` (stdlib) | toolchain stdlib | Table-driven subtests, failure reporting, and targeted `-run` execution | Official Go testing surface; no new dependency required |
| Existing root test helpers | current repo | Waiter and panic helpers in [`hsm_test_helpers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test_helpers_test.go#L263) | Matches current repository patterns and Phase 2 deterministic strategy |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `context` / `time` (stdlib) | toolchain stdlib | Runtime conformance scenarios and bounded waits | For runtime-settled dispatch, call, set, stop, and restart examples |
| `slices` / `sync` (stdlib) | toolchain stdlib | Trace comparison and shared test bookkeeping | For deterministic trace checks already used in root suites |
| `bash scripts/verify-workspace.sh` | current repo | Canonical full-suite gate | After wave merges and before phase completion |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `testing` + stdlib helpers | `stretchr/testify` or `gomega` | Adds dependency/style churn for little gain in this repo |
| Checked-in manual matrix | Parse `rules.md` and generate tests | Brittle and obscures review; the rules need enforcement classification anyway |
| Split model/runtime conformance files | One giant `hsm_rules_test.go` | Simpler at first, but harder to navigate across 54 rules |

**Installation:**
```bash
# No new dependencies are recommended for Phase 3.
go test ./...
```

## Architecture Patterns

### Recommended Project Structure
```text
./
├── hsm_rules_model_conformance_test.go      # define/build-time panic rules
├── hsm_rules_runtime_conformance_test.go    # runtime semantics and positive exemplars
├── hsm_test_helpers_test.go                 # extend with panic-message helpers only as needed
└── .planning/phases/03-rules-conformance-matrix/
    ├── 03-RESEARCH.md
    └── 03-RULES-MATRIX.md                   # rule -> enforcement class -> test mapping
```

### Pattern 1: Classify Rules Before Writing Tests
**What:** Add an explicit enforcement class per rule: `define_panic`, `runtime_semantic`, or `exemplar`.
**When to use:** Before creating plan slices or test files.
**Example:**
```markdown
| Rule | Enforcement | Evidence |
|------|-------------|----------|
| HSM06 | define_panic | TestModelRulesConformance/initial_transition_cannot_have_guard |
| HSM14 | runtime_semantic | TestRuntimeRulesConformance/specific_transitions_precede_any_event |
| HSM22 | exemplar | TestRuntimeRulesConformance/external_code_uses_attributes_not_internal_fields |
```

This avoids pretending that advisory API-usage rules can always be made to fail mechanically.

### Pattern 2: Use Rule-ID Subtests, Not Free-Form Test Names
**What:** Use one top-level test per suite and one `t.Run` subtest per rule or tightly-coupled rule pair.
**When to use:** For every conformance suite in this phase.
**Example:**
```go
func TestModelRulesConformance(t *testing.T) {
	cases := []struct {
		rule string
		name string
		build func()
		want string
	}{
		{
			rule: "HSM06",
			name: "initial_transition_cannot_have_guard",
			build: func() {
				hsm.Define(
					"BadInitialGuard",
					hsm.State("idle"),
					hsm.Initial(
						hsm.Target("idle"),
						hsm.Guard(func(context.Context, *THSM, hsm.Event) bool { return true }),
					),
				)
			},
			want: "cannot have a guard",
		},
	}

	for _, tc := range cases {
		t.Run(tc.rule+"/"+tc.name, func(t *testing.T) {
			assertPanicContains(t, tc.want, tc.build)
		})
	}
}
```

Source: official Go subtest guidance and repository-local panic-helper patterns. See `testing` package docs and the Go blog on subtests: https://pkg.go.dev/testing@go1.25.3 and https://go.dev/blog/subtests

### Pattern 3: Separate Build-Time Failure Tests From Runtime-Settled Semantics
**What:** Keep `hsm.Define(...)` panic cases apart from tests that must start a machine and wait for deterministic completion.
**When to use:** Whenever a rule could be asserted either at model build time or after dispatch.
**Example:**
```go
func TestRuntimeRulesConformance(t *testing.T) {
	t.Run("HSM53/wait_for_set_before_asserting_state", func(t *testing.T) {
		model := hsm.Define(
			"SetWaiterHSM",
			hsm.Attribute("count", 0),
			hsm.State("idle",
				hsm.Transition(hsm.OnSet("count"), hsm.Target("../changed")),
			),
			hsm.State("changed"),
			hsm.Initial(hsm.Target("idle")),
		)

		sm := hsm.Started(context.Background(), &AttrHSM{}, &model)
		awaitWaiter(t, "attribute change dispatch", hsm.Set(context.Background(), sm, "count", 1))

		if got := sm.State(); got != "/SetWaiterHSM/changed" {
			t.Fatalf("state = %s, want /SetWaiterHSM/changed", got)
		}
	})
}
```

Source: existing waiter helpers and runtime semantics already established in Phase 2. See [`hsm_test_helpers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test_helpers_test.go#L263) and [`hsm_runtime_concurrency_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_concurrency_test.go#L167).

### Anti-Patterns to Avoid
- **One giant “rules” integration test:** It hides which rule failed and makes `go test -run` targeting weak.
- **Exact full panic-string matching:** `traceback()` prefixes file and line numbers, so assert stable semantic fragments, not whole messages.
- **Sleep-based post-dispatch assertions:** Phase 2 already established waiter-based deterministic patterns; Phase 3 should reuse them.
- **Auto-generating test code from markdown:** The rules need human classification and curated examples more than code generation.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Per-rule result inventory | A parser that converts `rules.md` directly into code | A reviewed `03-RULES-MATRIX.md` maintained in-repo | The key missing data is enforcement class and evidence linkage, not markdown parsing |
| Assertion library | Custom DSL or third-party framework migration | Extend [`hsm_test_helpers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test_helpers_test.go#L281) with `assertPanicContains` | Keeps the suite aligned with current root tests |
| Synchronization for runtime rules | `time.Sleep` and polling loops | Returned waiter channels plus `awaitWaiter` | Deterministic and already proven in Phase 2 |
| Release gating | A new script just for conformance | Existing [`scripts/verify-workspace.sh`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/scripts/verify-workspace.sh) | One canonical gate is already a project decision |

**Key insight:** The difficult part of this phase is not mechanics. It is keeping every `HSMxx` rule reviewable, runnable, and honestly classified by what the library can actually enforce today.

## Common Pitfalls

### Pitfall 1: Treating Every Rule As A Negative Test
**What goes wrong:** Plans assume every rule violation should panic or fail at runtime.
**Why it happens:** `rules.md` mixes hard invariants with modeling guidance and API-usage discipline.
**How to avoid:** Add an enforcement-class column before writing tests and require each rule to be marked `define_panic`, `runtime_semantic`, or `exemplar`.
**Warning signs:** Test planning stalls on rules like HSM22-HSM28 because there is no honest failure mechanism to assert.

### Pitfall 2: Matching Full Panic Strings
**What goes wrong:** Tests become brittle because `traceback()` embeds file/line prefixes.
**Why it happens:** Existing helpers only assert panic/no-panic, so it is tempting to compare recovered panic values exactly.
**How to avoid:** Add helper(s) that check panic presence plus stable substring or regexp matching.
**Warning signs:** Conformance tests fail after unrelated line moves in [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go).

### Pitfall 3: Reintroducing Sleep-Based Runtime Assertions
**What goes wrong:** Runtime rule tests become flaky or encode timing luck.
**Why it happens:** Some older tests in [`hsm_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test.go#L851) and [`hsm_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test.go#L1720) still use deadline loops or `time.Sleep`.
**How to avoid:** Reuse `awaitWaiter`, returned channels from `Dispatch`/`Set`, and the Phase 2 runtime patterns.
**Warning signs:** A test must “sleep a little longer” to pass.

### Pitfall 4: Forgetting `AnyEvent` Internal-Event Filtering
**What goes wrong:** Catch-all transitions accidentally react to internal lifecycle events.
**Why it happens:** `processEvent` explicitly tries specific transitions first and then falls back to `AnyEvent`, but it does not filter internal events for user code.
**How to avoid:** Keep explicit guarded examples for HSM13, HSM14, and HSM52, using the same pattern already present in [`hsm_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test.go#L1720).
**Warning signs:** A generic `AnyEvent` test unexpectedly fires on `hsm_initial`, `hsm_started`, or completion events.

### Pitfall 5: Letting Phase 3 Expand Into Phase 4
**What goes wrong:** The plan starts introducing fuzzing/property work before the rule matrix is complete.
**Why it happens:** Once rule coverage gaps are visible, fuzzing looks attractive as a shortcut.
**How to avoid:** Keep Phase 3 deterministic and table-driven. Use official fuzzing support only in Phase 4.
**Warning signs:** Proposed commands include `go test -fuzz` before every rule has a deterministic matrix row.

## Code Examples

Verified patterns from official sources and current repo conventions:

### Rule-ID Subtests For Matrix Coverage
```go
func TestRuntimeRulesConformance(t *testing.T) {
	cases := []struct {
		rule string
		run  func(t *testing.T)
	}{
		{
			rule: "HSM14",
			run: func(t *testing.T) {
				// build model and assert specific event takes precedence over AnyEvent
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.rule, tc.run)
	}
}
```

Source: `testing` package docs and Go subtests guidance: https://pkg.go.dev/testing@go1.25.3 and https://go.dev/blog/subtests

### Panic-Contains Helper For Stable Negative Coverage
```go
func assertPanicContains(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q", want)
		}
		if !strings.Contains(fmt.Sprint(r), want) {
			t.Fatalf("panic %q does not contain %q", r, want)
		}
	}()
	fn()
}
```

Source: existing helper style in [`hsm_test_helpers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test_helpers_test.go#L281)

### Waiter-Based Runtime Assertions
```go
entered := hsm.AfterEntry(sm.Context(), sm, "/ExampleHSM/done")
awaitWaiter(t, "done state entry", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: "go"}))
awaitWaiter(t, "done entry", entered)

if sm.State() != "/ExampleHSM/done" {
	t.Fatalf("unexpected state: %s", sm.State())
}
```

Source: Phase 2 deterministic runtime suites, especially [`hsm_runtime_concurrency_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_concurrency_test.go#L167)

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Looping related cases inside one test without named subtests | `t.Run` subtests for per-rule IDs | Go 1.7 | Each `HSMxx` rule can be targeted with `go test -run` and reported independently |
| Sleep-based runtime checks | Waiter-channel and returned-channel synchronization | Repo pattern strengthened in Phase 2 on 2026-03-13 | Phase 3 can assert runtime rules deterministically |
| Ad hoc random regression exploration | Native Go fuzzing with seed corpus and `go test -fuzz` | Go 1.18+ | Useful later for Phase 4, but not a substitute for a deterministic Phase 3 matrix |

**Deprecated/outdated:**
- Sleep-heavy assertions after dispatch or call: replace with waiter-based settled-state assertions.
- Hiding rule evidence inside unrelated feature tests: replace with dedicated conformance suites plus a matrix artifact.

## Open Questions

1. **Should the phase update `rules.md` or only add matrix evidence?**
   - What we know: several rules are already enforced in code, while others are guidance only.
   - What's unclear: whether the current doc should explicitly mark enforceability classes.
   - Recommendation: keep `rules.md` as the contract text for now, and add enforceability/status columns in `03-RULES-MATRIX.md`. Only edit `rules.md` if the matrix exposes ambiguous wording that blocks implementation.

2. **Should every rule get its own dedicated subtest, or can tightly-coupled rules share one?**
   - What we know: subtests make per-rule targeting easy.
   - What's unclear: whether pairs like HSM13/HSM14 or HSM17/HSM18 should share one scenario.
   - Recommendation: allow one scenario to satisfy multiple rules only if the matrix names every satisfied rule explicitly. Do not let shared scenarios hide missing evidence.

3. **What happens if a new rule test exposes a real contract mismatch in production code?**
   - What we know: remediation is formally Phase 5, but `VER-04` requires executable conformance.
   - What's unclear: whether a minimal production fix should land during Phase 3 or be deferred.
   - Recommendation: plan for minimal, scope-bound fixes only when a failing rule test blocks the conformance matrix from honestly representing the documented contract. Everything else should be logged for Phase 5.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | `go test` (`testing` stdlib on Go `1.25.3`) |
| Config file | none |
| Quick run command | `go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance'` |
| Full suite command | `bash scripts/verify-workspace.sh` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VER-04 | Every documented rule in `rules.md` is mapped to executable evidence with deterministic pass/fail semantics where enforceable | unit/integration | `go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance'` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance'`
- **Per wave merge:** `go test ./...`
- **Phase gate:** `bash scripts/verify-workspace.sh` before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `hsm_rules_model_conformance_test.go` — table-driven build-time conformance cases for panic-enforced rules
- [ ] `hsm_rules_runtime_conformance_test.go` — deterministic runtime and exemplar rule coverage using waiter-based assertions
- [ ] `hsm_test_helpers_test.go` — add panic-message helpers such as `assertPanicContains`
- [ ] `.planning/phases/03-rules-conformance-matrix/03-RULES-MATRIX.md` — reviewable mapping from `HSMxx` rule IDs to enforcement class, test names, and status

## Sources

### Primary (HIGH confidence)
- [`rules.md`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/rules.md) - documented rule set for Phase 3
- [`hsm.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm.go) - model-build and runtime enforcement points
- [`hsm_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test.go) - current rule-adjacent behavior coverage
- [`hsm_test_helpers_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_test_helpers_test.go) - current panic and waiter helper patterns
- [`hsm_runtime_concurrency_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_concurrency_test.go) - Phase 2 deterministic waiter patterns and completion-event priority coverage
- [`scripts/verify-workspace.sh`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/scripts/verify-workspace.sh) - canonical verification gate
- https://pkg.go.dev/testing@go1.25.3 - official `testing` package documentation for Go test structure
- https://go.dev/blog/subtests - official Go guidance on `t.Run` subtests and targeted execution
- https://go.dev/doc/security/fuzz/ - official Go fuzzing guidance confirming native fuzzing belongs in a later phase, not as a replacement for deterministic rule coverage

### Secondary (MEDIUM confidence)
- None.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - the repo already uses stdlib-only Go tests, and official docs confirm subtest and test-structure guidance
- Architecture: HIGH - driven directly by the local codebase, Phase 2 patterns, and the need to map each `HSMxx` rule explicitly
- Pitfalls: MEDIUM - several are inferred from current test structure and rule semantics rather than from explicit upstream documentation

**Research date:** 2026-03-13
**Valid until:** 2026-04-12
