# Phase 05: Hardening Remediation Closure - Research

**Researched:** 2026-03-13
**Domain:** Go verification-hardening closeout for `hsm.go`
**Confidence:** HIGH

## Summary

Phase 5 should be planned as a narrow closure phase, not as another verification-expansion phase. The major correctness defects actually exposed by the new bar were already fixed during earlier execution: Phase 2 closed stop-path lifecycle waiter and queued-timeout processing bugs, Phase 3 tightened define-time ownership and top-level transition validation while removing a race-sensitive HSM35 assertion bug, and Phase 4 hardened typed-nil helper contexts plus locked in adversarial, fuzz-seed, allocation, and benchmark-smoke coverage. The current workspace gate is green when run with a repo-local `GOCACHE`, and the canonical verifier already exercises the hostile-path and regression suites end to end.

What remains is adjacent remediation and release-clean closure around surfaced contract seams and maintenance drift. The strongest product-level gap still visible is helper consistency: `Dispatch`, `Set`, `Call`, `DispatchAll`, and `DispatchTo` now degrade cleanly for nil or typed-nil hostile inputs, but `Stop(nil)` is still only characterized as a panic. There is also contract drift between `rules.md` and public docs/comments: HSM54 says the waiter APIs are not production synchronization mechanisms, while exported comments and `README.md` still describe them as general synchronization tools. Outside runtime behavior, the canonical verifier is still Phase 4-branded, legacy `.travis.yml` remains in-tree, and `rules.md` is currently untracked even though Phase 3 turned it into executable contract evidence.

**Primary recommendation:** Plan Phase 5 as 2-3 tightly scoped plans: close surfaced helper-contract inconsistencies, align public docs/comments and tracked contract artifacts with the hardened rules, then finish with a release-clean canonical gate and repo-hygiene audit.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| VER-06 | Fix any correctness, determinism, observability, or maintainability issues uncovered by the new verification bar. | Earlier phases already fixed the main runtime defects, so Phase 5 should focus on remaining surfaced seams: nil-helper consistency, waiter contract/documentation alignment, canonical gate naming/hygiene cleanup, tracked contract artifacts, and final end-to-end release evidence. |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | 1.25.3 | Build, test, race, benchmark, and fuzz-seed replay across the workspace | Pinned by `go.mod`, `go.work`, and `.tool-versions`; all verification already runs through standard `go test` flows |
| `testing` (stdlib) | Go 1.25.3 | Regression tests, benchmark smoke, and allocation ceilings | Existing suites already use standard `go test`, `testing.AllocsPerRun`, and benchmark entrypoints |
| `scripts/verify-workspace.sh` | repo-local | Canonical release-blocking verifier | Release and CI workflows already call this script directly; do not fork verification commands again |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go race detector | Go 1.25.3 | Prove concurrency-sensitive runtime paths remain clean | Required for any remediation touching dispatch, stop, lifecycle, or helper concurrency surfaces |
| Go fuzz seed replay | Go 1.25.3 | Replay checked-in property corpus deterministically under normal `go test` | Use to preserve Phase 4 adversarial/property evidence without open-ended fuzzing |
| `testing.AllocsPerRun` | Go 1.25.3 | Lock hot-path allocation ceilings | Use when remediation risks changing helper-path allocation behavior |
| GitHub Actions `setup-go` | v5 | CI toolchain alignment via `go-version-file` | Keep active workflows aligned to `go.mod`; do not add hard-coded version drift back |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `scripts/verify-workspace.sh` | New Phase 5-specific gate script | Wrong direction; duplicates the existing release contract and weakens traceability |
| Allocation ceilings + fixed bench smoke | Time-based performance thresholds | Flakier and less portable; current alloc contracts are already stable and release-relevant |
| Explicit HSM10/HSM11 completion semantics | Implement implicit completion transitions | That is feature work, not closure; rules and conformance tests already document implicit completion as unsupported |

**Installation:**
```bash
go mod download
```

## Architecture Patterns

### Recommended Project Structure
```text
.
├── hsm.go                              # Runtime, DSL, helpers, and exported comments
├── README.md                           # Public package contract
├── rules.md                            # Rule contract that Phase 3 turned into executable evidence
├── scripts/verify-workspace.sh         # Canonical release gate
├── hsm_runtime_adversarial_test.go     # Hostile helper/runtime regressions
├── hsm_regression_allocs_test.go       # Root allocation ceilings
├── hsm_rules_*_conformance_test.go     # Rule contract suites
└── .planning/phases/05-.../            # Phase 5 research, plans, and verification
```

### Pattern 1: Remediation-Only Closure Loop
**What:** Start from a surfaced gap that already has evidence, add or tighten the smallest regression proving the defect, apply the narrowest production or docs fix, then rerun the canonical gate.
**When to use:** For nil-helper contract cleanup, docs/rules alignment, verifier/hygiene cleanup, or any regression directly implied by Phases 1-4.
**Example:**
```go
// Source: hsm_runtime_adversarial_test.go
t.Run("nil_stop_contract", func(t *testing.T) {
    done := hsm.Stop(context.Background(), nil)
    assertWaiterClosed(t, "stop with nil instance", done)
})
```

### Pattern 2: Contract Alignment Across Runtime, Rules, And Docs
**What:** Whenever remediation changes or clarifies a public behavior, update `hsm.go` comments, `README.md`, and `rules.md` together, then keep the regression in the focused suite that exposed the issue.
**When to use:** For waiter API guidance, nil-helper semantics, or any change affecting user-facing behavior or diagnosis.
**Example:**
```go
// Source: hsm.go + rules.md
// Waiter helpers are for tests and deterministic observation only.
// Production synchronization must use returned completion channels from Dispatch/Set/Stop/...
func AfterProcess(ctx context.Context, hsm Instance, maybeEvent ...Event) <-chan struct{} { ... }
```

### Pattern 3: Final Gate Uses Existing Evidence, Not New Verification Families
**What:** The last plan should only wire already-established suites and close named hygiene gaps; it should not add new fuzz targets, new benchmark families, or broad runtime refactors.
**When to use:** For phase closeout and final verification evidence.
**Example:**
```bash
# Source: scripts/verify-workspace.sh
bash scripts/verify-workspace.sh
```

### Anti-Patterns to Avoid
- **Turning Phase 5 into Phase 4.5:** Do not add new verification categories unless a remediation change requires direct regression coverage.
- **Treating HSM10 as a bug:** Implicit completion transitions are explicitly documented as unsupported in `rules.md` and enforced by conformance tests.
- **Fixing environment-only cache issues in product code:** The `GOCACHE` workaround observed here is execution-environment noise, not repository behavior.
- **Doing a large internal refactor for cleanliness alone:** `hsm.go` is monolithic, but Phase 5 is not the place for speculative file-splitting without surfaced defects demanding it.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Final release verification | A new phase-specific script or copied workflow commands | `bash scripts/verify-workspace.sh` | Active CI and release workflows already depend on it |
| Performance regression checking | Wall-clock pass/fail thresholds | `testing.AllocsPerRun` plus fixed-iteration benchmark smoke | More stable across machines and already proven in Phase 4 |
| New adversarial harnesses | A separate fault runner or bespoke scheduler harness | Existing waiter helpers, deterministic clock harness, and adversarial suites | Reuse keeps failures diagnosable and consistent |
| Completion semantics | Implicit triggerless transitions | Explicit `CompletionEventKind` events per HSM11 | Matches the documented and tested contract |
| Production synchronization guidance | Custom external coordination based on waiter helpers | Returned completion channels from `Dispatch`, `Set`, `Restart`, `Stop`, `DispatchAll`, `DispatchTo` | HSM53/HSM54 already define the supported contract |

**Key insight:** Phase 5 should close consistency and hygiene gaps around the hardened contract, not invent new mechanisms that bypass the verified patterns the earlier phases already established.

## Common Pitfalls

### Pitfall 1: Reopening Already-Closed Runtime Defects
**What goes wrong:** Planning treats prior defects as still open and bloats Phase 5 with work already finished in Phases 2-4.
**Why it happens:** The roadmap says remediation is final, but earlier execution summaries already applied the necessary fixes during verification expansion.
**How to avoid:** Use previous phase summaries as the defect ledger; only pull forward gaps that still remain visible in the current codebase or gate.
**Warning signs:** A plan item restates Phase 2 stop-timeout handoff, Phase 4 typed-nil context guards, or Phase 3 HSM20/HSM35 fixes without identifying a remaining adjacent seam.

### Pitfall 2: Expanding Verification Instead Of Closing Remediation
**What goes wrong:** Phase 5 adds new property suites, broader benchmarks, or new matrices instead of closing surfaced defects and repo cleanliness.
**Why it happens:** The verification bar is now strong, so it is tempting to keep adding evidence instead of finishing.
**How to avoid:** Require every Phase 5 task to point to a previously surfaced issue, inconsistency, or release-clean gap.
**Warning signs:** New test files with no linked prior finding, or plans that say "add more coverage" without a known defect family.

### Pitfall 3: Leaving Public Contract Drift After Hardening
**What goes wrong:** Rules, comments, README, and tests disagree about what callers are supposed to rely on.
**Why it happens:** Earlier phases optimized for getting regressions green; docs and comments can lag behind.
**How to avoid:** Any public-behavior remediation must update `rules.md`, exported comments, and README in the same change.
**Warning signs:** HSM54 says waiter APIs are not for production synchronization, but `hsm.go` comments and `README.md` still market them as general synchronization tools.

### Pitfall 4: Shipping With Cosmetic But Misleading Release Drift
**What goes wrong:** The repo is green, but the final gate still looks phase-specific or contains stale legacy automation artifacts.
**Why it happens:** The verifier grew during earlier phases and never got renamed for final closeout.
**How to avoid:** Make the final plan include a release-clean audit of script messaging, tracked contract artifacts, and legacy workflow leftovers.
**Warning signs:** `scripts/verify-workspace.sh` still prints "phase 4 workspace verification", `.travis.yml` still advertises obsolete Go versions, or `rules.md` remains untracked.

### Pitfall 5: Mistaking Sandbox Noise For Product Failure
**What goes wrong:** Planning spends remediation budget on local cache workarounds rather than repository changes.
**Why it happens:** This environment blocks the default Go build cache path, which makes `go test` fail unless `GOCACHE` is redirected.
**How to avoid:** Treat local `GOCACHE="$PWD/.cache/go-build"` as an execution note for this environment only; product closure is the green canonical gate itself.
**Warning signs:** Proposed code changes mention cache directories or environment-specific filesystem permissions.

## Code Examples

Verified patterns from current sources:

### Helper-Adversarial Regression Pattern
```go
// Source: hsm_runtime_adversarial_test.go
t.Run("typed_nil_context_helpers_do_not_panic", func(t *testing.T) {
    assertNoPanic(t, "dispatch with typed nil context", func() {
        assertWaiterClosed(t, "dispatch with typed nil context", hsm.Dispatch(nilCtx, nil, hostileEvent))
    })
})
```

### Allocation Regression Pattern
```go
// Source: hsm_regression_allocs_test.go + testing.AllocsPerRun docs
allocs := testing.AllocsPerRun(runs, func() {
    <-hsm.Dispatch(context.Background(), sm, event)
})
if allocs > 5 {
    t.Fatalf("Dispatch allocs = %v, want <= 5", allocs)
}
```

### Canonical Gate Pattern
```bash
# Source: scripts/verify-workspace.sh
go test ./...
go test ./kind/...
go test ./muid/...
go test -race -short ./... ./kind/... ./muid/...
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Sleep-driven runtime assertions | Waiter-based deterministic suites with explicit clock control | Phase 2 on 2026-03-13 | Remediation should preserve deterministic synchronization and avoid reintroducing incidental timing |
| Scattered rule guidance | Checked-in rule matrix plus dedicated `TestModelRulesConformance` and `TestRuntimeRulesConformance` suites | Phase 3 on 2026-03-13 | Public contract changes now require synchronized docs and rule/test updates |
| Exploratory adversarial checks | Deterministic hostile-path tests, checked-in fuzz seeds, alloc ceilings, and benchmark smoke in the canonical gate | Phase 4 on 2026-03-13 | Phase 5 can use existing evidence as the release bar instead of adding new verification families |
| Phase-specific verifier messaging | Final release gate should be phase-neutral | Pending Phase 5 | Failure output should describe the workspace gate, not an obsolete project phase |

**Deprecated/outdated:**
- `.travis.yml`: Legacy CI metadata pinned to Go 1.16/1.17; it no longer matches active GitHub Actions or the pinned toolchain.
- Phase 4 verifier wording in `scripts/verify-workspace.sh`: Accurate during Phase 4 closeout, outdated for final release closure.

## Open Questions

1. **What should the nil-instance contract be for `Stop`?**
   - What we know: `Dispatch`, `Set`, `Call`, `DispatchAll`, and `DispatchTo` already degrade safely for hostile inputs, but `Stop(nil)` is still explicitly characterized as a panic.
   - What's unclear: Whether `Stop(nil)` should return a closed waiter, mirror an `ErrNilHSM` policy through a separate API shape, or remain a panic by design.
   - Recommendation: Treat this as the highest-priority Phase 5 product decision and align `Stop` with the broader helper contract rather than leaving it as a one-off panic seam.

2. **Should `rules.md` become a tracked release artifact?**
   - What we know: Phase 3 turns `rules.md` into executable contract evidence, but `git ls-files --error-unmatch rules.md` currently fails and `git status --short` shows `?? rules.md`.
   - What's unclear: Whether the file is intentionally local-only or accidentally unstaged drift.
   - Recommendation: Phase 5 should end with `rules.md` either tracked in git or deliberately removed from the repo’s hard contract references; leaving it half-authoritative is not release-clean.

3. **Is docs cleanup enough for waiter helpers, or should behavior change too?**
   - What we know: HSM54 forbids using waiter helpers as production synchronization, while `hsm.go` comments and `README.md` still describe them as general synchronization utilities.
   - What's unclear: Whether only wording should change or whether additional API naming/documentation boundaries are needed.
   - Recommendation: At minimum, fix comments and README in Phase 5. Only broaden API changes if a concrete misuse or ambiguity remains after wording is corrected.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` via Go 1.25.3 |
| Config file | none |
| Quick run command | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|Test.*RegressionAllocs'` |
| Full suite command | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VER-06 | Surfaced helper/runtime remediation remains closed, especially hostile-input and regression-prone public paths | integration | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|Test.*RegressionAllocs'` | ✅ |
| VER-06 | Final release gate stays green after remediation and still exercises baseline, adversarial, fuzz-seed, alloc, benchmark-smoke, and race-short coverage | smoke | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` | ✅ |
| VER-06 | Repo exits Phase 5 in a release-clean state: tracked contract artifact, no stale phase branding, no legacy CI ambiguity | manual-only | `git ls-files --error-unmatch rules.md && ! rg -n 'phase 4 workspace verification' scripts/verify-workspace.sh && test ! -f .travis.yml` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|Test.*RegressionAllocs'`
- **Per wave merge:** `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Phase gate:** Full suite green plus repo-hygiene audit before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] Add a small repo-hygiene check script or make-target that asserts `rules.md` is tracked, the canonical verifier is phase-neutral, and legacy `.travis.yml` is absent or intentionally archived.
- [ ] If `Stop(nil)` is normalized, extend `hsm_runtime_adversarial_test.go` with the new non-panic contract before changing production code.

## Sources

### Primary (HIGH confidence)
- Repository evidence:
  - `.planning/REQUIREMENTS.md`
  - `.planning/STATE.md`
  - `.planning/ROADMAP.md`
  - `.planning/phases/01-coverage-baseline-and-release-gates/01-VERIFICATION.md`
  - `.planning/phases/02-deterministic-runtime-semantics/02-01-SUMMARY.md`
  - `.planning/phases/02-deterministic-runtime-semantics/02-02-SUMMARY.md`
  - `.planning/phases/02-deterministic-runtime-semantics/02-03-SUMMARY.md`
  - `.planning/phases/03-rules-conformance-matrix/03-01-SUMMARY.md`
  - `.planning/phases/03-rules-conformance-matrix/03-02-SUMMARY.md`
  - `.planning/phases/03-rules-conformance-matrix/03-03-SUMMARY.md`
  - `.planning/phases/03-rules-conformance-matrix/03-04-SUMMARY.md`
  - `.planning/phases/04-adversarial-regression-matrix/04-01-SUMMARY.md`
  - `.planning/phases/04-adversarial-regression-matrix/04-02-SUMMARY.md`
  - `.planning/phases/04-adversarial-regression-matrix/04-03-SUMMARY.md`
  - `hsm.go`
  - `README.md`
  - `rules.md`
  - `scripts/verify-workspace.sh`
  - `.github/workflows/unit-tests.yml`
  - `.github/workflows/release.yml`
  - `.github/workflows/auto-release.yml`
  - `.travis.yml`
- Official Go docs:
  - https://pkg.go.dev/testing#AllocsPerRun - `testing.AllocsPerRun` contract
  - https://go.dev/doc/articles/race_detector - `go test -race` guidance
  - https://go.dev/doc/tutorial/fuzz - checked-in fuzz seed corpus and `testdata/fuzz`
  - https://pkg.go.dev/cmd/go#hdr-Test_packages - `go test` flags for `-run`, `-bench`, `-short`, and `-fuzz`

### Secondary (MEDIUM confidence)
- None.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Derived from pinned repo tooling plus official Go docs for the test mechanisms already in use
- Architecture: HIGH - Based on current repo structure, earlier phase summaries, and the live green canonical gate
- Pitfalls: HIGH - Grounded in concrete current seams: untracked `rules.md`, Phase 4-branded verifier messaging, waiter contract drift, and the `Stop(nil)` inconsistency

**Research date:** 2026-03-13
**Valid until:** 2026-04-12
