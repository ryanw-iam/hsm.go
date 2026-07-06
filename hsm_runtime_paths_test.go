package hsm_test

import (
	"context"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

type runtimeCallbackContextKey struct{}

type runtimeCallbackObservation struct {
	instanceID string
	currentID  string
	request    string
	source     string
	target     string
}

func nextRuntimeCallbackObservation(t *testing.T, observations <-chan runtimeCallbackObservation) runtimeCallbackObservation {
	t.Helper()
	select {
	case observation := <-observations:
		return observation
	case <-time.After(waiterDeadline):
		t.Fatal("timed out waiting for runtime callback context observation")
	}
	return runtimeCallbackObservation{}
}

func TestRuntimePathIdentityContextAndSnapshotHelpers(t *testing.T) {
	model := hsm.Define(
		"RuntimePathHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	defer func() { <-hsm.Stop(context.Background(), sm) }()

	if got, ok := hsm.FromContext(sm.Context()); !ok || hsm.ID(got) != hsm.ID(sm) {
		t.Fatalf("FromContext() = (%v, %v), want matching instance id %q", got, ok, hsm.ID(sm))
	}

	instances, ok := hsm.InstancesFromContext(sm.Context())
	if !ok || len(instances) != 1 || hsm.ID(instances[0]) != hsm.ID(sm) {
		t.Fatalf("InstancesFromContext() = (%v, %v), want one instance matching %q", instances, ok, hsm.ID(sm))
	}

	group := hsm.NewGroup(nil, hsm.NewGroup())
	if got := len(group.Instances()); got != 0 {
		t.Fatalf("group instance count = %d, want 0 for empty nested groups", got)
	}

	if got := hsm.Name(sm); got != "RuntimePathHSM" {
		t.Fatalf("Name() = %q, want RuntimePathHSM", got)
	}
	if got := hsm.QualifiedName(sm); got != "/RuntimePathHSM" {
		t.Fatalf("QualifiedName() = %q, want /RuntimePathHSM", got)
	}
	if got := hsm.ID(sm); got == "" {
		t.Fatal("ID() should not be empty")
	}

	snapshot := hsm.TakeSnapshot(context.Background(), sm)
	if snapshot.State != "/RuntimePathHSM/idle" {
		t.Fatalf("snapshot state = %q, want /RuntimePathHSM/idle", snapshot.State)
	}
	if snapshot.ID == "" || snapshot.QualifiedName == "" {
		t.Fatalf("snapshot should include identity fields, got %+v", snapshot)
	}
}

func TestRuntimeCallbackContextUsesExecutingMachineForBareDispatch(t *testing.T) {
	observe := hsm.Event{Name: "observe"}
	observations := make(chan runtimeCallbackObservation, 1)
	model := hsm.Define(
		"RuntimeCallbackContextBareHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(
				hsm.On(observe),
				hsm.Target("../idle"),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					current, _ := hsm.FromContext(ctx)
					request, _ := ctx.Value(runtimeCallbackContextKey{}).(string)
					observations <- runtimeCallbackObservation{
						instanceID: hsm.ID(sm),
						currentID:  hsm.ID(current),
						request:    request,
						source:     event.Source,
						target:     event.Target,
					}
				}),
			),
		),
	)
	sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{ID: "solo"})
	defer func() { <-hsm.Stop(context.Background(), sm) }()

	ctx := context.WithValue(context.Background(), runtimeCallbackContextKey{}, "direct")
	awaitWaiter(t, "bare-context callback dispatch", hsm.Dispatch(ctx, sm, observe))

	observation := nextRuntimeCallbackObservation(t, observations)
	if observation.instanceID != "solo" || observation.currentID != "solo" {
		t.Fatalf("callback current HSM = %+v, want executing solo instance", observation)
	}
	if observation.request != "direct" {
		t.Fatalf("callback request value = %q, want direct", observation.request)
	}
}

func TestRuntimeCallbackContextUsesRecipientForDispatchFanout(t *testing.T) {
	observe := hsm.Event{Name: "observe"}
	observations := make(chan runtimeCallbackObservation, 8)
	model := hsm.Define(
		"RuntimeCallbackContextFanoutHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(
				hsm.On(observe),
				hsm.Target("../idle"),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					current, _ := hsm.FromContext(ctx)
					request, _ := ctx.Value(runtimeCallbackContextKey{}).(string)
					observations <- runtimeCallbackObservation{
						instanceID: hsm.ID(sm),
						currentID:  hsm.ID(current),
						request:    request,
						source:     event.Source,
						target:     event.Target,
					}
				}),
			),
		),
	)
	source := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{ID: "source"})
	bravo := hsm.Started(source.Context(), &THSM{}, &model, hsm.Config{ID: "bravo"})
	charlie := hsm.Started(bravo.Context(), &THSM{}, &model, hsm.Config{ID: "charlie"})
	defer func() { <-hsm.Stop(context.Background(), source) }()
	defer func() { <-hsm.Stop(context.Background(), bravo) }()
	defer func() { <-hsm.Stop(context.Background(), charlie) }()

	dispatchToCtx := context.WithValue(source.Context(), runtimeCallbackContextKey{}, "dispatch-to")
	awaitWaiter(t, "recipient-context dispatch-to", hsm.DispatchTo(dispatchToCtx, observe, "bravo", "charlie"))
	dispatchToObservations := map[string]runtimeCallbackObservation{}
	for range 2 {
		observation := nextRuntimeCallbackObservation(t, observations)
		dispatchToObservations[observation.instanceID] = observation
	}
	for _, id := range []string{"bravo", "charlie"} {
		observation, ok := dispatchToObservations[id]
		if !ok {
			t.Fatalf("missing dispatch-to observation for %s: %v", id, dispatchToObservations)
		}
		if observation.currentID != id {
			t.Fatalf("dispatch-to current HSM for %s = %q, want recipient; observation=%+v", id, observation.currentID, observation)
		}
		if observation.request != "dispatch-to" || observation.source != "source" || observation.target != id {
			t.Fatalf("dispatch-to observation for %s = %+v, want request/source/target preserved", id, observation)
		}
	}

	dispatchAllCtx := context.WithValue(source.Context(), runtimeCallbackContextKey{}, "dispatch-all")
	awaitWaiter(t, "recipient-context dispatch-all", hsm.DispatchAll(dispatchAllCtx, observe))
	dispatchAllObservations := map[string]runtimeCallbackObservation{}
	for range 3 {
		observation := nextRuntimeCallbackObservation(t, observations)
		dispatchAllObservations[observation.instanceID] = observation
	}
	for _, id := range []string{"source", "bravo", "charlie"} {
		observation, ok := dispatchAllObservations[id]
		if !ok {
			t.Fatalf("missing dispatch-all observation for %s: %v", id, dispatchAllObservations)
		}
		if observation.currentID != id {
			t.Fatalf("dispatch-all current HSM for %s = %q, want recipient; observation=%+v", id, observation.currentID, observation)
		}
		if observation.request != "dispatch-all" || observation.source != "source" || observation.target != id {
			t.Fatalf("dispatch-all observation for %s = %+v, want request/source/target preserved", id, observation)
		}
	}
}

func TestRuntimePathRestartResetsState(t *testing.T) {
	event := hsm.Event{Name: "advance"}
	model := hsm.Define(
		"RestartPathHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
		hsm.Transition(hsm.On(event), hsm.Source("idle"), hsm.Target("done")),
		hsm.State("done"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	defer func() { <-hsm.Stop(context.Background(), sm) }()

	<-hsm.Dispatch(sm.Context(), sm, event)
	if got := sm.State(); got != "/RestartPathHSM/done" {
		t.Fatalf("state after dispatch = %q, want /RestartPathHSM/done", got)
	}

	<-hsm.Restart(context.Background(), sm)
	if got := sm.State(); got != "/RestartPathHSM/idle" {
		t.Fatalf("state after restart = %q, want /RestartPathHSM/idle", got)
	}
}

func TestRuntimeSourceQualifiedTransitionDeclaredUnderSiblingIsReachable(t *testing.T) {
	advance := hsm.Event{Name: "advance"}
	model := hsm.Define(
		"SiblingOwnerSourceHSM",
		hsm.Initial(hsm.Target("left")),
		hsm.State("left",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
		),
		hsm.State("right",
			hsm.Transition(
				hsm.On(advance),
				hsm.Source("/SiblingOwnerSourceHSM/left/idle"),
				hsm.Target("/SiblingOwnerSourceHSM/right"),
			),
		),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	defer func() { <-hsm.Stop(context.Background(), sm) }()

	<-hsm.Dispatch(context.Background(), sm, advance)
	if got := sm.State(); got != "/SiblingOwnerSourceHSM/right" {
		t.Fatalf("state after source-qualified dispatch = %q, want /SiblingOwnerSourceHSM/right", got)
	}
}

func TestRuntimeEntryPointSelectsNamedEntryPoint(t *testing.T) {
	start := hsm.Event{Name: "start"}
	trace := &runtimeTrace{}
	child := hsm.Define(
		"NativeEntryPointChildHSM",
		hsm.EntryPoint("warm",
			hsm.Target("running"),
			hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
				trace.record("entry:warm")
			}),
		),
		hsm.Initial(hsm.Target("cold")),
		hsm.State("cold"),
		hsm.State("running"),
	)
	model := hsm.Define(
		"NativeEntryPointHSM",
		hsm.Initial(hsm.Target("outside")),
		hsm.State("outside",
			hsm.Transition(
				hsm.On(start),
				hsm.Target("../drive"),
				hsm.EntryPoint("warm"),
			),
		),
		hsm.SubmachineState("drive", child),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	defer func() { <-hsm.Stop(context.Background(), sm) }()

	awaitWaiter(t, "entry point start dispatch", hsm.Dispatch(context.Background(), sm, start))
	if got := sm.State(); got != "/NativeEntryPointHSM/drive/running" {
		t.Fatalf("state after entry-point dispatch = %q, want /NativeEntryPointHSM/drive/running", got)
	}
	if entries := trace.snapshot(); len(entries) != 1 || entries[0] != "entry:warm" {
		t.Fatalf("entry-point trace = %v, want [entry:warm]", entries)
	}
}

func TestRuntimeEntryPointSelectorRejectsDescendantEntryPoint(t *testing.T) {
	start := hsm.Event{Name: "start"}
	assertPanicContains(t, "ordinary entry point selector", "requires a submachine transition target", func() {
		_ = hsm.Define(
			"BadOrdinaryEntryPointHSM",
			hsm.Initial(hsm.Target("outside")),
			hsm.State("outside",
				hsm.Transition(
					hsm.On(start),
					hsm.Target("../target"),
					hsm.EntryPoint("warm"),
				),
			),
			hsm.State("target",
				hsm.EntryPoint("warm", hsm.Target("ready")),
				hsm.Initial(hsm.Target("ready")),
				hsm.State("ready"),
			),
		)
	})
}

func TestRuntimeExitPointRoutesThroughSubmachineBoundary(t *testing.T) {
	finish := hsm.Event{Name: "finish"}
	child := hsm.Define(
		"NativeExitPointChildHSM",
		hsm.ExitPoint("done"),
		hsm.Initial(hsm.Target("running")),
		hsm.State("running",
			hsm.Transition(
				hsm.On(finish),
				hsm.Target("../done"),
			),
		),
	)
	model := hsm.Define(
		"NativeExitPointHSM",
		hsm.Initial(hsm.Target("drive")),
		hsm.SubmachineState("drive", child,
			hsm.Transition(
				hsm.ExitPoint("done"),
				hsm.Target("../complete"),
			),
		),
		hsm.State("complete"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	defer func() { <-hsm.Stop(context.Background(), sm) }()

	<-hsm.Dispatch(context.Background(), sm, finish)
	if got := sm.State(); got != "/NativeExitPointHSM/complete" {
		t.Fatalf("state after exit-point dispatch = %q, want /NativeExitPointHSM/complete", got)
	}
}

func TestRuntimeSourceQualifiedTransitionBeatsDeferredEvent(t *testing.T) {
	work := hsm.Event{Name: "work"}
	model := hsm.Define(
		"SourceQualifiedDeferHSM",
		hsm.Initial(hsm.Target("blocked")),
		hsm.State("blocked", hsm.Defer(work)),
		hsm.Transition(
			hsm.On(work),
			hsm.Source("blocked"),
			hsm.Target("done"),
		),
		hsm.State("done"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	defer func() { <-hsm.Stop(context.Background(), sm) }()

	<-hsm.Dispatch(context.Background(), sm, work)
	if got := sm.State(); got != "/SourceQualifiedDeferHSM/done" {
		t.Fatalf("state after source-qualified deferred event = %q, want /SourceQualifiedDeferHSM/done", got)
	}
}
