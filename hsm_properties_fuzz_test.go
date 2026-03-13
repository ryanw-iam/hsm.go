package hsm_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stateforward/hsm.go"
)

func FuzzHSMEventScriptProperties(f *testing.F) {
	f.Add("begin|finish")
	f.Add("|begin||finish|finish")

	f.Fuzz(func(t *testing.T, script string) {
		checkHSMEventScriptProperties(t, script)
	})
}

func checkHSMEventScriptProperties(t *testing.T, script string) {
	t.Helper()

	events := normalizeHSMEventScript(t, script)
	reducer := newHSMEventScriptReducer()
	var doneEntries atomic.Int32

	model := hsm.Define(
		"EventScriptPropertyHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
		hsm.State("running"),
		hsm.State("done",
			hsm.Entry(func(context.Context, *THSM, hsm.Event) {
				doneEntries.Add(1)
			}),
		),
		hsm.Transition(hsm.On(hsm.Event{Name: "begin"}), hsm.Source("idle"), hsm.Target("running")),
		hsm.Transition(hsm.On(hsm.Event{Name: "finish"}), hsm.Source("running"), hsm.Target("done")),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	t.Cleanup(func() {
		awaitWaiter(t, "EventScriptPropertyHSM stop", hsm.Stop(context.Background(), sm))
	})

	if got, want := sm.State(), reducer.statePath(); got != want {
		t.Fatalf("initial state = %q, want %q", got, want)
	}

	for index, name := range events {
		event := hsm.Event{Name: name}
		processed := hsm.AfterProcess(sm.Context(), sm, event)
		dispatchDone := hsm.Dispatch(sm.Context(), sm, event)

		awaitWaiter(t, fmt.Sprintf("EventScriptPropertyHSM dispatch %d", index), dispatchDone)
		awaitWaiter(t, fmt.Sprintf("EventScriptPropertyHSM process %d", index), processed)

		reducer.apply(name)

		if got, want := sm.State(), reducer.statePath(); got != want {
			t.Fatalf("state after step %d (%q) = %q, want %q", index, name, got, want)
		}
		if got, want := int(doneEntries.Load()), reducer.doneEntries; got != want {
			t.Fatalf("done entries after step %d (%q) = %d, want %d", index, name, got, want)
		}
	}
}

type hsmEventScriptReducer struct {
	state       string
	doneEntries int
}

func newHSMEventScriptReducer() *hsmEventScriptReducer {
	return &hsmEventScriptReducer{state: "idle"}
}

func (r *hsmEventScriptReducer) apply(name string) {
	switch r.state {
	case "idle":
		if name == "begin" {
			r.state = "running"
		}
	case "running":
		if name == "finish" {
			r.state = "done"
			r.doneEntries++
		}
	}
}

func (r *hsmEventScriptReducer) statePath() string {
	return "/EventScriptPropertyHSM/" + r.state
}

func normalizeHSMEventScript(t *testing.T, script string) []string {
	t.Helper()

	if len(script) > 64 {
		t.Skip()
	}

	events := strings.Split(script, "|")
	if len(events) > 8 {
		t.Skip()
	}

	for _, event := range events {
		if len(event) > 12 {
			t.Skip()
		}
		for _, r := range event {
			if r < 'a' || r > 'z' {
				t.Skip()
			}
		}
	}

	return events
}
