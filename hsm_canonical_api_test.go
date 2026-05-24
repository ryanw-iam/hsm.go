package hsm_test

import (
	"context"
	"testing"

	"github.com/stateforward/hsm.go"
)

func TestCanonicalKindAPIs(t *testing.T) {
	base := hsm.MakeKind()
	derived := hsm.MakeKind(base)

	if !hsm.IsKind(derived, base) {
		t.Fatal("IsKind() should report derived kind as base kind")
	}
	if !hsm.IsKind(derived, derived) {
		t.Fatal("IsKind() should report kind as itself")
	}
	if hsm.IsKind(base, derived) {
		t.Fatal("IsKind() should not report base kind as derived kind")
	}
}

func TestOnAcceptsStringEventNames(t *testing.T) {
	type eventName string

	model := hsm.Define(
		"CanonicalStringOnHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State(
			"idle",
			hsm.Transition(hsm.On("go"), hsm.Target("../done")),
		),
		hsm.State("done"),
	)
	sm := hsm.Started(context.Background(), &THSM{}, &model)

	<-hsm.Dispatch(context.Background(), sm, hsm.Event{Name: "go"})
	if got := sm.State(); got != "/CanonicalStringOnHSM/done" {
		t.Fatalf("string On transition state = %q, want done", got)
	}

	aliasModel := hsm.Define(
		"CanonicalStringAliasOnHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State(
			"idle",
			hsm.Transition(hsm.On(eventName("advance")), hsm.Target("../done")),
		),
		hsm.State("done"),
	)
	aliasSM := hsm.Started(context.Background(), &THSM{}, &aliasModel)

	<-hsm.Dispatch(context.Background(), aliasSM, hsm.Event{Name: "advance"})
	if got := aliasSM.State(); got != "/CanonicalStringAliasOnHSM/done" {
		t.Fatalf("string alias On transition state = %q, want done", got)
	}
}

func TestMakeGroupAliasesNewGroupAndSupportsLeadingID(t *testing.T) {
	model := hsm.Define(
		"CanonicalGroupHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
	)
	first := hsm.New(&THSM{}, &model)
	second := hsm.New(&THSM{}, &model)
	nested := hsm.NewGroup(second)

	group := hsm.MakeGroup("workers", first, nil, nested)

	if got := hsm.ID(group); got != "workers" {
		t.Fatalf("ID(MakeGroup(...)) = %q, want workers", got)
	}
	if got := len(group.Instances()); got != 2 {
		t.Fatalf("MakeGroup instance count = %d, want 2", got)
	}
}

func TestTakeSnapshotCopiesMutableAttributeAndSchemaValues(t *testing.T) {
	schema := map[string]any{
		"owner":  "model",
		"nested": map[string]any{"value": "model"},
	}
	model := hsm.Define(
		"CanonicalSnapshotCopyHSM",
		hsm.Attribute("bag", map[string]any{
			"owner":  "runtime",
			"nested": []string{"runtime"},
		}),
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
		hsm.Transition(
			hsm.On(hsm.Event{Name: "go", Kind: hsm.EventKind, Schema: schema}),
			hsm.Source("idle"),
			hsm.Target("idle"),
		),
	)
	sm := hsm.Started(context.Background(), &THSM{}, &model)

	snapshot := hsm.TakeSnapshot(context.Background(), sm)
	attribute := snapshot.Attributes["/CanonicalSnapshotCopyHSM/bag"].(map[string]any)
	attribute["owner"] = "snapshot"
	attribute["nested"].([]string)[0] = "snapshot"
	eventSchema := snapshot.Events[0].Schema.(map[string]any)
	eventSchema["owner"] = "snapshot"
	eventSchema["nested"].(map[string]any)["value"] = "snapshot"

	value, ok := hsm.Get(context.Background(), sm, "bag")
	if !ok {
		t.Fatal("expected runtime attribute to exist")
	}
	runtimeAttribute := value.(map[string]any)
	if got := runtimeAttribute["owner"]; got != "runtime" {
		t.Fatalf("runtime attribute owner = %v, want runtime", got)
	}
	if got := runtimeAttribute["nested"].([]string)[0]; got != "runtime" {
		t.Fatalf("runtime nested attribute = %v, want runtime", got)
	}
	if got := schema["owner"]; got != "model" {
		t.Fatalf("model schema owner = %v, want model", got)
	}
	if got := schema["nested"].(map[string]any)["value"]; got != "model" {
		t.Fatalf("model nested schema = %v, want model", got)
	}

	fresh := hsm.TakeSnapshot(context.Background(), sm)
	freshAttribute := fresh.Attributes["/CanonicalSnapshotCopyHSM/bag"].(map[string]any)
	if got := freshAttribute["owner"]; got != "runtime" {
		t.Fatalf("fresh snapshot attribute owner = %v, want runtime", got)
	}
	if got := freshAttribute["nested"].([]string)[0]; got != "runtime" {
		t.Fatalf("fresh snapshot nested attribute = %v, want runtime", got)
	}
	freshSchema := fresh.Events[0].Schema.(map[string]any)
	if got := freshSchema["owner"]; got != "model" {
		t.Fatalf("fresh snapshot schema owner = %v, want model", got)
	}
	if got := freshSchema["nested"].(map[string]any)["value"]; got != "model" {
		t.Fatalf("fresh snapshot nested schema = %v, want model", got)
	}
}
