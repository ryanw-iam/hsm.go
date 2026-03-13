package hsm_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stateforward/hsm.go"
)

func TestRuntimeConcurrentDispatchSerializesProcessing(t *testing.T) {
	const (
		concurrentDispatches = 4
		repeatedRuns         = 8
	)

	workEvent := hsm.Event{Name: "work"}
	for run := 0; run < repeatedRuns; run++ {
		t.Run(fmt.Sprintf("run_%02d", run+1), func(t *testing.T) {
			trace := &runtimeTrace{}
			barrier := newStartBarrier(concurrentDispatches)
			firstStarted := make(chan struct{})
			releaseFirst := make(chan struct{})
			waiters := make(chan (<-chan struct{}), concurrentDispatches)
			var firstOnce sync.Once
			var inFlight atomic.Int32
			var maxInFlight atomic.Int32

			model := hsm.Define(
				"ConcurrentDispatchHSM",
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
					waiters <- hsm.Dispatch(sm.Context(), sm, hsm.Event{
						Name: workEvent.Name,
						Data: dispatch,
					})
				}()
			}

			barrier.release(t, "concurrent dispatch callers to be ready")
			awaitWaiter(t, "first concurrent work effect to start", firstStarted)

			collected := make([]<-chan struct{}, 0, concurrentDispatches)
			for len(collected) < concurrentDispatches {
				select {
				case waiter := <-waiters:
					collected = append(collected, waiter)
				}
			}

			close(releaseFirst)
			for index, waiter := range collected {
				awaitWaiter(t, fmt.Sprintf("concurrent dispatch waiter %d", index+1), waiter)
			}

			if sm.State() != "/ConcurrentDispatchHSM/idle" {
				t.Fatalf("expected idle state after concurrent dispatches, got %s", sm.State())
			}
			if maxInFlight.Load() != 1 {
				t.Fatalf("expected serialized processing, observed max parallelism %d", maxInFlight.Load())
			}

			entries := trace.snapshot()
			if len(entries) != concurrentDispatches*2 {
				t.Fatalf("expected %d trace entries, got %d: %v", concurrentDispatches*2, len(entries), entries)
			}
			for index := 0; index < len(entries); index += 2 {
				start, ok := strings.CutPrefix(entries[index], "start:")
				if !ok {
					t.Fatalf("expected trace entry %d to be a start marker, got %q", index, entries[index])
				}
				end, ok := strings.CutPrefix(entries[index+1], "end:")
				if !ok {
					t.Fatalf("expected trace entry %d to be an end marker, got %q", index+1, entries[index+1])
				}
				if start != end {
					t.Fatalf("expected start/end pair to match, got %q and %q", entries[index], entries[index+1])
				}
			}
		})
	}
}

func TestRuntimeDeferredEventsReplayAfterTransition(t *testing.T) {
	activateEvent := hsm.Event{Name: "activate"}
	resumeEvent := hsm.Event{Name: "resume"}
	trace := &runtimeTrace{}

	model := hsm.Define(
		"DeferredReplayHSM",
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
	readyEntered := hsm.AfterEntry(sm.Context(), sm, "/DeferredReplayHSM/ready")
	doneEntered := hsm.AfterEntry(sm.Context(), sm, "/DeferredReplayHSM/done")

	awaitWaiter(t, "deferred resume dispatch cycle", hsm.Dispatch(sm.Context(), sm, resumeEvent))
	assertWaiterPending(t, "deferred resume event processing before activation", resumeProcessed)
	if sm.State() != "/DeferredReplayHSM/idle" {
		t.Fatalf("expected idle state while resume is deferred, got %s", sm.State())
	}

	awaitWaiter(t, "activate dispatch cycle", hsm.Dispatch(sm.Context(), sm, activateEvent))
	awaitWaiter(t, "ready state entry", readyEntered)
	awaitWaiter(t, "deferred resume replay", resumeProcessed)
	awaitWaiter(t, "done state entry", doneEntered)

	if sm.State() != "/DeferredReplayHSM/done" {
		t.Fatalf("expected done state after deferred replay, got %s", sm.State())
	}

	expectedTrace := []string{"activate", "resume"}
	if entries := trace.snapshot(); !slices.Equal(entries, expectedTrace) {
		t.Fatalf("expected deferred replay trace %v, got %v", expectedTrace, entries)
	}
}

func TestRuntimeCompletionEventsPreemptQueuedEvents(t *testing.T) {
	triggerEvent := hsm.Event{Name: "trigger"}
	regularEvent := hsm.Event{Name: "regular"}
	priorityEvent := hsm.Event{Name: "priority", Kind: hsm.CompletionEventKind}
	trace := &runtimeTrace{}

	model := hsm.Define(
		"CompletionPriorityHSM",
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

	awaitWaiter(t, "trigger dispatch cycle", hsm.Dispatch(sm.Context(), sm, triggerEvent))
	awaitWaiter(t, "priority completion event processing", priorityProcessed)
	awaitWaiter(t, "regular event processing after priority transition", regularProcessed)

	if sm.State() != "/CompletionPriorityHSM/priority" {
		t.Fatalf("expected priority state after completion-event preemption, got %s", sm.State())
	}

	expectedTrace := []string{"trigger", "priority"}
	if entries := trace.snapshot(); !slices.Equal(entries, expectedTrace) {
		t.Fatalf("expected completion-priority trace %v, got %v", expectedTrace, entries)
	}
}

func storeMaxInt32(target *atomic.Int32, value int32) {
	for {
		current := target.Load()
		if value <= current {
			return
		}
		if target.CompareAndSwap(current, value) {
			return
		}
	}
}
