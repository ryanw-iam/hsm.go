package hsm_test

import (
	"context"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

func TestWhenStringAliasesOnSet(t *testing.T) {
	model := hsm.Define(
		"CanonicalWhenStringHSM",
		hsm.Attribute("flag", false),
		hsm.Initial(hsm.Target("idle")),
		hsm.State(
			"idle",
			hsm.Transition(hsm.When("flag"), hsm.Target("../changed")),
		),
		hsm.State("changed"),
	)
	sm := hsm.Started(context.Background(), &THSM{}, &model)

	<-hsm.Set(context.Background(), sm, "flag", true)
	if got := sm.State(); got != "/CanonicalWhenStringHSM/changed" {
		t.Fatalf("When string transition state = %q, want changed", got)
	}
}

func TestTimerStringAttributesDriveTransitions(t *testing.T) {
	t.Run("After", func(t *testing.T) {
		harness := newDeterministicClockHarness()
		model := hsm.Define(
			"CanonicalAfterAttributeHSM",
			hsm.Attribute("delay", 3*time.Minute),
			hsm.Initial(hsm.Target("waiting")),
			hsm.State(
				"waiting",
				hsm.Transition(hsm.After("delay"), hsm.Target("../done")),
			),
			hsm.State("done"),
		)
		sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Clock: harness.Clock()})

		registration := harness.awaitRegistration(t, "After attribute transition")
		if registration.requested != 3*time.Minute {
			t.Fatalf("After attribute duration = %v, want %v", registration.requested, 3*time.Minute)
		}
		entered := hsm.AfterEntry(sm.Context(), sm, "/CanonicalAfterAttributeHSM/done")
		registration.trigger(t)
		awaitWaiter(t, "After attribute entering done", entered)
	})

	t.Run("Every", func(t *testing.T) {
		harness := newDeterministicClockHarness()
		ticks := make(chan struct{}, 1)
		model := hsm.Define(
			"CanonicalEveryAttributeHSM",
			hsm.Attribute("delay", 4*time.Minute),
			hsm.Initial(hsm.Target("waiting")),
			hsm.State(
				"waiting",
				hsm.Transition(
					hsm.Every("delay"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						ticks <- struct{}{}
					}),
				),
			),
		)
		hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Clock: harness.Clock()})

		registration := harness.awaitRegistration(t, "Every attribute transition")
		if registration.requested != 4*time.Minute {
			t.Fatalf("Every attribute duration = %v, want %v", registration.requested, 4*time.Minute)
		}
		registration.trigger(t)
		select {
		case <-ticks:
		case <-time.After(waiterDeadline):
			t.Fatal("timed out waiting for Every attribute tick")
		}
	})

	t.Run("At", func(t *testing.T) {
		harness := newDeterministicClockHarness()
		deadline := time.Now().Add(time.Hour)
		model := hsm.Define(
			"CanonicalAtAttributeHSM",
			hsm.Attribute("deadline", deadline),
			hsm.Initial(hsm.Target("waiting")),
			hsm.State(
				"waiting",
				hsm.Transition(hsm.At("deadline"), hsm.Target("../done")),
			),
			hsm.State("done"),
		)
		sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{Clock: harness.Clock()})

		registration := harness.awaitRegistration(t, "At attribute transition")
		if registration.requested < 59*time.Minute || registration.requested > time.Hour {
			t.Fatalf("At attribute duration = %v, want about one hour", registration.requested)
		}
		entered := hsm.AfterEntry(sm.Context(), sm, "/CanonicalAtAttributeHSM/done")
		registration.trigger(t)
		awaitWaiter(t, "At attribute entering done", entered)
	})
}
