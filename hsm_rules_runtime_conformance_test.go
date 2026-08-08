package hsm_test

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

type runtimeRuleContextKey string

type runtimeRuleRequest struct {
	Value string
	Reply chan string
}

type runtimeRuleAttributeResult struct {
	initial  int
	final    int
	state    string
	observed bool
	change   hsm.AttributeChange
	kind     uint64
}

func TestRuntimeRulesConformance(t *testing.T) {
	t.Run("HSM02/composite_initial_enters_nested_state", func(t *testing.T) {
		model := hsm.Define(
			"CompositeInitialRulesHSM",
			hsm.State("parent",
				hsm.State("child"),
				hsm.Initial(hsm.Target("child")),
			),
			hsm.Initial(hsm.Target("parent")),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		assertRuleState(t, sm.State(), "/CompositeInitialRulesHSM/parent/child")
	})

	t.Run("HSM10/completion_transition_follows_final_child", func(t *testing.T) {
		advanceEvent := hsm.Event{Name: "advance"}
		model := hsm.Define(
			"CompletionTransitionRulesHSM",
			hsm.Initial(hsm.Target("running")),
			hsm.State("running",
				hsm.Initial(hsm.Target("work")),
				hsm.State("work",
					hsm.Transition(hsm.On(advanceEvent), hsm.Target("../complete")),
				),
				hsm.Final("complete"),
			),
			hsm.State("done"),
			hsm.Transition(
				hsm.Source("running"),
				hsm.Target("done"),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "completion transition advance", hsm.Dispatch(sm.Context(), sm, advanceEvent))
		assertRuleState(t, sm.State(), "/CompletionTransitionRulesHSM/done")
	})

	t.Run("HSM11/completion_event_drives_prioritized_follow_on_transition", func(t *testing.T) {
		triggerEvent := hsm.Event{Name: "trigger"}
		regularEvent := hsm.Event{Name: "regular"}
		priorityEvent := hsm.Event{Name: "priority", Kind: hsm.CompletionEventKind}
		trace := &runtimeTrace{}

		model := hsm.Define(
			"CompletionPriorityRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(
					hsm.On(triggerEvent),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("trigger")
						hsm.Dispatch(ctx, sm, regularEvent)
						hsm.Dispatch(ctx, sm, priorityEvent)
					}),
				),
				hsm.Transition(
					hsm.On(priorityEvent),
					hsm.Target("../priority"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("priority")
					}),
				),
				hsm.Transition(
					hsm.On(regularEvent),
					hsm.Target("../regular"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("regular")
					}),
				),
			),
			hsm.State("priority"),
			hsm.State("regular"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		priorityProcessed := hsm.AfterProcess(sm.Context(), sm, priorityEvent)
		regularProcessed := hsm.AfterProcess(sm.Context(), sm, regularEvent)

		awaitWaiter(t, "completion-priority trigger dispatch", hsm.Dispatch(sm.Context(), sm, triggerEvent))
		awaitWaiter(t, "completion-priority completion event", priorityProcessed)
		awaitWaiter(t, "completion-priority queued regular event", regularProcessed)

		assertRuleState(t, sm.State(), "/CompletionPriorityRulesHSM/priority")
		assertRuleTrace(t, trace.snapshot(), []string{"trigger", "priority"})
	})

	t.Run("HSM12/explicit_events_or_any_event_only", func(t *testing.T) {
		trace := &runtimeTrace{}
		model := defineAnyEventRulesModel(trace)
		sm := hsm.Started(context.Background(), &THSM{}, &model)

		enteredSpecial := hsm.AfterEntry(sm.Context(), sm, "/AnyEventRulesHSM/special")
		awaitWaiter(t, "explicit special dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: "special"}))
		awaitWaiter(t, "explicit special entry", enteredSpecial)
		assertRuleState(t, sm.State(), "/AnyEventRulesHSM/special")

		enteredReady := hsm.AfterEntry(sm.Context(), sm, "/AnyEventRulesHSM/ready")
		awaitWaiter(t, "explicit reset dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: "reset"}))
		awaitWaiter(t, "explicit reset ready entry", enteredReady)

		enteredFallback := hsm.AfterEntry(sm.Context(), sm, "/AnyEventRulesHSM/fallback")
		awaitWaiter(t, "any-event fallback dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: "other"}))
		awaitWaiter(t, "any-event fallback entry", enteredFallback)
		assertRuleState(t, sm.State(), "/AnyEventRulesHSM/fallback")
		assertRuleTrace(t, trace.snapshot(), []string{"special", "fallback:other"})
	})

	t.Run("HSM13/any_event_only_fires_as_fallback", func(t *testing.T) {
		trace := &runtimeTrace{}
		model := defineAnyEventRulesModel(trace)
		sm := hsm.Started(context.Background(), &THSM{}, &model)

		enteredFallback := hsm.AfterEntry(sm.Context(), sm, "/AnyEventRulesHSM/fallback")
		awaitWaiter(t, "fallback any-event dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: "other"}))
		awaitWaiter(t, "fallback any-event entry", enteredFallback)

		assertRuleState(t, sm.State(), "/AnyEventRulesHSM/fallback")
		assertRuleTrace(t, trace.snapshot(), []string{"fallback:other"})
	})

	t.Run("HSM14/specific_event_precedes_any_event", func(t *testing.T) {
		trace := &runtimeTrace{}
		model := defineAnyEventRulesModel(trace)
		sm := hsm.Started(context.Background(), &THSM{}, &model)

		enteredSpecial := hsm.AfterEntry(sm.Context(), sm, "/AnyEventRulesHSM/special")
		awaitWaiter(t, "specific-event dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: "special"}))
		awaitWaiter(t, "specific-event entry", enteredSpecial)

		assertRuleState(t, sm.State(), "/AnyEventRulesHSM/special")
		assertRuleTrace(t, trace.snapshot(), []string{"special"})
	})

	t.Run("HSM15/first_passing_guard_wins_for_same_event", func(t *testing.T) {
		trace := &runtimeTrace{}
		model := defineAnyEventRulesModel(trace)
		sm := hsm.Started(context.Background(), &THSM{}, &model)

		enteredGuarded := hsm.AfterEntry(sm.Context(), sm, "/AnyEventRulesHSM/guarded")
		awaitWaiter(t, "guard-order dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: "guarded"}))
		awaitWaiter(t, "guard-order entry", enteredGuarded)

		assertRuleState(t, sm.State(), "/AnyEventRulesHSM/guarded")
		assertRuleTrace(t, trace.snapshot(), []string{"guard_second"})
	})

	t.Run("HSM16/choice_vertex_models_conditional_branching", func(t *testing.T) {
		chooseEvent := hsm.Event{Name: "choose"}
		model := hsm.Define(
			"ChoiceRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
			hsm.State("positive"),
			hsm.State("non_positive"),
			hsm.Transition(
				hsm.On(chooseEvent),
				hsm.Source("idle"),
				hsm.Target(hsm.Choice(
					"decision",
					hsm.Transition(
						hsm.Target("/ChoiceRulesHSM/positive"),
						hsm.Guard(func(ctx context.Context, sm *THSM, event hsm.Event) bool {
							value, ok := event.Data.(int)
							return ok && value > 0
						}),
					),
					hsm.Transition(hsm.Target("/ChoiceRulesHSM/non_positive")),
				)),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "choice dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: chooseEvent.Name, Data: 7}))
		assertRuleState(t, sm.State(), "/ChoiceRulesHSM/positive")
	})

	t.Run("HSM22/external_code_uses_attributes_instead_of_internal_context", func(t *testing.T) {
		result := runRuntimeRuleAttributeScenario(t, 2)
		if result.initial != 1 || result.final != 2 {
			t.Fatalf("expected attribute API values 1 -> 2, got %d -> %d", result.initial, result.final)
		}
		assertRuleState(t, result.state, "/AttributeRulesHSM/changed")
	})

	t.Run("HSM23/get_and_set_expose_stateful_machine_data", func(t *testing.T) {
		result := runRuntimeRuleAttributeScenario(t, 2)
		if result.initial != 1 {
			t.Fatalf("expected default attribute value 1, got %d", result.initial)
		}
		if result.final != 2 {
			t.Fatalf("expected updated attribute value 2, got %d", result.final)
		}
	})

	t.Run("HSM24/context_carries_requests_not_durable_machine_state", func(t *testing.T) {
		reply, handled := runContextPayloadScenario(t)
		if reply != "req-7:approve" {
			t.Fatalf("expected transient request context to flow through response, got %q", reply)
		}
		if !handled {
			t.Fatal("expected durable handled flag to persist as an attribute")
		}
	})

	t.Run("HSM25/external_input_arrives_via_events_or_attributes", func(t *testing.T) {
		state, message := runExternalInputScenario(t)
		assertRuleState(t, state, "/ExternalInputRulesHSM/updated")
		if message != "hello" {
			t.Fatalf("expected external input to arrive through supported APIs, got %q", message)
		}
	})

	t.Run("HSM26/event_payload_coordinates_request_response", func(t *testing.T) {
		reply, _ := runContextPayloadScenario(t)
		if reply != "req-7:approve" {
			t.Fatalf("expected reply channel coordination via event payload, got %q", reply)
		}
	})

	t.Run("HSM27/run_to_completion_serializes_context_access_without_mutexes", func(t *testing.T) {
		const concurrentDispatches = 4

		workEvent := hsm.Event{Name: "work"}
		trace := &runtimeTrace{}
		barrier := newStartBarrier(concurrentDispatches)
		firstStarted := make(chan struct{})
		releaseFirst := make(chan struct{})
		waiters := make(chan hsm.Completion, concurrentDispatches)
		var firstOnce sync.Once
		var inFlight atomic.Int32
		var maxInFlight atomic.Int32

		model := hsm.Define(
			"ConcurrentRuleHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(
					hsm.On(workEvent),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						active := inFlight.Add(1)
						storeMaxInt32(&maxInFlight, active)
						label := fmt.Sprint(event.Data)
						trace.record("start:" + label)
						firstOnce.Do(func() {
							close(firstStarted)
							<-releaseFirst
						})
						trace.record("end:" + label)
						inFlight.Add(-1)
					}),
				),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		for dispatch := 0; dispatch < concurrentDispatches; dispatch++ {
			dispatch := dispatch
			go func() {
				barrier.participantReadyAndWait()
				waiters <- hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: workEvent.Name, Data: dispatch})
			}()
		}

		barrier.release(t, "concurrent callers ready")
		awaitWaiter(t, "first serialized effect start", firstStarted)

		collected := make([]hsm.Completion, 0, concurrentDispatches)
		for len(collected) < concurrentDispatches {
			collected = append(collected, <-waiters)
		}

		close(releaseFirst)
		for index, waiter := range collected {
			awaitWaiter(t, fmt.Sprintf("serialized dispatch waiter %d", index+1), waiter)
		}

		if maxInFlight.Load() != 1 {
			t.Fatalf("expected serialized processing, observed max parallelism %d", maxInFlight.Load())
		}

		entries := trace.snapshot()
		if len(entries) != concurrentDispatches*2 {
			t.Fatalf("expected %d trace entries, got %d: %v", concurrentDispatches*2, len(entries), entries)
		}
		for index := 0; index < len(entries); index += 2 {
			start, ok := splitTrace(entries[index], "start:")
			if !ok {
				t.Fatalf("expected trace entry %d to be a start marker, got %q", index, entries[index])
			}
			end, ok := splitTrace(entries[index+1], "end:")
			if !ok {
				t.Fatalf("expected trace entry %d to be an end marker, got %q", index+1, entries[index+1])
			}
			if start != end {
				t.Fatalf("expected paired start/end labels, got %q and %q", entries[index], entries[index+1])
			}
		}
	})

	t.Run("HSM28/events_and_attributes_replace_shared_memory_coordination", func(t *testing.T) {
		state, message := runExternalInputScenario(t)
		assertRuleState(t, state, "/ExternalInputRulesHSM/updated")
		if message != "hello" {
			t.Fatalf("expected coordinated value to arrive through event+attribute flow, got %q", message)
		}
	})

	t.Run("HSM29/set_emits_attribute_reaction_when_value_changes", func(t *testing.T) {
		result := runRuntimeRuleAttributeScenario(t, 2)
		if !result.observed {
			t.Fatal("expected changed attribute set to emit an OnSet reaction")
		}
		if result.change.Name != "/AttributeRulesHSM/count" {
			t.Fatalf("expected attribute change name %q, got %q", "/AttributeRulesHSM/count", result.change.Name)
		}
		if result.change.Old != 1 || result.change.New != 2 {
			t.Fatalf("expected attribute change payload 1 -> 2, got %#v -> %#v", result.change.Old, result.change.New)
		}
		if result.kind != hsm.ChangeEventKind {
			t.Fatalf("expected change event kind %d, got %d", hsm.ChangeEventKind, result.kind)
		}
	})

	t.Run("HSM30/set_same_value_does_not_emit_onset", func(t *testing.T) {
		result := runRuntimeRuleAttributeScenario(t, 1)
		if result.observed {
			t.Fatal("expected identical Set to avoid OnSet emission")
		}
		assertRuleState(t, result.state, "/AttributeRulesHSM/idle")
	})

	t.Run("HSM31/onset_drives_attribute_transition", func(t *testing.T) {
		result := runRuntimeRuleAttributeScenario(t, 2)
		assertRuleState(t, result.state, "/AttributeRulesHSM/changed")
	})

	t.Run("HSM35/call_triggers_operation_protocol_without_fake_events", func(t *testing.T) {
		callDataCh := make(chan hsm.CallData, 1)
		kindCh := make(chan uint64, 1)
		sourceCh := make(chan string, 1)

		model := hsm.Define(
			"CallRulesHSM",
			hsm.Operation("do", func(ctx context.Context, sm *CallOrderHSM, a int, b string) string {
				sm.record("op")
				return fmt.Sprintf("%d:%s", a, b)
			}),
			hsm.State("idle",
				hsm.Transition(
					hsm.OnCall("do"),
					hsm.Target("../called"),
					hsm.Effect(func(ctx context.Context, sm *CallOrderHSM, event hsm.Event) {
						sm.record("effect")
						callDataCh <- event.Data.(hsm.CallData)
						kindCh <- event.Kind
						sourceCh <- event.Source
					}),
				),
			),
			hsm.State("called"),
			hsm.Initial(hsm.Target("idle")),
		)

		sm := hsm.Started(context.Background(), &CallOrderHSM{}, &model)
		enteredCalled := hsm.AfterEntry(sm.Context(), sm, "/CallRulesHSM/called")
		result, err := hsm.Call(context.Background(), sm, "do", 1, "two")
		if err != nil {
			t.Fatalf("expected call to succeed, got %v", err)
		}
		if result != "1:two" {
			t.Fatalf("expected operation result %q, got %#v", "1:two", result)
		}
		awaitWaiter(t, "call transition entry", enteredCalled)

		callData := <-callDataCh
		if callData.Name != "/CallRulesHSM/do" {
			t.Fatalf("expected call data name %q, got %q", "/CallRulesHSM/do", callData.Name)
		}
		if len(callData.Args) != 2 || callData.Args[0] != 1 || callData.Args[1] != "two" {
			t.Fatalf("unexpected call args %#v", callData.Args)
		}
		if kind := <-kindCh; kind != hsm.CallEventKind {
			t.Fatalf("expected call event kind %d, got %d", hsm.CallEventKind, kind)
		}
		if source := <-sourceCh; source != "/CallRulesHSM/do" {
			t.Fatalf("expected call event source %q, got %q", "/CallRulesHSM/do", source)
		}
		assertRuleState(t, sm.State(), "/CallRulesHSM/called")
		order := sm.orderSnapshot()
		if len(order) != 2 || !slices.Contains(order, "op") || !slices.Contains(order, "effect") {
			t.Fatalf("expected call protocol trace to include operation and effect, got %v", order)
		}
	})

	t.Run("HSM36/hsm_context_scopes_fire_and_forget_dispatch_to_machine_lifetime", func(t *testing.T) {
		doneEvent := hsm.Event{Name: "done"}
		trace := &runtimeTrace{}
		model := hsm.Define(
			"MachineContextDispatchRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(
					hsm.On(hsm.Event{Name: "kick"}),
					hsm.Target("../working"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("kick")
						hsm.Dispatch(sm.Context(), nil, doneEvent)
					}),
				),
			),
			hsm.State("working",
				hsm.Transition(
					hsm.On(doneEvent),
					hsm.Target("../done"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("done")
					}),
				),
			),
			hsm.State("done"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		enteredDone := hsm.AfterEntry(sm.Context(), sm, "/MachineContextDispatchRulesHSM/done")
		awaitWaiter(t, "machine-context kick dispatch", hsm.Dispatch(context.Background(), sm, hsm.Event{Name: "kick"}))
		awaitWaiter(t, "machine-context done entry", enteredDone)
		assertRuleState(t, sm.State(), "/MachineContextDispatchRulesHSM/done")
		assertRuleTrace(t, trace.snapshot(), []string{"kick", "done"})
	})

	t.Run("HSM37/background_context_allows_dispatch_to_outlive_machine_lifetime", func(t *testing.T) {
		doneEvent := hsm.Event{Name: "done"}
		trace := &runtimeTrace{}
		model := hsm.Define(
			"BackgroundDispatchRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(
					hsm.On(hsm.Event{Name: "kick"}),
					hsm.Target("../working"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("kick")
						hsm.Dispatch(context.Background(), sm, doneEvent)
					}),
				),
			),
			hsm.State("working",
				hsm.Transition(
					hsm.On(doneEvent),
					hsm.Target("../done"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("done")
					}),
				),
			),
			hsm.State("done"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		enteredDone := hsm.AfterEntry(sm.Context(), sm, "/BackgroundDispatchRulesHSM/done")
		awaitWaiter(t, "background-context kick dispatch", hsm.Dispatch(context.Background(), sm, hsm.Event{Name: "kick"}))
		awaitWaiter(t, "background-context done entry", enteredDone)
		assertRuleState(t, sm.State(), "/BackgroundDispatchRulesHSM/done")
		assertRuleTrace(t, trace.snapshot(), []string{"kick", "done"})
	})

	t.Run("HSM38/transient_behavior_context_is_used_only_for_intentional_cancellation", func(t *testing.T) {
		activityStarted := make(chan struct{})
		activityCancelled := make(chan struct{})

		model := hsm.Define(
			"TransientContextRulesHSM",
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
		awaitWaiter(t, "transient-context activity start", activityStarted)
		stopDone := hsm.Stop(context.Background(), sm)
		awaitWaiter(t, "transient-context activity cancellation", activityCancelled)
		awaitWaiter(t, "transient-context stop completion", stopDone)
	})

	t.Run("HSM39/id_is_unique_and_name_is_model_name", func(t *testing.T) {
		model := hsm.Define(
			"IdentityRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
		)

		alpha := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{ID: "alpha"})
		bravo := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{ID: "bravo"})

		if hsm.ID(alpha) == hsm.ID(bravo) {
			t.Fatal("expected started instances to expose unique IDs")
		}
		if hsm.Name(alpha) != "IdentityRulesHSM" {
			t.Fatalf("expected model name %q, got %q", "IdentityRulesHSM", hsm.Name(alpha))
		}
	})

	t.Run("HSM40/name_is_not_used_as_unique_instance_identifier", func(t *testing.T) {
		model := hsm.Define(
			"IdentityRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
		)

		alpha := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{ID: "alpha"})
		bravo := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{ID: "bravo"})

		if hsm.Name(alpha) != hsm.Name(bravo) {
			t.Fatalf("expected same model name for both instances, got %q and %q", hsm.Name(alpha), hsm.Name(bravo))
		}
		if hsm.ID(alpha) == hsm.ID(bravo) {
			t.Fatal("expected unique IDs when model names match")
		}
	})

	t.Run("HSM41/qualified_name_and_state_report_different_paths", func(t *testing.T) {
		model := hsm.Define(
			"IdentityRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model, hsm.Config{ID: "alpha"})
		if hsm.QualifiedName(sm) != "/IdentityRulesHSM" {
			t.Fatalf("expected qualified model path %q, got %q", "/IdentityRulesHSM", hsm.QualifiedName(sm))
		}
		assertRuleState(t, sm.State(), "/IdentityRulesHSM/idle")
	})

	t.Run("HSM43/negative_durations_do_not_fire_time_based_triggers", func(t *testing.T) {
		harness := newDeterministicClockHarness()
		model := hsm.Define(
			"NegativeAfterRulesHSM",
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
		enteredBar := hsm.AfterEntry(sm.Context(), sm, "/NegativeAfterRulesHSM/bar")

		harness.assertNoRegistration(t, "negative duration timer")
		assertWaiterPending(t, "negative duration should not enter bar", enteredBar)
		if harness.registrationCount() != 0 {
			t.Fatalf("expected no timer registrations, got %d", harness.registrationCount())
		}
		assertRuleState(t, sm.State(), "/NegativeAfterRulesHSM/foo")
	})

	t.Run("HSM44/activity_reserved_for_long_running_work", func(t *testing.T) {
		activityStarted := make(chan struct{})
		releaseActivity := make(chan struct{})
		activityFinished := make(chan struct{})

		model := hsm.Define(
			"LongRunningActivityRulesHSM",
			hsm.Initial(hsm.Target("running")),
			hsm.State("running",
				hsm.Activity(func(ctx context.Context, sm *THSM, event hsm.Event) {
					close(activityStarted)
					<-releaseActivity
					close(activityFinished)
				}),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "long-running activity start", activityStarted)
		assertWaiterPending(t, "long-running activity remains active", activityFinished)
		close(releaseActivity)
		awaitWaiter(t, "long-running activity finish", activityFinished)
		assertRuleState(t, sm.State(), "/LongRunningActivityRulesHSM/running")
	})

	t.Run("HSM45/short_synchronous_work_stays_in_entry_exit_or_effect", func(t *testing.T) {
		advanceEvent := hsm.Event{Name: "advance"}
		trace := &runtimeTrace{}

		model := hsm.Define(
			"SyncWorkRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Entry(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("idle.entry")
				}),
				hsm.Exit(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("idle.exit")
				}),
				hsm.Transition(
					hsm.On(advanceEvent),
					hsm.Target("../done"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("advance.effect")
					}),
				),
			),
			hsm.State("done",
				hsm.Entry(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("done.entry")
				}),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "sync-work dispatch", hsm.Dispatch(sm.Context(), sm, advanceEvent))
		assertRuleState(t, sm.State(), "/SyncWorkRulesHSM/done")
		assertRuleTrace(t, trace.snapshot(), []string{"idle.entry", "idle.exit", "advance.effect", "done.entry"})
	})

	t.Run("HSM46/activity_respects_context_cancellation_on_exit", func(t *testing.T) {
		finishEvent := hsm.Event{Name: "finish"}
		activityStarted := make(chan struct{})
		activityCancelled := make(chan struct{})

		model := hsm.Define(
			"ActivityCancellationRulesHSM",
			hsm.Initial(hsm.Target("running")),
			hsm.State("running",
				hsm.Activity(func(ctx context.Context, sm *THSM, event hsm.Event) {
					close(activityStarted)
					<-ctx.Done()
					close(activityCancelled)
				}),
				hsm.Transition(hsm.On(finishEvent), hsm.Target("../done")),
			),
			hsm.State("done"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "activity-cancellation activity start", activityStarted)
		enteredDone := hsm.AfterEntry(sm.Context(), sm, "/ActivityCancellationRulesHSM/done")
		awaitWaiter(t, "activity-cancellation finish dispatch", hsm.Dispatch(sm.Context(), sm, finishEvent))
		awaitWaiter(t, "activity-cancellation done entry", enteredDone)
		awaitWaiter(t, "activity-cancellation ctx done", activityCancelled)
		assertRuleState(t, sm.State(), "/ActivityCancellationRulesHSM/done")
	})

	t.Run("HSM47/decompose_behavior_into_separate_state_units", func(t *testing.T) {
		validateEvent := hsm.Event{Name: "validate"}
		completeEvent := hsm.Event{Name: "complete"}
		trace := &runtimeTrace{}

		model := hsm.Define(
			"DecomposedRulesHSM",
			hsm.Initial(hsm.Target("validate")),
			hsm.State("validate",
				hsm.Entry(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("validate")
				}),
				hsm.Transition(hsm.On(validateEvent), hsm.Target("../process")),
			),
			hsm.State("process",
				hsm.Entry(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("process")
				}),
				hsm.Transition(hsm.On(completeEvent), hsm.Target("../done")),
			),
			hsm.State("done",
				hsm.Entry(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("done")
				}),
			),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "decomposed validate dispatch", hsm.Dispatch(sm.Context(), sm, validateEvent))
		assertRuleState(t, sm.State(), "/DecomposedRulesHSM/process")
		awaitWaiter(t, "decomposed complete dispatch", hsm.Dispatch(sm.Context(), sm, completeEvent))
		assertRuleState(t, sm.State(), "/DecomposedRulesHSM/done")
		assertRuleTrace(t, trace.snapshot(), []string{"validate", "process", "done"})
	})

	t.Run("HSM48/deferred_events_replay_after_state_exit", func(t *testing.T) {
		activateEvent := hsm.Event{Name: "activate"}
		resumeEvent := hsm.Event{Name: "resume"}
		trace := &runtimeTrace{}

		model := hsm.Define(
			"DeferredReplayRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Defer(resumeEvent),
				hsm.Transition(
					hsm.On(activateEvent),
					hsm.Target("../ready"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("activate")
					}),
				),
			),
			hsm.State("ready",
				hsm.Transition(
					hsm.On(resumeEvent),
					hsm.Target("../done"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						trace.record("resume")
					}),
				),
			),
			hsm.State("done"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		resumeProcessed := hsm.AfterProcess(sm.Context(), sm, resumeEvent)
		enteredReady := hsm.AfterEntry(sm.Context(), sm, "/DeferredReplayRulesHSM/ready")
		enteredDone := hsm.AfterEntry(sm.Context(), sm, "/DeferredReplayRulesHSM/done")

		awaitWaiter(t, "deferred resume dispatch", hsm.Dispatch(sm.Context(), sm, resumeEvent))
		assertWaiterPending(t, "deferred resume should remain queued", resumeProcessed)

		awaitWaiter(t, "deferred activate dispatch", hsm.Dispatch(sm.Context(), sm, activateEvent))
		awaitWaiter(t, "deferred ready entry", enteredReady)
		awaitWaiter(t, "deferred replay processing", resumeProcessed)
		awaitWaiter(t, "deferred done entry", enteredDone)

		assertRuleState(t, sm.State(), "/DeferredReplayRulesHSM/done")
		assertRuleTrace(t, trace.snapshot(), []string{"activate", "resume"})
	})

	t.Run("HSM48/deferred_events_replay_after_ordinary_composite_exit", func(t *testing.T) {
		leaveEvent := hsm.Event{Name: "leave"}
		resumeEvent := hsm.Event{Name: "resume"}
		model := hsm.Define(
			"DeferredCompositeReplayRulesHSM",
			hsm.Initial(hsm.Target("parent")),
			hsm.State("parent",
				hsm.Initial(hsm.Target("idle")),
				hsm.State("idle",
					hsm.Defer(resumeEvent),
				),
				hsm.Transition(
					hsm.On(leaveEvent),
					hsm.Target("../done"),
				),
			),
			hsm.State("done",
				hsm.Transition(
					hsm.On(resumeEvent),
					hsm.Target("../handled"),
				),
			),
			hsm.State("handled"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		resumeProcessed := hsm.AfterProcess(sm.Context(), sm, resumeEvent)
		handledEntered := hsm.AfterEntry(sm.Context(), sm, "/DeferredCompositeReplayRulesHSM/handled")

		awaitWaiter(t, "deferred resume dispatch", hsm.Dispatch(sm.Context(), sm, resumeEvent))
		assertWaiterPending(t, "resume should be deferred under ordinary composite", resumeProcessed)
		awaitWaiter(t, "leave composite dispatch", hsm.Dispatch(sm.Context(), sm, leaveEvent))
		awaitWaiter(t, "resume replay after composite exit", resumeProcessed)
		awaitWaiter(t, "handled entry", handledEntered)
		assertRuleState(t, sm.State(), "/DeferredCompositeReplayRulesHSM/handled")
	})

	t.Run("HSM51/history_fallback_transition_overrides_parent_initial_on_first_reentry", func(t *testing.T) {
		reenterEvent := hsm.Event{Name: "reenter"}
		model := hsm.Define(
			"HistoryFallbackRulesHSM",
			hsm.State("A",
				hsm.State("A1"),
				hsm.State("A2"),
				hsm.Initial(hsm.Target("A1")),
				hsm.ShallowHistory("shallow", hsm.Transition(hsm.Target("A2"))),
			),
			hsm.State("B"),
			hsm.Transition(hsm.On(reenterEvent), hsm.Source("B"), hsm.Target("A/shallow")),
			hsm.Initial(hsm.Target("B")),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		awaitWaiter(t, "history fallback dispatch", hsm.Dispatch(sm.Context(), sm, reenterEvent))
		assertRuleState(t, sm.State(), "/HistoryFallbackRulesHSM/A/A2")
	})

	t.Run("HSM52/any_event_guards_filter_internal_lifecycle_events", func(t *testing.T) {
		trace := &runtimeTrace{}
		model := defineAnyEventRulesModel(trace)
		sm := hsm.Started(context.Background(), &THSM{}, &model)

		internalEvent := hsm.Event{Name: "internal", Kind: hsm.CompletionEventKind}
		internalProcessed := hsm.AfterProcess(sm.Context(), sm, internalEvent)

		awaitWaiter(t, "internal lifecycle dispatch", hsm.Dispatch(sm.Context(), sm, internalEvent))
		awaitWaiter(t, "internal lifecycle processed", internalProcessed)
		assertRuleState(t, sm.State(), "/AnyEventRulesHSM/ready")
		assertRuleTrace(t, trace.snapshot(), nil)
	})

	t.Run("HSM53/wait_for_returned_channels_before_post_transition_assertions", func(t *testing.T) {
		advanceEvent := hsm.Event{Name: "advance"}
		effectStarted := make(chan struct{})
		releaseEffect := make(chan struct{})

		model := hsm.Define(
			"WaiterContractRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(
					hsm.On(advanceEvent),
					hsm.Target("../done"),
					hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
						close(effectStarted)
						<-releaseEffect
					}),
				),
			),
			hsm.State("done"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		enteredDone := hsm.AfterEntry(sm.Context(), sm, "/WaiterContractRulesHSM/done")
		done := hsm.Dispatch(sm.Context(), sm, advanceEvent)
		awaitWaiter(t, "waiter-contract effect start", effectStarted)
		assertWaiterPending(t, "waiter-contract dispatch before release", done)

		close(releaseEffect)
		awaitWaiter(t, "waiter-contract dispatch completion", done)
		awaitWaiter(t, "waiter-contract done entry", enteredDone)
		assertRuleState(t, sm.State(), "/WaiterContractRulesHSM/done")
	})

	t.Run("HSM54/after_hooks_are_test_observers_not_production_sync", func(t *testing.T) {
		advanceEvent := hsm.Event{Name: "advance"}
		model := hsm.Define(
			"ObserverHookRulesHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Transition(hsm.On(advanceEvent), hsm.Target("../done")),
			),
			hsm.State("done"),
		)

		sm := hsm.Started(context.Background(), &THSM{}, &model)
		enteredDone := hsm.AfterEntry(sm.Context(), sm, "/ObserverHookRulesHSM/done")
		awaitWaiter(t, "observer-hook dispatch waiter", hsm.Dispatch(sm.Context(), sm, advanceEvent))
		awaitWaiter(t, "observer-hook after-entry waiter", enteredDone)
		assertRuleState(t, sm.State(), "/ObserverHookRulesHSM/done")
	})
}

func defineAnyEventRulesModel(trace *runtimeTrace) hsm.FinalizedModel {
	specialEvent := hsm.Event{Name: "special"}
	guardedEvent := hsm.Event{Name: "guarded"}
	resetEvent := hsm.Event{Name: "reset"}

	return hsm.Define(
		"AnyEventRulesHSM",
		hsm.Initial(hsm.Target("ready")),
		hsm.State("ready",
			hsm.Transition(
				hsm.On(specialEvent),
				hsm.Target("../special"),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("special")
				}),
			),
			hsm.Transition(
				hsm.On(guardedEvent),
				hsm.Target("../special"),
				hsm.Guard(func(ctx context.Context, sm *THSM, event hsm.Event) bool {
					return false
				}),
			),
			hsm.Transition(
				hsm.On(guardedEvent),
				hsm.Target("../guarded"),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("guard_second")
				}),
			),
			hsm.Transition(
				hsm.On(hsm.AnyEvent),
				hsm.Target("../fallback"),
				hsm.Guard(func(ctx context.Context, sm *THSM, event hsm.Event) bool {
					return event.Kind == hsm.EventKind
				}),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					trace.record("fallback:" + event.Name)
				}),
			),
		),
		hsm.State("special",
			hsm.Transition(hsm.On(resetEvent), hsm.Target("../ready")),
		),
		hsm.State("guarded",
			hsm.Transition(hsm.On(resetEvent), hsm.Target("../ready")),
		),
		hsm.State("fallback",
			hsm.Transition(hsm.On(resetEvent), hsm.Target("../ready")),
		),
	)
}

func runRuntimeRuleAttributeScenario(t *testing.T, next int) runtimeRuleAttributeResult {
	t.Helper()

	changeCh := make(chan hsm.AttributeChange, 1)
	kindCh := make(chan uint64, 1)
	model := hsm.Define(
		"AttributeRulesHSM",
		hsm.Attribute("count", 1),
		hsm.State("idle",
			hsm.Transition(
				hsm.OnSet("count"),
				hsm.Target("../changed"),
				hsm.Effect(func(ctx context.Context, sm *AttrHSM, event hsm.Event) {
					changeCh <- event.Data.(hsm.AttributeChange)
					kindCh <- event.Kind
				}),
			),
		),
		hsm.State("changed"),
		hsm.Initial(hsm.Target("idle")),
	)

	sm := hsm.Started(context.Background(), &AttrHSM{}, &model)
	initialAny, ok := hsm.Get(context.Background(), sm, "count")
	if !ok {
		t.Fatal("expected default count attribute to exist")
	}
	initial, ok := initialAny.(int)
	if !ok {
		t.Fatalf("expected default count attribute to be int, got %T", initialAny)
	}

	enteredChanged := hsm.AfterEntry(sm.Context(), sm, "/AttributeRulesHSM/changed")
	awaitWaiter(t, "attribute set waiter", hsm.Set(context.Background(), sm, "count", next))

	result := runtimeRuleAttributeResult{
		initial: initial,
		state:   sm.State(),
	}
	finalAny, ok := hsm.Get(context.Background(), sm, "count")
	if !ok {
		t.Fatal("expected updated count attribute to exist")
	}
	result.final, ok = finalAny.(int)
	if !ok {
		t.Fatalf("expected updated count attribute to be int, got %T", finalAny)
	}

	if next == initial {
		assertWaiterPending(t, "same-value attribute transition", enteredChanged)
		return result
	}

	awaitWaiter(t, "changed-state entry", enteredChanged)
	result.state = sm.State()
	result.observed = true
	result.change = <-changeCh
	result.kind = <-kindCh
	return result
}

func runContextPayloadScenario(t *testing.T) (string, bool) {
	t.Helper()

	requestEvent := hsm.Event{Name: "request"}
	model := hsm.Define(
		"ContextPayloadRulesHSM",
		hsm.Attribute("handled", false),
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(
				hsm.On(requestEvent),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					requestID, _ := ctx.Value(runtimeRuleContextKey("request_id")).(string)
					payload := event.Data.(runtimeRuleRequest)
					payload.Reply <- requestID + ":" + payload.Value
					hsm.Set(ctx, sm, "handled", true)
				}),
			),
		),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	replyCh := make(chan string, 1)
	ctx := context.WithValue(context.Background(), runtimeRuleContextKey("request_id"), "req-7")
	awaitWaiter(t, "context-payload dispatch", hsm.Dispatch(ctx, sm, hsm.Event{
		Name: requestEvent.Name,
		Data: runtimeRuleRequest{
			Value: "approve",
			Reply: replyCh,
		},
	}))

	var reply string
	select {
	case reply = <-replyCh:
	case <-time.After(waiterDeadline):
		t.Fatal("timed out waiting for request/response payload reply")
	}

	handledAny, ok := hsm.Get(context.Background(), sm, "handled")
	if !ok {
		t.Fatal("expected handled attribute to exist")
	}
	handled, ok := handledAny.(bool)
	if !ok {
		t.Fatalf("expected handled attribute to be bool, got %T", handledAny)
	}
	return reply, handled
}

func runExternalInputScenario(t *testing.T) (string, string) {
	t.Helper()

	requestEvent := hsm.Event{Name: "request"}
	model := hsm.Define(
		"ExternalInputRulesHSM",
		hsm.Attribute("message", ""),
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(
				hsm.On(requestEvent),
				hsm.Effect(func(ctx context.Context, sm *THSM, event hsm.Event) {
					hsm.Set(ctx, sm, "message", event.Data.(string))
				}),
			),
			hsm.Transition(
				hsm.OnSet("message"),
				hsm.Target("../updated"),
			),
		),
		hsm.State("updated"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	enteredUpdated := hsm.AfterEntry(sm.Context(), sm, "/ExternalInputRulesHSM/updated")
	awaitWaiter(t, "external-input dispatch", hsm.Dispatch(sm.Context(), sm, hsm.Event{Name: requestEvent.Name, Data: "hello"}))
	awaitWaiter(t, "external-input updated entry", enteredUpdated)

	messageAny, ok := hsm.Get(context.Background(), sm, "message")
	if !ok {
		t.Fatal("expected message attribute to exist")
	}
	message, ok := messageAny.(string)
	if !ok {
		t.Fatalf("expected message attribute to be string, got %T", messageAny)
	}
	return sm.State(), message
}

func assertRuleState(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
}

func assertRuleTrace(t *testing.T, got, want []string) {
	t.Helper()
	if want == nil {
		if len(got) != 0 {
			t.Fatalf("expected empty trace, got %v", got)
		}
		return
	}
	if !slices.Equal(got, want) {
		t.Fatalf("expected trace %v, got %v", want, got)
	}
}

func splitTrace(value, prefix string) (string, bool) {
	if len(value) < len(prefix) || value[:len(prefix)] != prefix {
		return "", false
	}
	return value[len(prefix):], true
}
