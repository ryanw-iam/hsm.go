package hsm_test

import (
	"context"
	"testing"

	"github.com/stateforward/hsm.go"
)

func TestDispatchRegressionAllocs(t *testing.T) {
	const runs = 1000

	workEvent := hsm.Event{Name: "work"}
	model := hsm.Define(
		"DispatchRegressionAllocsHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	t.Cleanup(func() {
		awaitWaiter(t, "DispatchRegressionAllocsHSM stop", hsm.Stop(context.Background(), sm))
	})

	<-hsm.Dispatch(sm.Context(), sm, workEvent)

	allocs := testing.AllocsPerRun(runs, func() {
		<-hsm.Dispatch(sm.Context(), sm, workEvent)
	})

	assertAllocsWithin(t, "Dispatch", allocs, 5)
}

func TestSetRegressionAllocs(t *testing.T) {
	const runs = 1000

	model := hsm.Define(
		"SetRegressionAllocsHSM",
		hsm.Attribute("count", 0),
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(hsm.OnSet("count"), hsm.Target("../idle")),
		),
	)

	sm := hsm.Started(context.Background(), &AttrHSM{}, &model)
	t.Cleanup(func() {
		awaitWaiter(t, "SetRegressionAllocsHSM stop", hsm.Stop(context.Background(), sm))
	})

	next := 0
	<-hsm.Set(sm.Context(), sm, "count", next)

	allocs := testing.AllocsPerRun(runs, func() {
		next = 1 - next
		<-hsm.Set(sm.Context(), sm, "count", next)
	})

	assertAllocsWithin(t, "Set", allocs, 9)
}

func assertAllocsWithin(t *testing.T, name string, got float64, max float64) {
	t.Helper()
	if got > max {
		t.Fatalf("%s allocs/run = %.2f, want <= %.0f", name, got, max)
	}
}
