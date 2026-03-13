package hsm_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

func TestRuntimeStopCancelsActivityAndClosesAfterExecuted(t *testing.T) {
	activityStarted := make(chan struct{})
	activityCancelled := make(chan struct{})

	model := hsm.Define(
		"RuntimeStopActivityHSM",
		hsm.Initial(hsm.Target("running")),
		hsm.State("running",
			hsm.Activity(func(ctx context.Context, sm *THSM, event hsm.Event) {
				close(activityStarted)
				<-ctx.Done()
				close(activityCancelled)
			}),
		),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	awaitWaiter(t, "RuntimeStopActivityHSM activity start", activityStarted)

	executed := hsm.AfterExecuted(sm.Context(), sm, "/RuntimeStopActivityHSM/running")
	enteredAgain := hsm.AfterEntry(sm.Context(), sm, "/RuntimeStopActivityHSM/running")

	awaitWaiter(t, "RuntimeStopActivityHSM stop", hsm.Stop(context.Background(), sm))
	awaitWaiter(t, "RuntimeStopActivityHSM activity cancellation", activityCancelled)
	awaitWaiter(t, "RuntimeStopActivityHSM AfterExecuted waiter", executed)
	assertWaiterPending(t, "RuntimeStopActivityHSM re-entry after stop", enteredAgain)
}

func TestRuntimeRestartReentersInitialStateWithData(t *testing.T) {
	advanceEvent := hsm.Event{Name: "advance"}
	initialData := make(chan any, 2)

	model := hsm.Define(
		"RuntimeRestartLifecycleHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Entry(func(ctx context.Context, sm *THSM, event hsm.Event) {
				initialData <- event.Data
			}),
		),
		hsm.Transition(hsm.On(advanceEvent), hsm.Source("idle"), hsm.Target("done")),
		hsm.State("done"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Data: "boot"})

	select {
	case got := <-initialData:
		if got != "boot" {
			t.Fatalf("expected initial start data %q, got %#v", "boot", got)
		}
	case <-time.After(waiterDeadline):
		t.Fatal("timed out waiting for initial start data")
	}

	awaitWaiter(t, "RuntimeRestartLifecycleHSM advance dispatch", hsm.Dispatch(sm.Context(), sm, advanceEvent))
	if sm.State() != "/RuntimeRestartLifecycleHSM/done" {
		t.Fatalf("expected state to be done before restart, got %s", sm.State())
	}

	reenteredIdle := hsm.AfterEntry(sm.Context(), sm, "/RuntimeRestartLifecycleHSM/idle")
	exitedDone := hsm.AfterExit(sm.Context(), sm, "/RuntimeRestartLifecycleHSM/done")
	awaitWaiter(t, "RuntimeRestartLifecycleHSM restart", hsm.Restart(context.Background(), sm, "again"))
	awaitWaiter(t, "RuntimeRestartLifecycleHSM done exit on restart", exitedDone)
	awaitWaiter(t, "RuntimeRestartLifecycleHSM idle re-entry on restart", reenteredIdle)

	if sm.State() != "/RuntimeRestartLifecycleHSM/idle" {
		t.Fatalf("expected state to reset to idle after restart, got %s", sm.State())
	}

	select {
	case got := <-initialData:
		if got != "again" {
			t.Fatalf("expected restart data %q, got %#v", "again", got)
		}
	case <-time.After(waiterDeadline):
		t.Fatal("timed out waiting for restart data")
	}
}

func TestRuntimeStopTimeoutDispatchesErrorEventDeterministically(t *testing.T) {
	harness := newDeterministicClockHarness()
	activityStarted := make(chan struct{})
	releaseActivity := make(chan struct{})
	errorsSeen := make(chan error, 1)

	t.Cleanup(func() {
		close(releaseActivity)
	})

	model := hsm.Define(
		"RuntimeStopTimeoutHSM",
		hsm.Initial(hsm.Target("running")),
		hsm.State("running",
			hsm.Activity(func(ctx context.Context, sm *THSM, event hsm.Event) {
				close(activityStarted)
				<-releaseActivity
			}),
		),
		hsm.Transition(
			hsm.On(hsm.ErrorEvent),
			hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
				err, ok := event.Data.(error)
				if !ok {
					return
				}
				select {
				case errorsSeen <- err:
				default:
				}
			}),
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
	awaitWaiter(t, "RuntimeStopTimeoutHSM activity start", activityStarted)

	errorProcessed := hsm.AfterProcess(sm.Context(), sm, hsm.ErrorEvent)
	stopCompleted := hsm.Stop(context.Background(), sm)

	registration := harness.awaitRegistration(t, "RuntimeStopTimeoutHSM termination timeout")
	if registration.kind != "after" {
		t.Fatalf("expected terminate timeout to use clock.After, got %s", registration.kind)
	}
	if registration.requested != 5*time.Second {
		t.Fatalf("expected terminate timeout duration %v, got %v", 5*time.Second, registration.requested)
	}

	assertWaiterPending(t, "RuntimeStopTimeoutHSM stop before timeout", stopCompleted)
	registration.trigger(t)
	awaitWaiter(t, "RuntimeStopTimeoutHSM stop after timeout", stopCompleted)
	awaitWaiter(t, "RuntimeStopTimeoutHSM error processing", errorProcessed)

	select {
	case err := <-errorsSeen:
		if !strings.Contains(err.Error(), "terminate timeout: /RuntimeStopTimeoutHSM/running/") {
			t.Fatalf("expected terminate timeout error for running activity, got %v", err)
		}
	case <-time.After(waiterDeadline):
		t.Fatal("timed out waiting for termination timeout error")
	}
}
