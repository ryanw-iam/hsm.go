package hsm_test

import (
	"context"
	"testing"

	"github.com/stateforward/hsm.go"
)

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
