package hsm_test

import (
	"context"
	"testing"

	hsm "github.com/stateforward/hsm.go"
)

func TestEventMetadataMutationsDoNotReachCallerEvent(t *testing.T) {
	type OwnershipHSM struct {
		hsm.HSM
	}

	handled := make(chan hsm.Event, 1)
	model := hsm.Define(
		"EventMetadataCallerOwnershipHSM",
		hsm.State("idle"),
		hsm.Initial(hsm.Target("idle")),
		hsm.Transition(
			hsm.On(hsm.Event{Name: "mutate"}),
			hsm.Source("idle"),
			hsm.Target("idle"),
			hsm.Effect(func(ctx context.Context, sm *OwnershipHSM, event hsm.Event) {
				event.Kind = 9001
				event.Name = "mutated-name"
				event.ID = "mutated-id"
				event.Source = "mutated-source"
				event.Target = "mutated-target"
				schema := event.Schema.(map[string]any)
				schema["owner"] = "behavior"
				schema["nested"].(map[string]any)["value"] = "behavior"
				event.Data.(map[string]string)["value"] = "behavior"
				handled <- event
			}),
		),
	)

	schema := map[string]any{
		"owner":  "caller",
		"nested": map[string]any{"value": "caller"},
	}
	data := map[string]string{"value": "caller"}
	event := hsm.Event{
		Kind:   42,
		Name:   "mutate",
		ID:     "caller-id",
		Source: "caller-source",
		Target: "caller-target",
		Schema: schema,
		Data:   data,
	}

	sm := hsm.Started(context.Background(), &OwnershipHSM{}, &model)
	awaitWaiter(t, "metadata mutation dispatch", hsm.Dispatch(sm.Context(), sm, event))
	<-handled

	if event.Kind != 42 || event.Name != "mutate" || event.ID != "caller-id" || event.Source != "caller-source" || event.Target != "caller-target" {
		t.Fatalf("caller event metadata was mutated: %#v", event)
	}
	if got := schema["owner"]; got != "caller" {
		t.Fatalf("caller schema owner = %v, want caller", got)
	}
	if got := schema["nested"].(map[string]any)["value"]; got != "caller" {
		t.Fatalf("caller nested schema value = %v, want caller", got)
	}
	if got := data["value"]; got != "behavior" {
		t.Fatalf("data value = %q, want behavior to preserve application-owned reference semantics", got)
	}
}

func TestDispatchToIsolatesSiblingEventMetadataButSharesDataReference(t *testing.T) {
	type OwnershipHSM struct {
		hsm.HSM
	}

	alphaDone := make(chan struct{})
	bravoSaw := make(chan struct {
		schemaOwner string
		nestedValue string
		dataValue   string
	}, 1)

	model := hsm.Define(
		"EventMetadataSiblingOwnershipHSM",
		hsm.State("idle"),
		hsm.Initial(hsm.Target("idle")),
		hsm.Transition(
			hsm.On(hsm.Event{Name: "broadcast"}),
			hsm.Source("idle"),
			hsm.Target("idle"),
			hsm.Effect(func(ctx context.Context, sm *OwnershipHSM, event hsm.Event) {
				switch hsm.ID(sm) {
				case "alpha":
					schema := event.Schema.(map[string]any)
					schema["owner"] = "alpha"
					schema["nested"].(map[string]any)["value"] = "alpha"
					event.Data.(map[string]string)["alpha"] = "seen"
					close(alphaDone)
				case "bravo":
					<-alphaDone
					schema := event.Schema.(map[string]any)
					bravoSaw <- struct {
						schemaOwner string
						nestedValue string
						dataValue   string
					}{
						schemaOwner: schema["owner"].(string),
						nestedValue: schema["nested"].(map[string]any)["value"].(string),
						dataValue:   event.Data.(map[string]string)["alpha"],
					}
				}
			}),
		),
	)

	ctx := context.Background()
	alpha := hsm.Started(ctx, &OwnershipHSM{}, &model, hsm.Config{ID: "alpha"})
	bravo := hsm.Started(alpha.Context(), &OwnershipHSM{}, &model, hsm.Config{ID: "bravo"})
	schema := map[string]any{
		"owner":  "caller",
		"nested": map[string]any{"value": "caller"},
	}
	event := hsm.Event{
		Name:   "broadcast",
		Schema: schema,
		Data:   map[string]string{},
	}

	awaitWaiter(t, "targeted metadata ownership dispatch", hsm.DispatchTo(bravo.Context(), event, "alpha", "bravo"))
	got := <-bravoSaw
	if got.schemaOwner != "caller" {
		t.Fatalf("bravo schema owner = %q, want caller", got.schemaOwner)
	}
	if got.nestedValue != "caller" {
		t.Fatalf("bravo nested schema value = %q, want caller", got.nestedValue)
	}
	if got.dataValue != "seen" {
		t.Fatalf("bravo data value = %q, want shared application-owned data", got.dataValue)
	}
	if got := schema["owner"]; got != "caller" {
		t.Fatalf("caller schema owner = %v, want caller", got)
	}
}
