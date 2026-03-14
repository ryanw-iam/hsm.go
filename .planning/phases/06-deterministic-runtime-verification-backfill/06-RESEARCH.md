# Phase 6: Deterministic Runtime Verification Backfill - Research

**Researched:** 2026-03-14
**Domain:** Phase-level audit evidence backfill for deterministic runtime verification
**Confidence:** HIGH

<user_constraints>
## User Constraints

- Treat this as evidence/backfill planning, not an excuse to re-open broad Phase 2 implementation.
- Assume current shipped code is the target; only verification, documentation reconciliation, or remediation directly required to keep evidence truthful should be planned.
- The plan must explicitly identify the likely commands, artifact shape, and must-have checks for a milestone-grade `02-VERIFICATION.md`.
- Include a section titled exactly `## Validation Architecture`.
- Call out checker pitfalls around planning scope, file ownership, and fake work.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| VER-03 | Add deterministic validation for concurrency, timing, dispatch ordering, and race-sensitive behavior. | Existing shipped Phase 2 runtime suites already cover timers, waiter synchronization, concurrency serialization, deferred replay, completion priority, multi-instance targeting, lifecycle restart/stop, and timeout/race-sensitive behavior. Phase 6 should backfill truthful milestone evidence for that shipped scope via `02-VERIFICATION.md` plus `02-VALIDATION.md` reconciliation. |
</phase_requirements>

## Summary

Phase 6 is not a runtime feature phase. The repository already contains the shipped deterministic runtime suites and the canonical gate passes against current code. The audit gap is documentary and traceability-specific: `.planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md` is missing, and `.planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md` still reads as an execution-time draft with all task rows marked pending.

The strongest planning shape is a single narrow evidence plan that re-verifies the existing Phase 2 scope against current code, writes a milestone-grade `02-VERIFICATION.md` modeled after Phases 1 and 5, and reconciles `02-VALIDATION.md` so it reflects executed checks rather than draft intent. Only if re-verification disproves an existing claim should the plan allow a tightly scoped remediation; otherwise Phase 6 should not touch runtime behavior.

**Primary recommendation:** Plan one documentation-and-verification pass that runs a Phase 2 targeted runtime command plus the canonical gate, writes `02-VERIFICATION.md`, and updates `02-VALIDATION.md` from draft/pending to truthful executed evidence.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | 1.25.3 | Repository toolchain for all verification commands | Defined in `go.mod`; every phase verifier and test suite depends on this exact workspace toolchain. |
| `go test` | Go 1.25.3 builtin | Run focused deterministic runtime suites | Existing Phase 2 validation and current runtime tests are already structured around direct `go test` invocations. |
| `bash scripts/verify-workspace.sh` | repo-local | Canonical release gate | Roadmap, state, Phase 1 verification, and Phase 5 verification all treat this as the single authoritative workspace verifier. |
| `rg` | repo-local CLI | Fast artifact and traceability checks | Existing verification patterns use direct file-presence and wiring checks; `rg` is the right tool for copy-safe documentation verification. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `.planning/phases/02-deterministic-runtime-semantics/02-01-SUMMARY.md` | repo doc | Evidence input for timer harness and clock semantics claims | Use when writing Observable Truths and Required Artifacts for timer scope. |
| `.planning/phases/02-deterministic-runtime-semantics/02-02-SUMMARY.md` | repo doc | Evidence input for concurrency, deferred replay, and multi-instance targeting claims | Use when mapping shipped tests to phase scope. |
| `.planning/phases/02-deterministic-runtime-semantics/02-03-SUMMARY.md` | repo doc | Evidence input for lifecycle timeout and race-sensitive claims | Use when verifying stop/restart/timeout coverage and race-gate closure. |
| `.planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md` | repo doc | Reconciliation target for Nyquist evidence | Update this in Phase 6 so it no longer claims draft/pending-only state. |
| `.planning/phases/01-coverage-baseline-and-release-gates/01-VERIFICATION.md` | repo doc | Minimal milestone-grade verification template | Use for report shape, tables, and goal-backward verification style. |
| `.planning/phases/05-hardening-remediation-closure/05-VERIFICATION.md` | repo doc | Re-verification template with stronger artifact and command wiring checks | Use when Phase 6 needs direct command proof and copy-safe documentation checks. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Phase-targeted runtime command plus canonical gate | Canonical gate only | Too broad for milestone traceability. It proves workspace health, but not Phase 2 scope explicitly enough for `VER-03`. |
| Backfilling `02-VERIFICATION.md` in the Phase 2 directory | Writing only Phase 6 docs in the Phase 6 directory | Fails the audit goal. The orphan is specifically the missing Phase 2 verification artifact. |
| Reconciling `02-VALIDATION.md` in place | Leaving `02-VALIDATION.md` draft and explaining it in `06-RESEARCH.md` or plan docs | Fake closure. The audit debt is that the Phase 2 validation artifact itself still says draft/pending. |

**Installation:**
```bash
# No new packages.
# Use the repo toolchain and repo-local build cache for copy-safe verification.
go version
```

## Architecture Patterns

### Recommended Project Structure
```text
.planning/phases/02-deterministic-runtime-semantics/
├── 02-01-SUMMARY.md        # Timer harness and deterministic clock scope
├── 02-02-SUMMARY.md        # Concurrency, queueing, and multi-instance scope
├── 02-03-SUMMARY.md        # Lifecycle, timeout, and race-sensitive scope
├── 02-VALIDATION.md        # Reconciled from draft/pending to executed evidence
└── 02-VERIFICATION.md      # New milestone-grade verification artifact
```

### Pattern 1: Goal-Backward Verification Report
**What:** Build `02-VERIFICATION.md` in the same shape as `01-VERIFICATION.md` and `05-VERIFICATION.md`: frontmatter, phase goal, Observable Truths, Required Artifacts, Key Link Verification, Requirements Coverage, Anti-Patterns Found, Human Verification Required, Gaps Summary, and Verification Metadata.
**When to use:** Always. This is the missing audit artifact.
**Example:**
```markdown
---
phase: 02-deterministic-runtime-semantics
verified: 2026-03-14T00:00:00Z
status: passed
score: 8/8 must-haves verified
---

# Phase 2: Deterministic Runtime Semantics Verification Report

## Goal Achievement

### Observable Truths
| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Timer-driven transitions are verified without wall-clock sleeps. | ✓ VERIFIED | `hsm_runtime_timers_test.go` plus targeted command pass. |
```

### Pattern 2: Re-Verify Shipped Scope, Do Not Re-Implement It
**What:** Treat the current repository state as the product under audit. Re-run the existing deterministic runtime suites and document what they prove now.
**When to use:** For every Phase 2 claim in the verification report.
**Example:**
```bash
GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(AfterUsesDeterministicTimer|AfterNegativeDurationDoesNotSchedule|AtUsesDeterministicTimer|EveryUsesManualTriggers|WhenWaitsForExplicitSignal|ConfigClockOverridesDefaultClock|DefaultClockFallbackUsesDeterministicTimer|ConcurrentDispatchSerializesProcessing|DeferredEventsReplayAfterTransition|CompletionEventsPreemptQueuedEvents|DispatchAllReachesEveryStartedInstance|DispatchToTargetsMatchingInstancesOnly|StopCancelsActivityAndClosesAfterExecuted|RestartReentersInitialStateWithData|StopTimeoutDispatchesErrorEventDeterministically)$' -count=1
```

### Pattern 3: Reconcile Validation In Place
**What:** Update `02-VALIDATION.md` so its frontmatter is no longer `status: draft`, its task rows no longer say `⬜ pending`, and its commands reflect the copy-safe commands actually used in verification.
**When to use:** In the same plan that writes `02-VERIFICATION.md`.
**Example:**
```markdown
| 02-02-01 | 02 | 2 | VER-03 | concurrency/order | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(ConcurrentDispatchSerializesProcessing|DeferredEventsReplayAfterTransition|CompletionEventsPreemptQueuedEvents)$' -count=1` | ✅ | ✅ green |
```

### Anti-Patterns to Avoid
- **Backfill in the wrong directory:** Phase 6 can have its own research and plans, but the missing audit artifact must be written to `.planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md`.
- **Summary-only verification:** Do not write `02-VERIFICATION.md` from the three summary files alone. Every claim needs current-code evidence.
- **Broad runtime reopening:** Do not treat Phase 6 as permission to refactor `hsm.go`, rename tests, or redesign the deterministic harness unless re-verification exposes a real false claim.
- **Gate-only storytelling:** Do not rely only on `bash scripts/verify-workspace.sh`. It is necessary but not sufficiently Phase-2-specific by itself.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Phase verification format | A new custom doc schema | The existing `01-VERIFICATION.md` / `05-VERIFICATION.md` structure | The audit already accepts this shape, and it keeps later backfill phases consistent. |
| Runtime evidence selection | New one-off helper scripts or ad hoc shell transcripts | Existing Phase 2 runtime suites plus direct `go test` commands | The shipped tests already encode the runtime contract precisely. |
| Audit closure | Manual edits to `v1.0-MILESTONE-AUDIT.md` without underlying proof | Real `02-VERIFICATION.md` plus reconciled `02-VALIDATION.md` | The audit orphan is caused by missing primary evidence, not by missing prose in the audit file. |
| Validation closure | An explanatory note in Phase 6 docs | Direct updates to `02-VALIDATION.md` | The current debt is that the artifact itself still says `draft` and `pending`. |

**Key insight:** This phase closes an evidence gap by using the artifacts that already exist. New mechanisms would increase the risk of fake work and weaken audit traceability.

## Common Pitfalls

### Pitfall 1: Scope Creep Into Runtime Feature Work
**What goes wrong:** The plan starts changing tests, helpers, or `hsm.go` because Phase 2 touched runtime behavior historically.
**Why it happens:** The backfill phase is attached to a runtime phase, so maintainers may assume runtime edits are expected.
**How to avoid:** Make the first plan task artifact-only: run verification, inventory the shipped scope, and draft `02-VERIFICATION.md`. Allow code changes only behind an explicit "evidence is false against current code" condition.
**Warning signs:** Plan tasks mention refactors, helper redesign, test reorganization, or "clean up legacy Phase 2 coverage."

### Pitfall 2: Fake Verification Through Summaries
**What goes wrong:** `02-VERIFICATION.md` repeats summary claims without re-running commands.
**Why it happens:** The summaries are detailed and easy to paraphrase.
**How to avoid:** Require every Observable Truth to cite a current artifact and a current command result.
**Warning signs:** Evidence lines reference only `02-01-SUMMARY.md`, `02-02-SUMMARY.md`, and `02-03-SUMMARY.md` with no executed command evidence.

### Pitfall 3: Validation Theater
**What goes wrong:** `02-VALIDATION.md` stays `status: draft` or keeps `⬜ pending` rows while a new report claims the phase is verified.
**Why it happens:** The verification report gets treated as sufficient, and validation reconciliation is deferred.
**How to avoid:** Put validation reconciliation in the same plan as `02-VERIFICATION.md`.
**Warning signs:** Phase 6 completes without touching `.planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md`.

### Pitfall 4: Wrong File Ownership
**What goes wrong:** Work lands only in the Phase 6 directory, leaving Phase 2 evidence unchanged.
**Why it happens:** Maintainers optimize for "keep changes inside the current phase folder."
**How to avoid:** Treat Phase 6 as owning Phase 2's missing evidence artifacts.
**Warning signs:** The plan writes `06-VERIFICATION.md` or `06-VALIDATION.md` instead of backfilling `02-VERIFICATION.md` and reconciling `02-VALIDATION.md`.

### Pitfall 5: Generic Commands That Are Not Copy-Safe
**What goes wrong:** Docs keep `go test ./...` and `bash scripts/verify-workspace.sh` without the repo-local cache prefix, then fail in stricter environments.
**Why it happens:** Phase 2 summaries were written during an environment that needed `GOCACHE`, but `02-VALIDATION.md` still uses generic commands.
**How to avoid:** Standardize on `GOCACHE="$PWD/.cache/go-build"` in Phase 6 docs.
**Warning signs:** Validation or verification commands omit `GOCACHE` even though later phase docs already use it.

### Pitfall 6: Missing Traceability Reconciliation
**What goes wrong:** The phase closes its report and validation docs, but leaves requirement or audit traceability inconsistent.
**Why it happens:** The team treats `02-VERIFICATION.md` as the only output and misses surrounding traceability tables.
**How to avoid:** Check `REQUIREMENTS.md`, roadmap traceability, and any follow-on audit rerun expectations before closing Phase 6.
**Warning signs:** `VER-03` remains pending or audit text still says the verification file is missing after the backfill lands.

## Code Examples

Verified patterns from the current repository state:

### Phase-Targeted Runtime Re-Verification
```bash
GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(AfterUsesDeterministicTimer|AfterNegativeDurationDoesNotSchedule|AtUsesDeterministicTimer|EveryUsesManualTriggers|WhenWaitsForExplicitSignal|ConfigClockOverridesDefaultClock|DefaultClockFallbackUsesDeterministicTimer|ConcurrentDispatchSerializesProcessing|DeferredEventsReplayAfterTransition|CompletionEventsPreemptQueuedEvents|DispatchAllReachesEveryStartedInstance|DispatchToTargetsMatchingInstancesOnly|StopCancelsActivityAndClosesAfterExecuted|RestartReentersInitialStateWithData|StopTimeoutDispatchesErrorEventDeterministically)$' -count=1
```

### Canonical Gate Confirmation
```bash
GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh
```

### Documentation Reconciliation Check
```bash
bash -lc "test -f .planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md \
  && rg -n 'Status: passed|Observable Truths|Required Artifacts|Key Link Verification|Requirements Coverage' .planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md \
  && ! rg -n '^status: draft|⬜ pending|Approval: pending' .planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md"
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Plan summaries claim `VER-03` completion without a phase verification report | Milestone-grade `02-VERIFICATION.md` proves the shipped scope against current code | Needed by 2026-03-14 audit | Restores audit traceability for Phase 2. |
| Validation doc records intended commands with `draft` + `pending` rows | Validation doc records executed commands and completed statuses | Needed by 2026-03-14 audit | Removes Nyquist partial-status debt for Phase 2. |
| Generic `go test ./...` command in `02-VALIDATION.md` | Copy-safe repo-local-cache commands that match current verifier practice | Established by later phase docs by 2026-03-14 | Makes backfill evidence runnable in stricter environments. |

**Deprecated/outdated:**
- `02-VALIDATION.md` frontmatter `status: draft` for Phase 2: outdated for a completed, shipped phase once backfill begins.
- `02-VALIDATION.md` task rows marked `⬜ pending`: outdated once Phase 6 re-verifies and reconciles the actual evidence.

## Open Questions

1. **Should Phase 6 also reconcile `REQUIREMENTS.md` if `VER-03` remains pending there?**
   - What we know: The current `REQUIREMENTS.md` file lists `VER-03` as unchecked, but the milestone audit text describes it as already checked and orphaned only because `02-VERIFICATION.md` is missing.
   - What's unclear: Whether the audit text is stale or Phase 6 is expected to update the requirement table as part of closure.
   - Recommendation: Treat this as an explicit planning checkpoint. If `VER-03` still reads pending after backfill, update the requirement traceability in the same phase so evidence stays internally consistent.

2. **Does the milestone audit get regenerated in Phase 6 or only after Phases 6-8 land?**
   - What we know: The audit file currently records the missing `02-VERIFICATION.md` and draft validation debt.
   - What's unclear: Whether a new audit run is part of each backfill phase or part of the final multi-phase closure.
   - Recommendation: Do not hand-edit the current audit as substitute evidence. Either rerun the audit after the artifact exists, or leave audit regeneration to the final closure phase, but make the decision explicit in the plan.

3. **What exact non-draft status token should `02-VALIDATION.md` use?**
   - What we know: The repo currently shows only draft validation files, so there is no existing final-state convention to copy.
   - What's unclear: Whether the expected terminal status is `completed`, `reconciled`, `passed`, or another repo-specific token.
   - Recommendation: Pick one convention for Phases 6-8 and apply it consistently, but the hard requirement is that the file no longer communicates draft/pending-only debt.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | `go test` plus shell-level artifact checks |
| Config file | `go.mod` |
| Quick run command | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(AfterUsesDeterministicTimer|AfterNegativeDurationDoesNotSchedule|AtUsesDeterministicTimer|EveryUsesManualTriggers|WhenWaitsForExplicitSignal|ConfigClockOverridesDefaultClock|DefaultClockFallbackUsesDeterministicTimer|ConcurrentDispatchSerializesProcessing|DeferredEventsReplayAfterTransition|CompletionEventsPreemptQueuedEvents|DispatchAllReachesEveryStartedInstance|DispatchToTargetsMatchingInstancesOnly|StopCancelsActivityAndClosesAfterExecuted|RestartReentersInitialStateWithData|StopTimeoutDispatchesErrorEventDeterministically)$' -count=1` |
| Full suite command | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VER-03 | Deterministic timer scheduling, negative-duration no-fire behavior, `At`, `Every`, `When`, and clock precedence/fallback | unit | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(AfterUsesDeterministicTimer|AfterNegativeDurationDoesNotSchedule|AtUsesDeterministicTimer|EveryUsesManualTriggers|WhenWaitsForExplicitSignal|ConfigClockOverridesDefaultClock|DefaultClockFallbackUsesDeterministicTimer)$' -count=1` | ✅ |
| VER-03 | Concurrent dispatch serialization, deferred replay, completion-event priority, broadcast delivery, and targeted delivery | unit | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(ConcurrentDispatchSerializesProcessing|DeferredEventsReplayAfterTransition|CompletionEventsPreemptQueuedEvents|DispatchAllReachesEveryStartedInstance|DispatchToTargetsMatchingInstancesOnly)$' -count=1` | ✅ |
| VER-03 | Stop cancellation, restart semantics, timeout-driven error dispatch, and race-sensitive lifecycle behavior | unit + gate | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(StopCancelsActivityAndClosesAfterExecuted|RestartReentersInitialStateWithData|StopTimeoutDispatchesErrorEventDeterministically)$' -count=1 && GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` | ✅ |
| VER-03 | Milestone-grade evidence exists and validation debt is reconciled | artifact | `bash -lc "test -f .planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md && ! rg -n '^status: draft|⬜ pending|Approval: pending' .planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md"` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** Run the smallest relevant Phase 2 targeted `go test` command plus any affected artifact check.
- **Per wave merge:** Run `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`.
- **Phase gate:** All Phase 2 targeted commands, artifact checks, and the canonical gate must be green before `/gsd:verify-work`.

### Wave 0 Gaps
- [ ] `.planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md` - missing required milestone artifact for VER-03
- [ ] `.planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md` - still `status: draft`
- [ ] `.planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md` - all six task rows still `⬜ pending`
- [ ] `.planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md` - commands should be reconciled to copy-safe `GOCACHE="$PWD/.cache/go-build"` form used by later phase docs
- [ ] `REQUIREMENTS.md` traceability for `VER-03` - confirm and reconcile if still inconsistent with audit expectations

## Sources

### Primary (HIGH confidence)
- `.planning/ROADMAP.md` - Phase 6 goal, success criteria, and dependency position
- `.planning/REQUIREMENTS.md` - current requirement ownership and traceability state
- `.planning/v1.0-MILESTONE-AUDIT.md` - exact orphaned `VER-03` gap and validation-draft debt
- `.planning/phases/02-deterministic-runtime-semantics/02-VALIDATION.md` - current draft/pending validation state
- `.planning/phases/02-deterministic-runtime-semantics/02-01-SUMMARY.md` - timer and harness scope
- `.planning/phases/02-deterministic-runtime-semantics/02-02-SUMMARY.md` - concurrency and multi-instance scope
- `.planning/phases/02-deterministic-runtime-semantics/02-03-SUMMARY.md` - lifecycle and race-sensitive scope
- `.planning/phases/01-coverage-baseline-and-release-gates/01-VERIFICATION.md` - accepted milestone-grade verification format
- `.planning/phases/05-hardening-remediation-closure/05-VERIFICATION.md` - accepted re-verification format with direct command checks
- `hsm_runtime_timers_test.go` - shipped timer and clock determinism coverage
- `hsm_runtime_concurrency_test.go` - shipped concurrency, queueing, and multi-instance coverage
- `hsm_runtime_lifecycle_test.go` - shipped lifecycle and timeout coverage
- `hsm_test_helpers_test.go` - shipped deterministic harness and waiter helpers
- `scripts/verify-workspace.sh` - canonical gate contents
- Executed locally on 2026-03-14:
  - `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntime(AfterUsesDeterministicTimer|AfterNegativeDurationDoesNotSchedule|AtUsesDeterministicTimer|EveryUsesManualTriggers|WhenWaitsForExplicitSignal|ConfigClockOverridesDefaultClock|DefaultClockFallbackUsesDeterministicTimer|ConcurrentDispatchSerializesProcessing|DeferredEventsReplayAfterTransition|CompletionEventsPreemptQueuedEvents|DispatchAllReachesEveryStartedInstance|DispatchToTargetsMatchingInstancesOnly|StopCancelsActivityAndClosesAfterExecuted|RestartReentersInitialStateWithData|StopTimeoutDispatchesErrorEventDeterministically)$' -count=1`
  - `GOCACHE="$PWD/.cache/go-build" go test -race -short ./... ./kind/... ./muid/...`
  - `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`

### Secondary (MEDIUM confidence)
- None needed. This research is repository-local.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - derived from the current repository toolchain, existing verification artifacts, and commands executed successfully during research.
- Architecture: HIGH - based on accepted Phase 1 and Phase 5 verifier patterns plus the actual Phase 2 shipped artifact layout.
- Pitfalls: HIGH - directly reflected by the current milestone audit gap, validation debt, and requirement-traceability inconsistency.

**Research date:** 2026-03-14
**Valid until:** 2026-03-21
