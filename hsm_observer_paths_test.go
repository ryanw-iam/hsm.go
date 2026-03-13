package hsm_test

import (
	"context"
	"testing"

	"github.com/stateforward/hsm.go"
)

func TestObserverPathWaiters(t *testing.T) {
	event := hsm.Event{Name: "advance"}
	model := hsm.Define(
		"ObserverPathHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
		hsm.Transition(hsm.On(event), hsm.Source("idle"), hsm.Target("done")),
		hsm.State("done"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	defer func() { <-hsm.Stop(context.Background(), sm) }()

	dispatched := hsm.AfterDispatch(sm.Context(), sm, event)
	processed := hsm.AfterProcess(sm.Context(), sm, event)
	entered := hsm.AfterEntry(sm.Context(), sm, "/ObserverPathHSM/done")
	exited := hsm.AfterExit(sm.Context(), sm, "/ObserverPathHSM/idle")
	<-hsm.Dispatch(sm.Context(), sm, event)

	for name, waiter := range map[string]<-chan struct{}{
		"dispatch": dispatched,
		"process":  processed,
		"entry":    entered,
		"exit":     exited,
	} {
		select {
		case <-waiter:
		default:
			t.Fatalf("%s waiter did not fire", name)
		}
	}
}
