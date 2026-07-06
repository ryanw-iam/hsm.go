package hsm_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

func TestRuntimeAdversarial(t *testing.T) {
	t.Run("custom_queue_push_error_dispatches_error_event_without_reentering_queue", func(t *testing.T) {
		queueErr := errors.New("queue push failed")
		pushes := 0
		model := hsm.Define(
			"QueuePushErrorHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
			hsm.State("failed"),
			hsm.Transition(
				hsm.On(hsm.ErrorEvent),
				hsm.Source("idle"),
				hsm.Target("failed"),
			),
		)
		queue := hsm.Queue{
			Push: func(context.Context, hsm.Event) error {
				pushes++
				return queueErr
			},
			Pop: func(context.Context) (hsm.Event, bool, error) {
				return hsm.Event{}, false, nil
			},
			Len: func(context.Context) (int, error) {
				return 0, nil
			},
		}

		sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Queue: queue})
		awaitWaiter(t, "queue push failure dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: "go"}))

		if pushes != 1 {
			t.Fatalf("expected one failed push, got %d", pushes)
		}
		if sm.State() != "/QueuePushErrorHSM/failed" {
			t.Fatalf("expected failed after queue push failure, got %s", sm.State())
		}
	})

	t.Run("snapshot_len_error_does_not_mutate_queue", func(t *testing.T) {
		queueErr := errors.New("queue len failed")
		pushes := 0
		model := hsm.Define(
			"SnapshotQueueLenErrorHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
		)
		queue := hsm.Queue{
			Push: func(context.Context, hsm.Event) error {
				pushes++
				return nil
			},
			Pop: func(context.Context) (hsm.Event, bool, error) {
				return hsm.Event{}, false, nil
			},
			Len: func(context.Context) (int, error) {
				return 0, queueErr
			},
		}

		sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Queue: queue})
		snapshot := hsm.TakeSnapshot(context.Background(), sm)

		if snapshot.QueueLen != 0 {
			t.Fatalf("snapshot QueueLen = %d, want 0 after Len error", snapshot.QueueLen)
		}
		if pushes != 0 {
			t.Fatalf("snapshot pushed %d events after Len error", pushes)
		}
	})

	t.Run("concurrent_behavior_panics_dispatch_error_event_and_stop_settles", func(t *testing.T) {
		recorder := newErrorEventRecorder()
		activityStarted := make(chan struct{})
		releasePanic := make(chan struct{})

		model := hsm.Define(
			"ConcurrentBehaviorPanicAdversarialHSM",
			hsm.Initial(hsm.Target("running")),
			hsm.State("running",
				hsm.Activity(func(ctx context.Context, sm *THSM, event hsm.Event) {
					close(activityStarted)
					<-releasePanic
					panic("concurrent boom")
				}),
			),
			hsm.State("failed"),
			hsm.Transition(
				hsm.On(hsm.ErrorEvent),
				hsm.Source("running"),
				hsm.Target("failed"),
				hsm.Effect(recorder.effect),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "concurrent panic activity start", activityStarted)

		errorProcessed := hsm.AfterProcess(sm.Context(), sm, hsm.ErrorEvent)
		failedEntered := hsm.AfterEntry(sm.Context(), sm, "/ConcurrentBehaviorPanicAdversarialHSM/failed")

		close(releasePanic)
		awaitWaiter(t, "concurrent panic error processing", errorProcessed)
		awaitWaiter(t, "concurrent panic failed entry", failedEntered)

		err := recorder.await(t, "concurrent behavior panic")
		if !strings.Contains(err.Error(), "panic in concurrent behavior /ConcurrentBehaviorPanicAdversarialHSM/running/") {
			t.Fatalf("expected concurrent panic error path, got %v", err)
		}
		if !strings.Contains(err.Error(), "concurrent boom") {
			t.Fatalf("expected concurrent panic message, got %v", err)
		}
		if sm.State() != "/ConcurrentBehaviorPanicAdversarialHSM/failed" {
			t.Fatalf("expected failed state after concurrent panic, got %s", sm.State())
		}

		awaitWaiter(t, "concurrent panic stop", hsm.Stop(context.Background(), sm))
	})

	t.Run("synchronous_behavior_panic_error_event_drains_before_dispatch_completion", func(t *testing.T) {
		recorder := newErrorEventRecorder()
		boomEvent := hsm.Event{Name: "boom"}

		model := hsm.Define(
			"SynchronousBehaviorPanicAdversarialHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(
					hsm.On(boomEvent),
					hsm.Target("../unreachable"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						panic("sync boom")
					}),
				),
			),
			hsm.State("failed"),
			hsm.State("unreachable"),
			hsm.Transition(
				hsm.On(hsm.ErrorEvent),
				hsm.Source("idle"),
				hsm.Target("failed"),
				hsm.Effect(recorder.effect),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "synchronous panic dispatch", hsm.Dispatch(sm.Context(), sm, boomEvent))

		err := recorder.await(t, "synchronous behavior panic")
		if !strings.Contains(err.Error(), "panic in behavior /SynchronousBehaviorPanicAdversarialHSM/idle/") {
			t.Fatalf("expected synchronous panic error path, got %v", err)
		}
		if !strings.Contains(err.Error(), "sync boom") {
			t.Fatalf("expected synchronous panic message, got %v", err)
		}
		if sm.State() != "/SynchronousBehaviorPanicAdversarialHSM/failed" {
			t.Fatalf("expected failed state after synchronous panic, got %s", sm.State())
		}
	})

	t.Run("processing_panic_dispatches_error_event_and_machine_remains_live", func(t *testing.T) {
		recorder := newErrorEventRecorder()
		boomEvent := hsm.Event{Name: "boom"}
		recoverEvent := hsm.Event{Name: "recover"}

		model := hsm.Define(
			"ProcessingPanicAdversarialHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(
					hsm.On(boomEvent),
					hsm.Target("../unreachable"),
					hsm.Guard(func(ctx context.Context, sm *THSM, event hsm.Event) bool {
						panic("processing boom")
					}),
				),
			),
			hsm.State("failed",
				hsm.Transition(
					hsm.On(recoverEvent),
					hsm.Target("../recovered"),
				),
			),
			hsm.State("recovered"),
			hsm.State("unreachable"),
			hsm.Transition(
				hsm.On(hsm.ErrorEvent),
				hsm.Source("idle"),
				hsm.Target("failed"),
				hsm.Effect(recorder.effect),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		errorProcessed := hsm.AfterProcess(sm.Context(), sm, hsm.ErrorEvent)
		failedEntered := hsm.AfterEntry(sm.Context(), sm, "/ProcessingPanicAdversarialHSM/failed")

		awaitWaiter(t, "processing panic dispatch", hsm.Dispatch(sm.Context(), sm, boomEvent))
		awaitWaiter(t, "processing panic error processing", errorProcessed)
		awaitWaiter(t, "processing panic failed entry", failedEntered)

		err := recorder.await(t, "processing panic")
		if !strings.Contains(err.Error(), "panic in guard /ProcessingPanicAdversarialHSM/idle/") {
			t.Fatalf("expected processing panic error, got %v", err)
		}
		if sm.State() != "/ProcessingPanicAdversarialHSM/failed" {
			t.Fatalf("expected failed state after processing panic, got %s", sm.State())
		}

		recoveredEntered := hsm.AfterEntry(sm.Context(), sm, "/ProcessingPanicAdversarialHSM/recovered")
		awaitWaiter(t, "processing panic recovery dispatch", hsm.Dispatch(sm.Context(), sm, recoverEvent))
		awaitWaiter(t, "processing panic recovered entry", recoveredEntered)

		if sm.State() != "/ProcessingPanicAdversarialHSM/recovered" {
			t.Fatalf("expected recovered state after post-failure dispatch, got %s", sm.State())
		}
	})

	t.Run("operation_panic_returns_error_without_oncall_dispatch", func(t *testing.T) {
		type OperationPanicHSM struct {
			hsm.HSM
		}
		model := hsm.Define(
			"OperationPanicAdversarialHSM",
			hsm.Operation("explode", func(context.Context, *OperationPanicHSM) any {
				panic("operation boom")
			}),
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(hsm.OnCall("explode"), hsm.Target("../called")),
			),
			hsm.State("called"),
		)
		sm := hsm.Started(context.Background(), &OperationPanicHSM{}, &model)

		result, err := hsm.Call(context.Background(), sm, "explode")
		if err == nil || !strings.Contains(err.Error(), "operation /OperationPanicAdversarialHSM/explode panic: operation boom") {
			t.Fatalf("operation panic error = %v", err)
		}
		if result != nil {
			t.Fatalf("operation panic result = %#v, want nil", result)
		}
		if sm.State() != "/OperationPanicAdversarialHSM/idle" {
			t.Fatalf("operation panic dispatched OnCall, state = %s", sm.State())
		}
	})

	t.Run("stop_timeout_dispatches_error_event_with_injected_clock", func(t *testing.T) {
		harness := newDeterministicClockHarness()
		recorder := newErrorEventRecorder()
		activityStarted := make(chan struct{})
		releaseActivity := make(chan struct{})

		t.Cleanup(func() {
			close(releaseActivity)
		})

		model := hsm.Define(
			"TerminateTimeoutAdversarialHSM",
			hsm.Initial(hsm.Target("running")),
			hsm.State("running",
				hsm.Activity(func(ctx context.Context, sm *THSM, event hsm.Event) {
					close(activityStarted)
					<-releaseActivity
				}),
			),
			hsm.Transition(
				hsm.On(hsm.ErrorEvent),
				hsm.Effect(recorder.effect),
			),
		)

		sm := hsm.Started(
			context.Background(),
			&THSM{},
			&model,
			hsm.Config{
				Clock:           harness.Clock(),
				ActivityTimeout: 5 * time.Second,
			},
		)
		awaitWaiter(t, "terminate timeout activity start", activityStarted)

		errorProcessed := hsm.AfterProcess(sm.Context(), sm, hsm.ErrorEvent)
		stopCompleted := hsm.Stop(context.Background(), sm)

		registration := harness.awaitRegistration(t, "terminate timeout adversarial")
		if registration.kind != "after" {
			t.Fatalf("expected terminate timeout to use clock.After, got %s", registration.kind)
		}
		if registration.requested != 5*time.Second {
			t.Fatalf("expected terminate timeout duration %v, got %v", 5*time.Second, registration.requested)
		}
		assertWaiterPending(t, "terminate timeout stop before injected timeout", stopCompleted)

		registration.trigger(t)
		awaitWaiter(t, "terminate timeout stop completion", stopCompleted)
		awaitWaiter(t, "terminate timeout error processing", errorProcessed)

		err := recorder.await(t, "terminate timeout")
		if !strings.Contains(err.Error(), "terminate timeout: /TerminateTimeoutAdversarialHSM/running/") {
			t.Fatalf("expected terminate timeout error, got %v", err)
		}
		select {
		case <-sm.Context().Done():
		default:
			t.Fatal("expected machine context to be done after terminate timeout stop")
		}
	})
}

func TestPublicHelperAdversarial(t *testing.T) {
	nilCtx := context.Context(nil)
	hostileEvent := hsm.Event{Name: "hostile"}

	t.Run("nil_instance_dispatch_and_set_close_immediately", func(t *testing.T) {
		assertCompletionErr(t, "dispatch with nil instance", hsm.Dispatch(context.Background(), nil, hostileEvent), hsm.ErrMissingHSM)
		assertCompletionErr(t, "set with nil instance", hsm.Set(context.Background(), nil, "hostile", 1), hsm.ErrMissingHSM)
	})

	t.Run("missing_context_call_returns_missing_hsm", func(t *testing.T) {
		_, err := hsm.Call(context.Background(), nil, "missing")
		if !errors.Is(err, hsm.ErrMissingHSM) {
			t.Fatalf("expected ErrMissingHSM for missing context call, got %v", err)
		}
	})

	t.Run("typed_nil_context_helpers_do_not_panic", func(t *testing.T) {
		assertNoPanic(t, "dispatch with typed nil context", func() {
			assertCompletionErr(t, "dispatch with typed nil context", hsm.Dispatch(nilCtx, nil, hostileEvent), hsm.ErrMissingHSM)
		})
		assertNoPanic(t, "set with typed nil context", func() {
			assertCompletionErr(t, "set with typed nil context", hsm.Set(nilCtx, nil, "hostile", 1), hsm.ErrMissingHSM)
		})
		assertNoPanic(t, "dispatch all with typed nil context", func() {
			assertWaiterClosed(t, "dispatch all with typed nil context", hsm.DispatchAll(nilCtx, hostileEvent))
		})
		assertNoPanic(t, "dispatch to with typed nil context", func() {
			assertWaiterClosed(t, "dispatch to with typed nil context", hsm.DispatchTo[string](nilCtx, hostileEvent, "alpha"))
		})
		assertNoPanic(t, "call with typed nil context", func() {
			_, err := hsm.Call(nilCtx, nil, "missing")
			if !errors.Is(err, hsm.ErrMissingHSM) {
				t.Fatalf("expected ErrMissingHSM for typed nil context call, got %v", err)
			}
		})
	})

	t.Run("nil_stop_closes_immediately_without_panicking", func(t *testing.T) {
		assertNoPanic(t, "Stop nil hsm", func() {
			assertWaiterClosed(t, "stop with nil instance", hsm.Stop(context.Background(), nil))
		})
	})
}
