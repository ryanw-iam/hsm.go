---
phase: 5
slug: hardening-remediation-closure
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-13
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` |
| **Config file** | `go.mod` |
| **Quick run command** | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|Test.*RegressionAllocs'` |
| **Full suite command** | `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial|Test.*RegressionAllocs'`
- **After every plan wave:** Run `GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | VER-06 | unit | `GOCACHE="$PWD/.cache/go-build" go test ./... -run 'TestPublicHelperAdversarial|TestRuntimeStopCancelsActivityAndClosesAfterExecuted|TestRuntimeStopTimeoutDispatchesErrorEventDeterministically'` | ✅ | ⬜ pending |
| 05-02-01 | 02 | 2 | VER-06 | unit+docs | `bash -lc 'rg -n "HSM53: ALWAYS wait on the channel returned by \`Dispatch\`, \`Set\`, \`Restart\`, \`Stop\`, \`DispatchAll\`, or \`DispatchTo\` before asserting on post-transition state\\." rules.md >/dev/null && rg -n "HSM54: NEVER use \`hsm.AfterProcess\\(\\.\\.\\.\\)\`, \`hsm.AfterDispatch\\(\\.\\.\\.\\)\`, \`hsm.AfterEntry\\(\\.\\.\\.\\)\`, \`hsm.AfterExit\\(\\.\\.\\.\\)\`, or \`hsm.AfterExecuted\\(\\.\\.\\.\\)\` as production synchronization mechanisms\\." rules.md >/dev/null && for f in hsm.go README.md; do rg -n "completion channel|supported production synchronization path" "$f" >/dev/null && rg -n "tests and deterministic observation only|deterministic observation" "$f" >/dev/null; done && ! rg -n "coordinating external operations with state transitions|general synchronization" hsm.go README.md >/dev/null && git ls-files --error-unmatch rules.md >/dev/null && GOCACHE="$PWD/.cache/go-build" go test ./... -run "TestRuntimeRulesConformance|TestPublicHelperAdversarial"'` | ✅ | ⬜ pending |
| 05-03-01 | 03 | 3 | VER-06 | smoke+hygiene | `bash -lc '! rg -n "phase 4 workspace verification|phase 4" scripts/verify-workspace.sh >/dev/null && rg -n "canonical workspace verification|workspace release gate|workspace verification passed|release gate passed" scripts/verify-workspace.sh >/dev/null && GOCACHE="$PWD/.cache/go-build" bash scripts/verify-workspace.sh && test ! -f .travis.yml'` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None | n/a | All Phase 5 checks are covered by automated verification commands. | n/a |

---

## Validation Sign-Off

- [x] All tasks have automated verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all missing references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
