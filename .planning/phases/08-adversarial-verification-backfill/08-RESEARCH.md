# Phase 8: Adversarial Verification Backfill - Research

**Researched:** 2026-03-14
**Domain:** Phase-level audit evidence backfill for adversarial verification
**Confidence:** HIGH

<user_constraints>
## User Constraints

No phase-specific `08-CONTEXT.md` exists for this phase.

### Locked Decisions
- Phase 8 is backfill work: re-verify the shipped Phase 4 adversarial regression scope with milestone-grade evidence rather than reopen broad implementation.
- The target artifact is `.planning/phases/04-adversarial-regression-matrix/04-VERIFICATION.md`.
- Phase 8 must satisfy `VER-05`.
- Phase 8 goal: Phase 4 adversarial regression work is re-verified with a milestone-grade verification artifact and reconciled validation evidence.
- Success criteria:
  - `.planning/phases/04-adversarial-regression-matrix/04-VERIFICATION.md` exists and verifies the shipped adversarial scope against current code.
  - Phase 4 validation evidence no longer presents executed checks as draft/pending-only audit debt.
  - Milestone audit traceability can point to a real Phase 4 verification artifact for `VER-05`.
- The output must include a `## Validation Architecture` section suitable for Nyquist validation scaffolding.

### Claude's Discretion
- Exact plan split for fresh re-verification, artifact backfill, and top-level traceability closure.
- Exact command text used to re-verify the adversarial slice and canonical workspace gate.
- Exact verification-report truths, artifact rows, and traceability checks used in `04-VERIFICATION.md`.

### Deferred Ideas (OUT OF SCOPE)
- New Phase 4 feature expansion beyond tiny fixes strictly required to keep verification claims truthful.
- Cleanup of other draft validation records not needed to close `VER-05`.
- Milestone archival work belongs after the Phase 8 backfill lands and the audit is rerun.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| VER-05 | Add fault-injection, fuzz/property-style, and compatibility/performance regression checks where needed to close verification gaps. | The shipped Phase 4 adversarial suites, fuzz-seed corpora, allocation gates, and canonical benchmark smoke already implement this scope. Phase 8 should backfill truthful milestone evidence for that shipped surface through `04-VERIFICATION.md`, `04-VALIDATION.md` reconciliation, and top-level audit/requirements traceability repair. |
</phase_requirements>

## Summary

Phase 8 should be planned as an audit-evidence backfill, not as a second adversarial implementation phase. The current repository already has the shipped Phase 4 artifacts that matter for product behavior: [`hsm_runtime_adversarial_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_runtime_adversarial_test.go), [`hsm_properties_fuzz_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_properties_fuzz_test.go), [`muid/muid_fuzz_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/muid/muid_fuzz_test.go), [`hsm_regression_allocs_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/hsm_regression_allocs_test.go), [`muid/muid_regression_allocs_test.go`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/muid/muid_regression_allocs_test.go), and the canonical gate in [`scripts/verify-workspace.sh`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/scripts/verify-workspace.sh). Fresh March 14, 2026 execution already confirmed that the shipped adversarial slice is green on current code.

The exact current-code evidence is strong enough to anchor the backfill directly. `GOCACHE="$PWD/.cache/go-build" go test ./... ./muid/... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|FuzzHSMEventScriptProperties|FuzzMUIDGeneratorProperties|Test.*RegressionAllocs' -count=1` passed on March 14, 2026 (`ok github.com/stateforward/hsm.go 0.223s`, `ok github.com/stateforward/hsm.go/muid 0.313s`). `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` also passed end to end on March 14, 2026, including the shipped benchmark-smoke steps for `BenchmarkNestedStates_NoEntryExitActivity` and `BenchmarkMUIDGeneration` and the existing race-short workspace pass.

The actual audit gap is documentary and traceability-specific. The milestone audit still marks `VER-05` orphaned because `.planning/phases/04-adversarial-regression-matrix/04-VERIFICATION.md` does not exist, and [`04-VALIDATION.md`](/Users/gabrielwillen/VSCode/stateforward/hsm/hsm.go/.planning/phases/04-adversarial-regression-matrix/04-VALIDATION.md) still reads as `status: draft` with all six task rows pending. That means the strongest Phase 8 plan is the same three-step shape used successfully in Phases 6 and 7: re-verify the shipped adversarial slice against current code, write a milestone-grade `04-VERIFICATION.md`, reconcile `04-VALIDATION.md` in place, then update top-level requirement and audit traceability so `VER-05` is no longer orphaned.

One concrete detail matters for truthful closeout: Phase 4 already expanded the canonical verifier to cover all three adversarial families, not just the targeted slice command. The backfilled report should explicitly mention deterministic adversarial runtime checks, ordinary `go test` replay of checked-in fuzz seeds, allocation regression suites, benchmark smoke, and the race-short workspace gate, because the canonical verifier currently proves all of them together.

**Primary recommendation:** Mirror the successful Phase 6 and Phase 7 backfill pattern for Phase 4: fresh adversarial-slice re-verification, in-place `04-VERIFICATION.md` plus `04-VALIDATION.md` reconciliation, then `VER-05` traceability and milestone-audit regeneration.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | `go1.25.3` | Repository toolchain for all adversarial verification commands | Already pinned by the repo and used by the shipped Phase 4 suites |
| `go test` | Go stdlib tool | Run targeted adversarial, fuzz-seed, and allocation suites | Existing Phase 4 validation and tests are already structured around direct `go test` invocations |
| `bash scripts/verify-workspace.sh` | repo-local | Canonical workspace release gate | Roadmap, state, audit, and prior verification reports treat this as the single authoritative gate |
| `rg` | repo-local CLI | Fast artifact and traceability checks | Existing verification patterns use direct file-presence and wiring checks rather than custom tooling |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `.planning/phases/04-adversarial-regression-matrix/04-01-SUMMARY.md` | repo doc | Shipped evidence for adversarial runtime, panic recovery, and helper contracts | Use when writing Phase 4 runtime truths in `04-VERIFICATION.md` |
| `.planning/phases/04-adversarial-regression-matrix/04-02-SUMMARY.md` | repo doc | Shipped evidence for fuzz/property targets and checked-in seed corpora | Use when mapping root and `muid` property coverage in the verification report |
| `.planning/phases/04-adversarial-regression-matrix/04-03-SUMMARY.md` | repo doc | Shipped evidence for allocation regression and canonical benchmark-smoke wiring | Use when proving closeout and canonical gate coverage in the verification report |
| `.planning/phases/04-adversarial-regression-matrix/04-VALIDATION.md` | repo doc | Reconciliation target for Nyquist evidence | Update in Phase 8 so it no longer communicates draft/pending debt |
| `.planning/phases/03-rules-conformance-matrix/03-VERIFICATION.md` | repo doc | Closest completed backfill verification template | Use for report shape and requirements coverage style |
| `.planning/phases/07-rules-conformance-verification-backfill/07-VALIDATION.md` | repo doc | Closest completed backfill validation template | Use for task-map and command-text structure |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Targeted adversarial rerun plus canonical gate | Canonical gate only | Too broad for `VER-05` traceability; it proves workspace health but hides which Phase 4 evidence families were re-verified directly |
| Backfilling `04-VERIFICATION.md` in the Phase 4 directory | Writing only Phase 8-local docs | Fails the audit goal because the orphan is specifically the missing Phase 4 verification artifact |
| Reconciling the original six Phase 4 validation rows in place | Inventing Phase 8-only validation rows for old work | Obscures the original Phase 4 execution shape and leaves Nyquist debt unresolved in the actual Phase 4 record |

**Installation:**
```bash
# No new packages.
# Use the repo toolchain and repo-local build cache for copy-safe verification.
go version
```

## Architecture Patterns

### Recommended Project Structure
```text
.planning/phases/04-adversarial-regression-matrix/
├── 04-01-SUMMARY.md        # Runtime adversarial and helper-contract evidence
├── 04-02-SUMMARY.md        # Root and MUID fuzz/property evidence
├── 04-03-SUMMARY.md        # Allocation regression and canonical-gate evidence
├── 04-VALIDATION.md        # Reconciled from draft/pending to executed evidence
└── 04-VERIFICATION.md      # New milestone-grade verification artifact

.planning/phases/08-adversarial-verification-backfill/
├── 08-RESEARCH.md          # Phase 8 planning input
└── 08-VALIDATION.md        # Per-plan verification contract for backfill execution
```

### Pattern 1: Re-Verify The Shipped Adversarial Slice Before Writing Docs
**What:** Treat current repository state as the product under audit and rerun the focused adversarial suites plus canonical gate before closing documentation.
**When to use:** First plan slice in Phase 8.
**Example:**
```bash
GOCACHE="$PWD/.cache/go-build" go test ./... ./muid/... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|FuzzHSMEventScriptProperties|FuzzMUIDGeneratorProperties|Test.*RegressionAllocs' -count=1
GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh
```

### Pattern 2: Write The Missing Report In The Original Phase Directory
**What:** Backfill `.planning/phases/04-adversarial-regression-matrix/04-VERIFICATION.md` using the same milestone-grade report shape already accepted for Phases 1, 2, 3, and 5.
**When to use:** Always. This is the missing primary evidence the audit needs.
**Example:**
```markdown
---
phase: 04-adversarial-regression-matrix
verified: 2026-03-14T00:00:00Z
status: passed
score: 8/8 must-haves verified
---

# Phase 4: Adversarial Regression Matrix Verification Report

## Goal Achievement

### Observable Truths
| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | The shipped adversarial runtime, fuzz-seed, allocation, and canonical-gate claims still hold on current code. | ✓ VERIFIED | Fresh March 14, 2026 rerun evidence plus shipped Phase 4 summaries. |
```

### Pattern 3: Reconcile Validation In Place Without Inventing New Work
**What:** Preserve the original six Phase 4 task rows in `04-VALIDATION.md`, but update frontmatter, commands, file-exists markers, and statuses to reflect executed evidence.
**When to use:** In the same plan that writes `04-VERIFICATION.md`.
**Example:**
```markdown
| 04-01-02 | 01 | 1 | VER-05 | unit-adversarial | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial' -count=1` | ✅ | ✅ green |
| 04-02-01 | 02 | 1 | VER-05 | fuzz-seed-root | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'FuzzHSMEventScriptProperties' -count=1` | ✅ | ✅ green |
| 04-02-02 | 02 | 1 | VER-05 | fuzz-seed-muid | `GOCACHE="$PWD/.cache/go-build" go test ./muid/... -run 'FuzzMUIDGeneratorProperties' -count=1` | ✅ | ✅ green |
| 04-03-01 | 03 | 2 | VER-05 | alloc-regression | `bash -lc 'GOCACHE="$PWD/.cache/go-build" go test ./... -run '\''Test.*RegressionAllocs'\'' -count=1 && GOCACHE="$PWD/.cache/go-build" go test ./muid/... -run '\''Test.*RegressionAllocs'\'' -count=1'` | ✅ | ✅ green |
| 04-03-02 | 03 | 2 | VER-05 | canonical-gate | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` | ✅ | ✅ green |
```

### Pattern 4: Close Top-Level Traceability The Same Way Phase 6 And Phase 7 Did
**What:** Once `04-VERIFICATION.md` exists and `04-VALIDATION.md` is reconciled, update `REQUIREMENTS.md` and regenerate the milestone audit so `VER-05` is satisfied through the Phase 8 backfill of the Phase 4 artifact.
**When to use:** Final plan slice in Phase 8.
**Example:**
```bash
bash -lc 'AUDIT=.planning/v1.0-MILESTONE-AUDIT.md \
  && rg -n "^- \\[x\\] \\*\\*VER-05\\*\\*" .planning/REQUIREMENTS.md \
  && rg -n "^\\| VER-05 \\| Phase 8 \\| Complete \\|" .planning/REQUIREMENTS.md \
  && rg -n "04-VERIFICATION\\.md" "$AUDIT" \
  && ! rg -n "VER-05.*orphaned|04-VERIFICATION\\.md does not exist" "$AUDIT"'
```

### Anti-Patterns to Avoid
- **Backfill in the wrong directory:** Phase 8 may own the work, but the missing primary evidence belongs in `.planning/phases/04-adversarial-regression-matrix/`.
- **Summary-only verification:** Do not write `04-VERIFICATION.md` from the three summaries alone; every major claim needs fresh current-code evidence.
- **Reopening Phase 4 implementation by default:** Fresh command evidence is already green. Only allow tiny fixes if reruns expose a false claim.
- **Leaving `04-VALIDATION.md` in draft form:** The audit debt is not closed if the validation record still reads like an unexecuted plan.
- **Treating canonical benchmark smoke as optional prose:** The current release gate already includes it, so the backfilled report must acknowledge it explicitly.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Phase verification format | A new custom doc schema | The existing `01-VERIFICATION.md`, `02-VERIFICATION.md`, `03-VERIFICATION.md`, and `05-VERIFICATION.md` structure | The audit already accepts this shape and later backfill phases stay consistent |
| Adversarial inventory | A regenerated test parser or new matrix format | Existing Phase 4 summaries plus direct file and command evidence | The shipped Phase 4 work is already divided cleanly by runtime, fuzz, and regression slices |
| Verification entrypoint | A new Phase 8-only verifier script | Existing `scripts/verify-workspace.sh` | One canonical gate is already a project decision and audit expectation |
| Validation history | New backfill-only task rows | The original six Phase 4 validation rows, reconciled in place | Keeps Nyquist evidence aligned with the shipped Phase 4 execution shape |
| Traceability proof | Narrative-only explanation in Phase 8 docs | `REQUIREMENTS.md` plus regenerated `v1.0-MILESTONE-AUDIT.md` | The orphaned requirement lives in top-level traceability, not only in phase-local prose |

**Key insight:** The hard part of Phase 8 is not proving the adversarial suites exist. That is already true. The hard part is turning shipped but scattered Phase 4 evidence into milestone-grade primary artifacts that the audit can consume automatically.

## Common Pitfalls

### Pitfall 1: Treating Phase 8 As New Adversarial Product Work
**What goes wrong:** The plan starts editing runtime, fuzz, or allocation suites without a fresh failing signal.
**Why it happens:** Phase 8 references `VER-05`, so it is easy to mistake the backfill for unfinished Phase 4 implementation.
**How to avoid:** Make the first task a fresh adversarial-slice rerun and require evidence of failure before allowing code changes.
**Warning signs:** The plan opens with new tests or new benchmark ideas instead of command re-verification.

### Pitfall 2: Writing The Verification Report In Phase 8 Instead Of Phase 4
**What goes wrong:** Work lands only under `.planning/phases/08-adversarial-verification-backfill/`.
**Why it happens:** Maintainers optimize for keeping all changes inside the current phase folder.
**How to avoid:** Treat Phase 8 as owning the missing Phase 4 artifact, exactly as Phase 6 owned the missing Phase 2 artifact and Phase 7 owned the missing Phase 3 artifact.
**Warning signs:** A proposed `08-VERIFICATION.md` appears while `04-VERIFICATION.md` is still absent.

### Pitfall 3: Leaving `04-VALIDATION.md` In Draft Form
**What goes wrong:** The audit still reports Nyquist partial debt even after `04-VERIFICATION.md` exists.
**Why it happens:** The team treats the verification report as sufficient and overlooks the validation record's `draft` frontmatter and pending task rows.
**How to avoid:** Reconcile frontmatter, command text, and statuses in the same plan that writes `04-VERIFICATION.md`.
**Warning signs:** `04-VALIDATION.md` still contains `status: draft`, `⬜ pending`, or `Approval: pending`.

### Pitfall 4: Narrowing Evidence To Only One Adversarial Family
**What goes wrong:** The verification report proves only runtime faults or only fuzz seeds, leaving other shipped Phase 4 claims implicit.
**Why it happens:** The phase shipped across three plans, so it is easy to over-focus on a single family of checks.
**How to avoid:** Require the report to cover runtime adversarial tests, root and `muid` fuzz seeds, allocation regression, benchmark smoke, and the canonical gate explicitly.
**Warning signs:** `04-VERIFICATION.md` never names `hsm_properties_fuzz_test.go`, `muid/muid_fuzz_test.go`, or the benchmark-smoke step.

### Pitfall 5: Closing `VER-05` Locally But Not Top-Level
**What goes wrong:** Phase 4 artifacts are fixed, but `REQUIREMENTS.md` and the milestone audit still show `VER-05` as pending/orphaned.
**Why it happens:** The team stops once the phase-local docs look complete.
**How to avoid:** Reserve a final traceability and audit plan slice and verify it with exact grepable checks.
**Warning signs:** `.planning/v1.0-MILESTONE-AUDIT.md` still lists Phase 4 as unverified after the backfill lands.

## Validation Architecture

Phase 8 should run in three waves that mirror the backfill workflow already proven in Phases 6 and 7.

### Wave 1
- Re-run the shipped Phase 4 adversarial slice with a repo-local build cache.
- Re-run the canonical workspace verifier on the same workspace state.
- Allow only tiny Phase 4-slice remediation if those reruns expose a false current claim.

### Wave 2
- Write `.planning/phases/04-adversarial-regression-matrix/04-VERIFICATION.md` from the fresh Wave 1 evidence plus the shipped Phase 4 summaries.
- Reconcile `.planning/phases/04-adversarial-regression-matrix/04-VALIDATION.md` in place so its original six rows no longer read as draft or pending debt.

### Wave 3
- Update `.planning/REQUIREMENTS.md` so `VER-05` is complete through Phase 8.
- Regenerate `.planning/v1.0-MILESTONE-AUDIT.md` so it recognizes `04-VERIFICATION.md` and stops reporting `VER-05` as orphaned.
- Preserve truthful milestone status based on whatever remaining debt exists after the regenerated audit, rather than forcing a passed audit outcome in advance.

### Command Family
- Focused rerun: `GOCACHE="$PWD/.cache/go-build" go test ./... ./muid/... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|FuzzHSMEventScriptProperties|FuzzMUIDGeneratorProperties|Test.*RegressionAllocs' -count=1`
- Canonical gate: `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- Artifact verification: `rg`-based presence and content checks for `04-VERIFICATION.md`, `04-VALIDATION.md`, `REQUIREMENTS.md`, and `v1.0-MILESTONE-AUDIT.md`

## Code Examples

Verified patterns from the current repository state:

### Targeted Adversarial Re-Verification
```bash
GOCACHE="$PWD/.cache/go-build" go test ./... ./muid/... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|FuzzHSMEventScriptProperties|FuzzMUIDGeneratorProperties|Test.*RegressionAllocs' -count=1
GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh
```
