# Phase 2 Research: Deterministic Runtime Semantics

## Objective

Determine how to turn the `hsm` runtime's concurrency, timer, cancellation, and dispatch-ordering behavior into deterministic, repeatable tests that can satisfy `VER-03` without relying on scheduler luck or arbitrary sleeps.

## Current State

- Phase 1 established a blocking workspace gate and stabilized two obviously flaky root tests (`TestEvery`, `TestWhen`), but the runtime suite still mixes deterministic waiter-based checks with many sleep-based assertions.
- The runtime already exposes the main seams needed for deterministic testing:
  - injectable clocks via `Config.Clock` and `DefaultClock`
  - processing waiters via `AfterProcess`, `AfterDispatch`, `AfterEntry`, `AfterExit`, and `AfterExecuted`
  - explicit lifecycle APIs via `Dispatch`, `DispatchAll`, `DispatchTo`, `Restart`, `Stop`, `TakeSnapshot`, `FromContext`, and `InstancesFromContext`
- Critical runtime behavior lives in a few concentrated areas:
  - timer activities: `After`, `At`, `Every`, `When`
  - queue and processing coordination: `queue`, `mutex`, `dispatch`, `process`, `processEvent`
  - lifecycle termination: `stop`, `restart`, `terminate`
  - group/context fan-out: `DispatchAll`, `DispatchTo`, `NewGroup`

## Key Findings

### 1. Deterministic testing should target runtime seams, not just public APIs

The runtime has enough instrumentation to test deterministic behavior directly, but the current suite does not organize around those seams. Planning should slice work by runtime mechanism:

- queue ordering and processing serialization
- timer-driven transition semantics
- cancellation and termination behavior
- observer/waiter synchronization semantics
- multi-instance broadcast and targeted dispatch behavior

This is a better Phase 2 split than grouping by file alone.

### 2. Injected clocks are the main replacement for `time.Sleep`

`Clock` already supports `After` and `NewTimer`, and `hsm.start()` installs a fully defaulted clock on each instance. That means Phase 2 can build a deterministic fake clock/timer harness without changing the public API.

Implication:

- `After`, `At`, and `Every` should be tested through injected timers and explicit trigger control.
- Tests should avoid waiting on real elapsed wall time except as last-resort safety deadlines.

### 3. Waiter channels are already the correct synchronization primitive

The runtime closes channels through `AfterDispatch`, `AfterProcess`, `AfterEntry`, `AfterExit`, and `AfterExecuted`, and `Dispatch` returns `sm.processing.wait()`. Phase 2 should standardize on those primitives as the assertion boundary.

Implication:

- Tests should assert on state only after the corresponding waiter closes.
- Post-dispatch snapshots should be taken after settled processing, not after a sleep interval.

### 4. There are still risky nondeterministic surfaces that need dedicated coverage

Several behaviors are likely to hide bugs or flakes if tested only indirectly:

- `DispatchAll` and `DispatchTo` iterate through a `sync.Map`, so arrival order across instances is intentionally unstable and should not be asserted by map iteration order.
- `process` and `queue` blend FIFO with LIFO completion events, which needs explicit ordering tests.
- `terminate` uses `clock.After(activity timeout)` and emits `ErrorEvent` on timeout, which is exactly the kind of edge path that Phase 2 should make deterministic.
- `When` repeatedly selects on the returned channel and will keep dispatching if the expression returns a permanently-closed channel. That is a behavior edge worth explicitly covering, even if remediation falls later.

### 5. Phase 2 should create reusable deterministic harness helpers first

Trying to write each runtime test with its own ad hoc clock, timer, and waiter plumbing will recreate the same problem Phase 1 just reduced in root coverage structure.

Planning should start with a harness layer, then use it to cover:

- timer advancement
- concurrent dispatch coordination
- multi-instance broadcast targeting
- cancellation and termination observation

### 6. The release gate should keep using the canonical Phase 1 script

Phase 2 does not need a second release entrypoint. The right shape is:

- task-level quick checks on the root runtime suite while developing deterministic harnesses
- wave-level/full verification through `bash scripts/verify-workspace.sh`

That preserves one shipping gate while allowing faster feedback during implementation.

## Recommended Plan Shape

### Plan slice 1: deterministic timer and waiter harness

Build shared test helpers for fake timers/clocks and waiter-based synchronization. This should cover `After`, `At`, `Every`, `When`, and observer signaling without sleeps.

### Plan slice 2: dispatch ordering and multi-instance semantics

Cover queue ordering, processing serialization, `DispatchAll`, `DispatchTo`, and cross-instance targeting. Assertions should prove stable outcomes without assuming `sync.Map` iteration order.

### Plan slice 3: termination, cancellation, and race-focused hardening

Cover `Stop`, `Restart`, activity termination timeouts, and race-enabled verification on the newly deterministic runtime suites. This is where the phase should close the `VER-03` race/determinism contract.

## Validation Architecture

Phase 2 validation should be layered:

- Quick command: `go test ./...`
- Full suite command: `bash scripts/verify-workspace.sh`

The root runtime package is where Phase 2 work lands, so quick feedback can stay focused there, while wave completion and phase completion should still use the canonical workspace gate introduced in Phase 1.

## Risks And Watchouts

- Do not assert ordering based on goroutine scheduling or `sync.Map` iteration order.
- Do not use raw `time.Sleep` for causal assertions when a waiter or fake timer can express the same behavior.
- `AfterExecuted` only fires for concurrent behavior completion; plans should avoid assuming it applies to every state transition path.
- `NewGroup` inspects instance started-state, so deterministic tests must use real started instances or nil/empty groups rather than unstarted embedded structs.
- Phase 2 will likely expose real runtime defects; those tests belong here, but broad remediation beyond the scoped runtime semantics should still roll forward into Phase 5 unless necessary to make the phase pass.

## Recommendation

Create Phase 2 plans that:

1. Introduce reusable deterministic timer/waiter harnesses.
2. Convert runtime timing and dispatch tests away from sleep-driven assertions.
3. Add explicit concurrency, broadcast, and cancellation semantics coverage.
4. Keep `bash scripts/verify-workspace.sh` as the blocking full-suite command while using faster root-package checks for task feedback.
