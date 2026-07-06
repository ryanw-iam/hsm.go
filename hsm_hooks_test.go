package hsm_test

import (
	"context"
	"path"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

type operationRefHSM struct {
	hsm.HSM
	mu    sync.Mutex
	order []string
}

func (sm *operationRefHSM) Approve(ctx context.Context, event hsm.Event) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.order = append(sm.order, "approve:"+event.Name)
}

func (sm *operationRefHSM) Allowed(ctx context.Context, event hsm.Event) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.order = append(sm.order, "allowed:"+event.Name)
	return true
}

func (sm *operationRefHSM) Order() []string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return append([]string(nil), sm.order...)
}

func TestModelValidatorAndFinalizerHooks(t *testing.T) {
	validatorCalls := []string{}
	finalizerCalls := 0
	model := hsm.Define(
		"HookHSM",
		hsm.Validator(func(model *hsm.Model) {
			validatorCalls = append(validatorCalls, "first")
		}),
		hsm.Validator(func(model *hsm.Model) {
			if model.Members()[path.Join(model.QualifiedName(), "ready")] == nil {
				t.Fatal("validator did not receive fully built model")
			}
			validatorCalls = append(validatorCalls, "second")
		}),
		hsm.Finalizer(func(model *hsm.Model) *hsm.FinalizedModel {
			finalizerCalls++
			return hsm.DefaultModelFinalizer{}.Finalize(model)
		}),
		hsm.Initial(hsm.Target("ready")),
		hsm.State("ready"),
	)

	if !slices.Equal(validatorCalls, []string{"second"}) {
		t.Fatalf("validator calls = %v, want last validator only", validatorCalls)
	}
	if finalizerCalls != 1 {
		t.Fatalf("finalizer calls = %d, want 1", finalizerCalls)
	}
	sm := hsm.Started(context.Background(), &THSM{}, &model)
	if sm.State() != "/HookHSM/ready" {
		t.Fatalf("hooked model did not start: %s", sm.State())
	}

	redefined := hsm.Redefine(model, "HookHSMRedefined")
	if !slices.Equal(validatorCalls, []string{"second", "second"}) {
		t.Fatalf("validator calls after redefine = %v, want inherited validator", validatorCalls)
	}
	if finalizerCalls != 2 {
		t.Fatalf("finalizer calls after redefine = %d, want 2", finalizerCalls)
	}
	redefinedSM := hsm.Started(context.Background(), &THSM{}, &redefined)
	if redefinedSM.State() != "/HookHSMRedefined/ready" {
		t.Fatalf("redefined hooked model did not start: %s", redefinedSM.State())
	}
}

func TestModelMembersReturnsSnapshot(t *testing.T) {
	model := hsm.Define(
		"MembersSnapshotHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
	)
	members := model.Members()
	delete(members, "/MembersSnapshotHSM/idle")
	if model.Members()["/MembersSnapshotHSM/idle"] == nil {
		t.Fatal("Members() returned a live mutable member map")
	}
}

func TestOperationReferencesResolveInstanceMethods(t *testing.T) {
	model := hsm.Define(
		"OperationRefHSM",
		hsm.Operation("approve"),
		hsm.Operation("allowed"),
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Entry("approve"),
			hsm.Transition(
				hsm.On("go"),
				hsm.Guard("allowed"),
				hsm.Target("../done"),
				hsm.Effect("approve"),
			),
		),
		hsm.State("done"),
	)
	sm := hsm.Started(context.Background(), &operationRefHSM{}, &model)
	awaitWaiter(t, "operation-ref dispatch", hsm.Dispatch(context.Background(), sm, hsm.Event{Name: "go"}))
	if sm.State() != "/OperationRefHSM/done" {
		t.Fatalf("state = %s, want done", sm.State())
	}
	want := []string{"approve:hsm/initial", "allowed:go", "approve:go"}
	if got := sm.Order(); !slices.Equal(got, want) {
		t.Fatalf("operation reference order = %v, want %v", got, want)
	}
}

func TestObserveWrapsBehaviorsAndTransitionEvents(t *testing.T) {
	observations := make(chan hsm.Event, 8)
	effectRan := make(chan struct{}, 1)
	goEvent := hsm.Event{Name: "go"}
	model := hsm.Define(
		"ObserveHSM",
		hsm.Observe(func(ctx context.Context, sm *THSM, event hsm.Event) {
			observations <- event
		}),
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(
				hsm.On(goEvent),
				hsm.Target("../done"),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					effectRan <- struct{}{}
				}),
			),
		),
		hsm.State("done"),
	)
	sm := hsm.Started(context.Background(), &THSM{}, &model)
	for {
		select {
		case <-observations:
		default:
			goto drained
		}
	}
drained:

	awaitWaiter(t, "observed dispatch", hsm.Dispatch(context.Background(), sm, goEvent))
	select {
	case <-effectRan:
	case <-time.After(time.Second):
		t.Fatal("effect did not run")
	}

	sawTransition := false
	sawBehavior := false
	deadline := time.After(time.Second)
	for !sawTransition || !sawBehavior {
		select {
		case event := <-observations:
			if event.Name != hsm.ObservationEvent.Name {
				t.Fatalf("observation name = %q", event.Name)
			}
			data, ok := event.Data.(hsm.ObservationData)
			if !ok {
				t.Fatalf("observation data type = %T", event.Data)
			}
			if data.Event.Name != goEvent.Name {
				continue
			}
			switch data.Occurrence {
			case "event":
				sawTransition = true
			case "behavior":
				sawBehavior = true
			}
		case <-deadline:
			t.Fatalf("missing observations: transition=%v behavior=%v", sawTransition, sawBehavior)
		}
	}
}
