package hsm_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

func TestRuntimeAfterUsesDeterministicTimer(t *testing.T) {
	harness := newDeterministicClockHarness()
	model := hsm.Define(
		"RuntimeAfterHSM",
		hsm.Initial(hsm.Target("foo")),
		hsm.State("foo",
			hsm.Transition(
				hsm.After(func(ctx context.Context, sm *THSM, event hsm.Event) time.Duration {
					return 5 * time.Minute
				}),
				hsm.Target("../bar"),
			),
		),
		hsm.State("bar"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Clock: harness.Clock()})
	if sm.State() != "/RuntimeAfterHSM/foo" {
		t.Fatalf("expected state to be foo, got %s", sm.State())
	}

	registration := harness.awaitRegistration(t, "After transition")
	if registration.kind != "timer" {
		t.Fatalf("expected timer registration, got %s", registration.kind)
	}
	if registration.requested != 5*time.Minute {
		t.Fatalf("expected After duration %v, got %v", 5*time.Minute, registration.requested)
	}

	entered := hsm.AfterEntry(sm.Context(), sm, "/RuntimeAfterHSM/bar")
	registration.trigger(t)
	awaitWaiter(t, "RuntimeAfterHSM entering bar", entered)

	if sm.State() != "/RuntimeAfterHSM/bar" {
		t.Fatalf("expected state to be bar, got %s", sm.State())
	}
}

func TestRuntimeAfterNegativeDurationDoesNotSchedule(t *testing.T) {
	harness := newDeterministicClockHarness()
	model := hsm.Define(
		"RuntimeAfterNegativeHSM",
		hsm.Initial(hsm.Target("foo")),
		hsm.State("foo",
			hsm.Transition(
				hsm.After(func(ctx context.Context, sm *THSM, event hsm.Event) time.Duration {
					return -time.Minute
				}),
				hsm.Target("../bar"),
			),
		),
		hsm.State("bar"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Clock: harness.Clock()})
	entered := hsm.AfterEntry(sm.Context(), sm, "/RuntimeAfterNegativeHSM/bar")

	harness.assertNoRegistration(t, "negative After duration")
	assertWaiterPending(t, "RuntimeAfterNegativeHSM entering bar", entered)

	if harness.registrationCount() != 0 {
		t.Fatalf("expected no timer registrations, got %d", harness.registrationCount())
	}
	if sm.State() != "/RuntimeAfterNegativeHSM/foo" {
		t.Fatalf("expected state to remain foo, got %s", sm.State())
	}
}

func TestRuntimeAtUsesDeterministicTimer(t *testing.T) {
	harness := newDeterministicClockHarness()
	targetTime := time.Now().Add(2 * time.Hour)
	model := hsm.Define(
		"RuntimeAtHSM",
		hsm.Initial(hsm.Target("foo")),
		hsm.State("foo",
			hsm.Transition(
				hsm.At(func(ctx context.Context, sm *THSM, event hsm.Event) time.Time {
					return targetTime
				}),
				hsm.Target("../bar"),
			),
		),
		hsm.State("bar"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Clock: harness.Clock()})
	registration := harness.awaitRegistration(t, "At transition")
	if registration.kind != "timer" {
		t.Fatalf("expected timer registration, got %s", registration.kind)
	}
	if registration.requested < 90*time.Minute {
		t.Fatalf("expected At duration to stay well in the future, got %v", registration.requested)
	}

	entered := hsm.AfterEntry(sm.Context(), sm, "/RuntimeAtHSM/bar")
	registration.trigger(t)
	awaitWaiter(t, "RuntimeAtHSM entering bar", entered)

	if sm.State() != "/RuntimeAtHSM/bar" {
		t.Fatalf("expected state to be bar, got %s", sm.State())
	}
}

func TestRuntimeEveryUsesManualTriggers(t *testing.T) {
	harness := newDeterministicClockHarness()
	var ticks atomic.Int32
	tickProcessed := make(chan int32, 3)
	model := hsm.Define(
		"RuntimeEveryHSM",
		hsm.Initial(hsm.Target("foo")),
		hsm.State("foo",
			hsm.Transition(
				hsm.Every(func(ctx context.Context, sm *THSM, event hsm.Event) time.Duration {
					return 10 * time.Minute
				}),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					tickProcessed <- ticks.Add(1)
				}),
			),
		),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Clock: harness.Clock()})
	if sm.State() != "/RuntimeEveryHSM/foo" {
		t.Fatalf("expected state to be foo, got %s", sm.State())
	}

	registration := harness.awaitRegistration(t, "Every transition")
	if registration.requested != 10*time.Minute {
		t.Fatalf("expected Every duration %v, got %v", 10*time.Minute, registration.requested)
	}

	for expected := int32(1); expected <= 3; expected++ {
		registration.trigger(t)
		select {
		case got := <-tickProcessed:
			if got != expected {
				t.Fatalf("expected tick %d, got %d", expected, got)
			}
		case <-time.After(waiterDeadline):
			t.Fatalf("timed out waiting for tick %d", expected)
		}
	}

	if ticks.Load() != 3 {
		t.Fatalf("expected 3 Every ticks, got %d", ticks.Load())
	}
}

func TestRuntimeWhenWaitsForExplicitSignal(t *testing.T) {
	ready := make(chan struct{})
	model := hsm.Define(
		"RuntimeWhenHSM",
		hsm.Initial(hsm.Target("foo")),
		hsm.State("foo",
			hsm.Transition(
				hsm.When(func(ctx context.Context, sm *THSM, event hsm.Event) <-chan struct{} {
					return ready
				}),
				hsm.Target("../bar"),
			),
		),
		hsm.State("bar"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	entered := hsm.AfterEntry(sm.Context(), sm, "/RuntimeWhenHSM/bar")
	assertWaiterPending(t, "RuntimeWhenHSM entering bar before signal", entered)

	close(ready)
	awaitWaiter(t, "RuntimeWhenHSM entering bar", entered)

	if sm.State() != "/RuntimeWhenHSM/bar" {
		t.Fatalf("expected state to be bar, got %s", sm.State())
	}
}

func TestRuntimeConfigClockOverridesDefaultClock(t *testing.T) {
	defaultHarness := newDeterministicClockHarness()
	configHarness := newDeterministicClockHarness()
	originalClock := hsm.DefaultClock
	hsm.DefaultClock = defaultHarness.Clock()
	t.Cleanup(func() {
		hsm.DefaultClock = originalClock
	})

	model := hsm.Define(
		"RuntimeConfigClockHSM",
		hsm.Initial(hsm.Target("foo")),
		hsm.State("foo",
			hsm.Transition(
				hsm.After(func(ctx context.Context, sm *THSM, event hsm.Event) time.Duration {
					return time.Minute
				}),
				hsm.Target("../bar"),
			),
		),
		hsm.State("bar"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Clock: configHarness.Clock()})
	registration := configHarness.awaitRegistration(t, "config clock transition")
	defaultHarness.assertNoRegistration(t, "default clock fallback while config clock is set")

	entered := hsm.AfterEntry(sm.Context(), sm, "/RuntimeConfigClockHSM/bar")
	registration.trigger(t)
	awaitWaiter(t, "RuntimeConfigClockHSM entering bar", entered)

	if sm.State() != "/RuntimeConfigClockHSM/bar" {
		t.Fatalf("expected state to be bar, got %s", sm.State())
	}
}

func TestRuntimeDefaultClockFallbackUsesDeterministicTimer(t *testing.T) {
	defaultHarness := newDeterministicClockHarness()
	originalClock := hsm.DefaultClock
	hsm.DefaultClock = defaultHarness.Clock()
	t.Cleanup(func() {
		hsm.DefaultClock = originalClock
	})

	model := hsm.Define(
		"RuntimeDefaultClockHSM",
		hsm.Initial(hsm.Target("foo")),
		hsm.State("foo",
			hsm.Transition(
				hsm.After(func(ctx context.Context, sm *THSM, event hsm.Event) time.Duration {
					return 3 * time.Minute
				}),
				hsm.Target("../bar"),
			),
		),
		hsm.State("bar"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	registration := defaultHarness.awaitRegistration(t, "default clock transition")
	if registration.requested != 3*time.Minute {
		t.Fatalf("expected default clock duration %v, got %v", 3*time.Minute, registration.requested)
	}

	entered := hsm.AfterEntry(sm.Context(), sm, "/RuntimeDefaultClockHSM/bar")
	registration.trigger(t)
	awaitWaiter(t, "RuntimeDefaultClockHSM entering bar", entered)

	if sm.State() != "/RuntimeDefaultClockHSM/bar" {
		t.Fatalf("expected state to be bar, got %s", sm.State())
	}
}
