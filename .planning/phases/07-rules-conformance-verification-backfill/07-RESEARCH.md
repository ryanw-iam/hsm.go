# Phase 7: Rules Conformance Verification Backfill - Research

**Researched:** 2026-03-14
**Domain:** Phase-level audit evidence backfill for rules conformance verification
**Confidence:** HIGH

<user_constraints>
## User Constraints

No phase-specific `07-CONTEXT.md` exists for this phase.

### Locked Decisions
- Phase 7 is backfill work: re-verify the shipped Phase 3 rules conformance scope with milestone-grade evidence rather than reopen broad implementation.
- The target artifact is `.planning/phases/03-rules-conformance-matrix/03-VERIFICATION.md`.
- Phase 7 must satisfy `VER-04`.
- Phase 7 goal: Phase 3 rules conformance work is re-verified with a milestone-grade verification artifact and reconciled validation evidence.
- Success criteria:
  - `.planning/phases/03-rules-conformance-matrix/03-VERIFICATION.md` exists and verifies the shipped rules conformance scope against current code.
  - Phase 3 validation evidence no longer presents executed checks as draft/pending-only audit debt.
  - Milestone audit traceability can point to a real Phase 3 verification artifact for `VER-04`.
- The output must include a `## Validation Architecture` section suitable for Nyquist validation scaffolding.

### Claude's Discretion
- Exact plan split for fresh re-verification, artifact backfill, and top-level traceability closure.
- Exact command text used to re-verify the rules slice and canonical workspace gate.
- Exact verification-report truths, artifact table rows, and traceability checks used in `03-VERIFICATION.md`.

### Deferred Ideas (OUT OF SCOPE)
- New Phase 3 feature expansion beyond tiny fixes strictly required to keep verification claims truthful.
- Phase 4 adversarial backfill work belongs to Phase 8.
- Cleanup of other draft validation records not needed to close `VER-04`.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| VER-04 | Add conformance coverage for the library rules captured in `rules.md`. | The shipped Phase 3 matrix and dedicated rule-ID suites already implement this coverage. Phase 7 should backfill truthful milestone evidence for that shipped scope through `03-VERIFICATION.md`, `03-VALIDATION.md` reconciliation, and top-level audit/requirements traceability repair. |
</phase_requirements>

## Summary

Phase 7 should be planned as an audit-evidence backfill, not as a second rules-conformance implementation phase. The current repository already has the shipped Phase 3 artifacts that matter for product behavior: the checked-in `03-RULES-MATRIX.md`, the dedicated `TestModelRulesConformance` suite, the dedicated `TestRuntimeRulesConformance` suite, and a green canonical workspace verifier. Fresh March 14, 2026 execution confirmed `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance' -count=1` passed (`ok github.com/stateforward/hsm.go 0.338s`) and `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` passed end to end on the current workspace state.

The actual audit gap is documentary and traceability-specific. The milestone audit still marks Phase 3 unverified because `.planning/phases/03-rules-conformance-matrix/03-VERIFICATION.md` does not exist, and `03-VALIDATION.md` still reads as `status: draft` with all task rows pending. That means the strongest Phase 7 plan is the same three-step shape used successfully in Phase 6: re-verify the shipped rules slice against current code, write a milestone-grade `03-VERIFICATION.md`, reconcile `03-VALIDATION.md` in place, then update top-level requirement/audit traceability so `VER-04` is no longer orphaned.

One concrete pitfall is already verified: the old Phase 3 validation commands are not copy-safe in this environment. Bare `go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance' -count=1` failed on March 14, 2026 with a default Go build-cache permission error, while the repo-local-cache form succeeded. Phase 7 should standardize `GOCACHE="$PWD/.cache/go-build"` in all backfilled verification and validation commands just as Phase 6 did for `VER-03`.

**Primary recommendation:** Mirror the Phase 6 backfill pattern for Phase 3: fresh rules-slice re-verification, in-place `03-VERIFICATION.md` plus `03-VALIDATION.md` reconciliation, then `VER-04` traceability and audit regeneration.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go toolchain | `go1.25.3` | Repository toolchain for all rules verification commands | Already pinned by the repo and used by the shipped Phase 3 suites |
| `go test` | Go stdlib tool | Run focused rules conformance suites | Existing Phase 3 validation and tests are already structured around direct `go test` invocations |
| `bash scripts/verify-workspace.sh` | repo-local | Canonical workspace release gate | Roadmap, state, audit, and prior verification reports treat this as the single authoritative gate |
| `rg` | repo-local CLI | Fast artifact and traceability checks | Existing verification patterns use direct file-presence and wiring checks rather than custom tooling |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `.planning/phases/03-rules-conformance-matrix/03-RULES-MATRIX.md` | repo doc | Rule inventory and rule-ID evidence mapping | Use when writing observable truths and verifying matrix-to-test traceability |
| `.planning/phases/03-rules-conformance-matrix/03-01-SUMMARY.md` | repo doc | Shipped evidence for matrix classification and panic helper contract | Use when verifying define-panic inventory and helper assumptions |
| `.planning/phases/03-rules-conformance-matrix/03-02-SUMMARY.md` | repo doc | Shipped evidence for define-panic conformance | Use when mapping `define_panic` claims in `03-VERIFICATION.md` |
| `.planning/phases/03-rules-conformance-matrix/03-03-SUMMARY.md` | repo doc | Shipped evidence for runtime and exemplar conformance | Use when mapping runtime/exemplar claims in `03-VERIFICATION.md` |
| `.planning/phases/03-rules-conformance-matrix/03-04-SUMMARY.md` | repo doc | Shipped evidence for matrix reconciliation and canonical gate proof | Use when proving closeout and gate coverage in the verification report |
| `.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md` | repo doc | Reconciliation target for Nyquist evidence | Update in Phase 7 so it no longer communicates draft/pending debt |
| `.planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md` | repo doc | Closest completed backfill template | Use for Phase 7 report shape and traceability style |
| `.planning/phases/06-deterministic-runtime-verification-backfill/06-VALIDATION.md` | repo doc | Closest completed backfill validation template | Use for task-map and command-text structure |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Targeted rules command plus canonical gate | Canonical gate only | Too broad for `VER-04` traceability; it proves workspace health but not the specific Phase 3 rules slice clearly enough |
| Backfilling `03-VERIFICATION.md` in the Phase 3 directory | Writing only Phase 7-local docs | Fails the audit goal because the orphan is specifically the missing Phase 3 verification artifact |
| Reconciling the original five Phase 3 validation rows in place | Inventing Phase 7-only validation rows for old work | Obscures the original Phase 3 execution shape and leaves Nyquist debt unresolved in the actual Phase 3 record |

**Installation:**
```bash
# No new packages.
# Use the repo toolchain and repo-local build cache for copy-safe verification.
go version
```

## Architecture Patterns

### Recommended Project Structure
```text
.planning/phases/03-rules-conformance-matrix/
├── 03-01-SUMMARY.md        # Matrix inventory and helper contract
├── 03-02-SUMMARY.md        # Define-panic suite evidence
├── 03-03-SUMMARY.md        # Runtime/exemplar suite evidence
├── 03-04-SUMMARY.md        # Matrix closeout and canonical gate evidence
├── 03-RULES-MATRIX.md      # Rule -> enforcement -> rule-ID test mapping
├── 03-VALIDATION.md        # Reconciled from draft/pending to executed evidence
└── 03-VERIFICATION.md      # New milestone-grade verification artifact

.planning/phases/07-rules-conformance-verification-backfill/
└── 07-RESEARCH.md          # Phase 7 planning input
```

### Pattern 1: Re-Verify The Shipped Rules Slice Before Writing Docs
**What:** Treat current repository state as the product under audit and rerun the focused rules suites plus canonical gate before closing documentation.
**When to use:** First plan slice in Phase 7.
**Example:**
```bash
GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance' -count=1
GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh
```

### Pattern 2: Write The Missing Report In The Original Phase Directory
**What:** Backfill `.planning/phases/03-rules-conformance-matrix/03-VERIFICATION.md` using the same milestone-grade report shape already accepted for Phases 1, 2, and 5.
**When to use:** Always. This is the missing primary evidence the audit needs.
**Example:**
```markdown
---
phase: 03-rules-conformance-matrix
verified: 2026-03-14T00:00:00Z
status: passed
score: 7/7 must-haves verified
---

# Phase 3: Rules Conformance Matrix Verification Report

## Goal Achievement

### Observable Truths
| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | The shipped rules matrix still reconciles one-to-one with `rules.md` and real rule-ID test targets. | ✓ VERIFIED | `03-RULES-MATRIX.md` plus fresh matrix integrity check passed. |
```

### Pattern 3: Reconcile Validation In Place Without Inventing New Work
**What:** Preserve the original five Phase 3 task rows in `03-VALIDATION.md`, but update frontmatter, commands, file-exists markers, and statuses to reflect executed evidence.
**When to use:** In the same plan that writes `03-VERIFICATION.md`.
**Example:**
```markdown
| 03-02-01 | 02 | 2 | VER-04 | define-panic conformance | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance' -count=1` | ✅ | ✅ green |
| 03-03-01 | 03 | 2 | VER-04 | runtime/exemplar conformance | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeRulesConformance' -count=1` | ✅ | ✅ green |
| 03-04-01 | 04 | 3 | VER-04 | matrix closeout and canonical gate | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` | ✅ | ✅ green |
```

### Pattern 4: Close Top-Level Traceability The Same Way Phase 6 Did
**What:** Once `03-VERIFICATION.md` exists and `03-VALIDATION.md` is reconciled, update `REQUIREMENTS.md` and regenerate the milestone audit so `VER-04` is satisfied through the Phase 7 backfill of the Phase 3 artifact.
**When to use:** Final plan slice in Phase 7.
**Example:**
```bash
bash -lc 'AUDIT=.planning/v1.0-MILESTONE-AUDIT.md \
  && rg -n "^- \\[x\\] \\*\\*VER-04\\*\\*" .planning/REQUIREMENTS.md \
  && rg -n "^\\| VER-04 \\| Phase 7 \\| Complete \\|" .planning/REQUIREMENTS.md \
  && rg -n "03-VERIFICATION\\.md" "$AUDIT" \
  && ! rg -n "VER-04.*orphaned|03-VERIFICATION\\.md does not exist" "$AUDIT" \
  && rg -n "VER-05" "$AUDIT"'
```

### Anti-Patterns to Avoid
- **Backfill in the wrong directory:** Phase 7 may own the work, but the missing primary evidence belongs in `.planning/phases/03-rules-conformance-matrix/`.
- **Summary-only verification:** Do not write `03-VERIFICATION.md` from the four summaries alone; every major claim needs fresh current-code evidence.
- **Reopening Phase 3 implementation by default:** Fresh command evidence is already green. Only allow tiny fixes if reruns expose a false claim.
- **Leaving generic commands in place:** `go test ./...` without repo-local `GOCACHE` is already known to fail in this environment.
- **Hand-editing audit prose as substitute evidence:** Regenerate the audit from real artifacts after the backfill lands.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Phase verification format | A new custom doc schema | The existing `01-VERIFICATION.md`, `02-VERIFICATION.md`, and `05-VERIFICATION.md` structure | The audit already accepts this shape and later backfill phases stay consistent |
| Rules inventory | A regenerated parser or new matrix format | Existing `03-RULES-MATRIX.md` plus focused integrity checks | The matrix already maps every `HSMxx` rule to a real evidence target |
| Verification entrypoint | A new Phase 7-only verifier script | Existing `scripts/verify-workspace.sh` | One canonical gate is already a project decision and audit expectation |
| Validation history | New backfill-only task rows | The original five Phase 3 validation rows, reconciled in place | Keeps Nyquist evidence aligned with the shipped Phase 3 execution shape |
| Traceability proof | Narrative-only explanation in Phase 7 docs | `REQUIREMENTS.md` plus regenerated `v1.0-MILESTONE-AUDIT.md` | The orphaned requirement lives in top-level traceability, not only in phase-local prose |

**Key insight:** The hard part of Phase 7 is not proving the rules tests exist. That is already true. The hard part is turning scattered but real Phase 3 evidence into milestone-grade primary artifacts that the audit can consume automatically.

## Common Pitfalls

### Pitfall 1: Treating Phase 7 As New Rules Work
**What goes wrong:** The plan starts editing model/runtime suites or `rules.md` without a fresh failing signal.
**Why it happens:** Phase 7 references `VER-04`, so it is easy to mistake the backfill for unfinished Phase 3 product work.
**How to avoid:** Make the first task a fresh rules-slice rerun and require evidence of failure before allowing code changes.
**Warning signs:** The plan opens with test refactors instead of command re-verification.

### Pitfall 2: Writing The Verification Report In Phase 7 Instead Of Phase 3
**What goes wrong:** Work lands only under `.planning/phases/07-rules-conformance-verification-backfill/`.
**Why it happens:** Maintainers optimize for keeping all changes inside the current phase folder.
**How to avoid:** Treat Phase 7 as owning the missing Phase 3 artifact, exactly as Phase 6 owned the missing Phase 2 artifact.
**Warning signs:** A proposed `07-VERIFICATION.md` appears while `03-VERIFICATION.md` is still absent.

### Pitfall 3: Leaving `03-VALIDATION.md` In Draft Form
**What goes wrong:** The audit still reports Nyquist partial debt even after `03-VERIFICATION.md` exists.
**Why it happens:** The team treats the verification report as sufficient and overlooks the validation record's `draft` frontmatter and pending task rows.
**How to avoid:** Reconcile frontmatter, command text, and statuses in the same plan that writes `03-VERIFICATION.md`.
**Warning signs:** `03-VALIDATION.md` still contains `status: draft`, `⬜ pending`, or `Approval: pending`.

### Pitfall 4: Copying The Old Bare `go test` Commands Forward
**What goes wrong:** Backfilled validation commands fail when maintainers run them in stricter environments.
**Why it happens:** The original Phase 3 validation file predates the later repo-local-cache convention.
**How to avoid:** Standardize on `GOCACHE="$PWD/.cache/go-build"` everywhere in Phase 7 docs.
**Warning signs:** Any new verification or validation command omits `GOCACHE` despite the fresh March 14, 2026 bare-command failure.

### Pitfall 5: Breaking Matrix-To-Test Traceability During Reconciliation
**What goes wrong:** The verification report claims rule coverage broadly, but the matrix, summaries, and actual subtest names drift apart.
**Why it happens:** The planner treats `03-RULES-MATRIX.md` as background context instead of the primary traceability artifact.
**How to avoid:** Preserve matrix-first evidence language and keep rule-ID test names stable.
**Warning signs:** Proposed edits rename `TestModelRulesConformance/HSMxx/...` or `TestRuntimeRulesConformance/HSMxx/...` without a matching matrix update.

### Pitfall 6: Closing `VER-04` Locally But Not Top-Level
**What goes wrong:** Phase 3 artifacts are fixed, but `REQUIREMENTS.md` and the milestone audit still show `VER-04` as pending/orphaned.
**Why it happens:** The team stops once the phase-local docs look complete.
**How to avoid:** Reserve a final traceability/audit plan slice and verify it with exact grepable checks.
**Warning signs:** `.planning/v1.0-MILESTONE-AUDIT.md` still lists Phase 3 as unverified after the backfill lands.

## Code Examples

Verified patterns from the current repository state:

### Targeted Rules Re-Verification
```bash
GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance' -count=1
```

### Canonical Gate Confirmation
```bash
GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh
```

### Matrix Integrity Check
```bash
bash -lc 'RULES=$(sed -n '\''s/^- \(HSM[0-9][0-9]\):.*/\1/p'\'' rules.md | sort | tr "\n" " "); \
  MATRIX=$(awk -F'\''|'\'' '\''/^\\| HSM[0-9][0-9] / {gsub(/ /, "", $2); print $2}'\'' .planning/phases/03-rules-conformance-matrix/03-RULES-MATRIX.md | sort | tr "\n" " "); \
  test "$RULES" = "$MATRIX"'
```

### Validation Reconciliation Check
```bash
bash -lc 'VALIDATION=.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md \
  && ! rg -n "^status: draft|⬜ pending|Approval: pending" "$VALIDATION" \
  && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" go test ./... -run" "$VALIDATION" \
  && rg -n "GOCACHE=\"\\$PWD/.cache/go-build\" bash scripts/verify-workspace\\.sh" "$VALIDATION"'
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Phase 3 summaries mark `VER-04` complete without a phase verification report | Milestone-grade `03-VERIFICATION.md` proves the shipped rules scope against current code | Needed by the 2026-03-14 audit | Restores audit traceability for Phase 3 |
| `03-VALIDATION.md` records intended commands with `draft` + pending rows | Validation record should reflect executed commands and completed statuses | Needed by the 2026-03-14 audit | Removes Phase 3 Nyquist partial-status debt |
| Bare `go test ./...` and `bash scripts/verify-workspace.sh` command text in Phase 3 validation | Copy-safe `GOCACHE="$PWD/.cache/go-build"` command text matching later verifier practice | Established by completed backfill work on 2026-03-14 | Keeps Phase 7 docs runnable in this environment |

**Deprecated/outdated:**
- `.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md` frontmatter `status: draft` for a shipped phase under active backfill closure.
- `.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md` task rows marked `⬜ pending` after fresh March 14, 2026 rules verification evidence already exists.
- `VER-04` pending/orphaned top-level traceability once `03-VERIFICATION.md` is backfilled.

## Open Questions

1. **Should Phase 7 update `REQUIREMENTS.md` in the same phase once `03-VERIFICATION.md` exists?**
   - What we know: `VER-04` is still unchecked and still mapped as pending in `.planning/REQUIREMENTS.md`.
   - What's unclear: Whether that reconciliation is considered mandatory Phase 7 work or deferred to a later final audit sweep.
   - Recommendation: Treat it as mandatory in the same phase, mirroring the Phase 6 `VER-03` closure pattern.

2. **What exact terminal frontmatter/status convention should `03-VALIDATION.md` use?**
   - What we know: Phase 2 now uses `status: completed`, while Phase 3 still says `status: draft`.
   - What's unclear: Whether the repository expects every backfilled validation record to adopt `completed` or another shared final token.
   - Recommendation: Reuse the Phase 2 convention unless a planner finds a stronger repo-wide standard.

3. **Should the fresh rules rerun be recorded as one combined command or as separate model/runtime commands?**
   - What we know: the combined command passes today, and the original validation file separates model and runtime rows.
   - What's unclear: Whether the summary evidence should emphasize one combined slice or preserve fully split command outputs.
   - Recommendation: Use the combined command for the re-verification summary, but keep separate model/runtime commands in `03-VALIDATION.md` so each original row remains directly runnable.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | `go test` plus shell-level artifact and traceability checks |
| Config file | `go.mod` |
| Quick run command | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance' -count=1` |
| Full suite command | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VER-04 | Checked-in rules matrix still reconciles one-to-one with `rules.md` and valid enforcement classes | artifact | `bash -lc 'RULES=$(sed -n '\''s/^- \(HSM[0-9][0-9]\):.*/\1/p'\'' rules.md | sort | tr "\n" " "); MATRIX=$(awk -F'\''|'\'' '\''/^\\| HSM[0-9][0-9] / {gsub(/ /, "", $2); print $2}'\'' .planning/phases/03-rules-conformance-matrix/03-RULES-MATRIX.md | sort | tr "\n" " "); test "$RULES" = "$MATRIX" && awk -F'\''|'\'' '\''/^\\| HSM[0-9][0-9] / {gsub(/ /, "", $3); if ($3 != "define_panic" && $3 != "runtime_semantic" && $3 != "exemplar") exit 1} END {exit 0}'\'' .planning/phases/03-rules-conformance-matrix/03-RULES-MATRIX.md'` | ✅ |
| VER-04 | Define-time `define_panic` rule coverage still passes on current code | unit | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance' -count=1` | ✅ |
| VER-04 | Runtime-semantic and exemplar rule coverage still passes on current code | unit | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeRulesConformance' -count=1` | ✅ |
| VER-04 | Canonical workspace gate still proves the rules suites in the broader shipped verification surface | gate | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` | ✅ |
| VER-04 | Milestone-grade evidence exists, validation debt is reconciled, and top-level traceability closes the orphan | artifact | `bash -lc 'test -f .planning/phases/03-rules-conformance-matrix/03-VERIFICATION.md && ! rg -n "^status: draft|⬜ pending|Approval: pending" .planning/phases/03-rules-conformance-matrix/03-VALIDATION.md && rg -n "^- \\[x\\] \\*\\*VER-04\\*\\*" .planning/REQUIREMENTS.md && ! rg -n "VER-04.*orphaned|03-VERIFICATION\\.md does not exist" .planning/v1.0-MILESTONE-AUDIT.md'` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** Run the smallest relevant Phase 3 targeted command plus any affected artifact check.
- **Per wave merge:** Run `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`.
- **Phase gate:** Matrix integrity, both conformance suites, artifact checks, and the canonical gate must all be green before `/gsd:verify-work`.

### Wave 0 Gaps
- [ ] `.planning/phases/03-rules-conformance-matrix/03-VERIFICATION.md` - missing required milestone artifact for `VER-04`
- [ ] `.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md` - still `status: draft`
- [ ] `.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md` - all five task rows still `⬜ pending`
- [ ] `.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md` - quick/full commands still omit `GOCACHE="$PWD/.cache/go-build"`
- [ ] `.planning/REQUIREMENTS.md` - `VER-04` remains unchecked and traceability still says `Pending`
- [ ] `.planning/v1.0-MILESTONE-AUDIT.md` - Phase 3 remains unverified and `VER-04` remains orphaned

## Sources

### Primary (HIGH confidence)
- `.planning/ROADMAP.md` - Phase 7 goal, success criteria, and dependency position
- `.planning/REQUIREMENTS.md` - current `VER-04` ownership and traceability state
- `.planning/v1.0-MILESTONE-AUDIT.md` - exact orphaned `VER-04` gap and Phase 3 validation debt
- `.planning/phases/03-rules-conformance-matrix/03-RESEARCH.md` - original Phase 3 implementation approach and rule-classification decisions
- `.planning/phases/03-rules-conformance-matrix/03-VALIDATION.md` - current draft/pending validation state and original task map
- `.planning/phases/03-rules-conformance-matrix/03-01-SUMMARY.md` - matrix/helper evidence
- `.planning/phases/03-rules-conformance-matrix/03-02-SUMMARY.md` - define-panic suite evidence
- `.planning/phases/03-rules-conformance-matrix/03-03-SUMMARY.md` - runtime/exemplar suite evidence
- `.planning/phases/03-rules-conformance-matrix/03-04-SUMMARY.md` - matrix closeout and canonical gate evidence
- `.planning/phases/03-rules-conformance-matrix/03-RULES-MATRIX.md` - rule-by-rule evidence mapping
- `rules.md` - canonical HSM01-HSM54 contract wording
- `hsm_rules_model_conformance_test.go` - shipped define-panic rule suite
- `hsm_rules_runtime_conformance_test.go` - shipped runtime/exemplar rule suite
- `scripts/verify-workspace.sh` - canonical workspace gate
- `.planning/phases/02-deterministic-runtime-semantics/02-VERIFICATION.md` - completed backfill verification template
- `.planning/phases/06-deterministic-runtime-verification-backfill/06-VALIDATION.md` - completed backfill validation template
- Fresh command evidence from 2026-03-14:
  - `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance' -count=1`
  - `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
  - bare `go test ./... -run 'TestModelRulesConformance|TestRuntimeRulesConformance' -count=1` failure due default build-cache permissions
  - matrix integrity check confirming `rules.md` and `03-RULES-MATRIX.md` still match

### Secondary (MEDIUM confidence)
- `$HOME/.codex/get-shit-done/workflows/audit-milestone.md` - confirms the audit consumes phase `VERIFICATION.md` files, requirements traceability, summaries, and Nyquist validation state

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - entirely repo-local and verified against current commands and artifacts
- Architecture: HIGH - directly derived from completed Phase 6 backfill and current Phase 3 audit gap
- Pitfalls: HIGH - grounded in current draft/orphaned files plus a fresh bare-command failure

**Research date:** 2026-03-14
**Valid until:** 2026-03-21
