package hsm_test

import (
	"context"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

type activityLister interface {
	Activities() []string
}

func activityCount(t *testing.T, model hsm.FinalizedModel, state string) int {
	t.Helper()
	element := model.Members()[state]
	if element == nil {
		t.Fatalf("state %q not found", state)
	}
	lister, ok := element.(activityLister)
	if !ok {
		t.Fatalf("state %q does not expose activities", state)
	}
	return len(lister.Activities())
}

func TestModelPathMatchers(t *testing.T) {
	t.Run("match supports exact and wildcard paths", func(t *testing.T) {
		sample := "abc/def/abcde/xyz"
		if !hsm.Match(sample, "abc/*/a*/xyz") {
			t.Fatalf("expected %q to match wildcard pattern", sample)
		}
		if hsm.Match(sample, "abc/*/a*/") {
			t.Fatalf("did not expect %q to match trailing slash pattern", sample)
		}
	})

	t.Run("lca returns the nearest shared ancestor", func(t *testing.T) {
		if got, want := hsm.LCA("/foo/bar", "/foo/bar/baz"), "/foo/bar"; got != want {
			t.Fatalf("LCA = %q, want %q", got, want)
		}
		if got, want := hsm.LCA("/foo/bar/baz", "/foo/qux"), "/foo"; got != want {
			t.Fatalf("LCA = %q, want %q", got, want)
		}
	})

	t.Run("isancestor respects ancestry boundaries", func(t *testing.T) {
		if !hsm.IsAncestor("/foo/bar", "/foo/bar/baz") {
			t.Fatal("expected direct ancestry to match")
		}
		if hsm.IsAncestor("/foo/bar/baz", "/foo/bar") {
			t.Fatal("did not expect child to be ancestor of parent")
		}
		if hsm.IsAncestor("/foo/bar/baz", "/foo/bar/baz") {
			t.Fatal("did not expect a path to be its own ancestor")
		}
	})
}

func TestModelPathConstructionWithHistory(t *testing.T) {
	assertNoPanic(t, "define model with history helpers", func() {
		_ = hsm.Define(
			"ModelPathHSM",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle",
				hsm.Initial(hsm.Target("ready")),
				hsm.State("ready"),
				hsm.ShallowHistory("remember", hsm.Transition(hsm.Target("ready"))),
				hsm.DeepHistory("deep-remember", hsm.Transition(hsm.Target("ready"))),
			),
		)
	})
}

func TestRedefineReplaysBaseModelElements(t *testing.T) {
	advance := hsm.Event{Name: "advance"}
	base := hsm.Define(
		"BaseRedefineHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(hsm.On(advance), hsm.Target("../done")),
		),
		hsm.State("done"),
	)
	derived := hsm.Redefine(base, "DerivedRedefineHSM")
	sm := hsm.Started(context.Background(), &struct{ hsm.HSM }{}, &derived)
	<-hsm.Dispatch(context.Background(), sm, advance)
	if got, want := sm.State(), "/DerivedRedefineHSM/done"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}

	replayed := hsm.Redefine(base)
	replayedSM := hsm.Started(context.Background(), &struct{ hsm.HSM }{}, &replayed)
	<-hsm.Dispatch(context.Background(), replayedSM, advance)
	if got, want := replayedSM.State(), "/BaseRedefineHSM/done"; got != want {
		t.Fatalf("state after zero-partial Redefine = %q, want %q", got, want)
	}
}

func TestRedefineSameRootReplacesElements(t *testing.T) {
	advance := hsm.Event{Name: "advance"}
	base := hsm.Define(
		"SameRootRedefineHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(hsm.On(advance), hsm.Target("../base")),
		),
		hsm.State("base"),
		hsm.State("replacement"),
	)
	derived := hsm.Redefine(
		base,
		hsm.State("idle",
			hsm.Transition(hsm.On(advance), hsm.Target("../replacement")),
		),
	)

	sm := hsm.Started(context.Background(), &struct{ hsm.HSM }{}, &derived)
	<-hsm.Dispatch(context.Background(), sm, advance)
	if got, want := sm.State(), "/SameRootRedefineHSM/replacement"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestRedefineOverridesBaseOperationAndAttribute(t *testing.T) {
	base := hsm.Define(
		"BaseRedefineOverrideHSM",
		hsm.Attribute("flag", false),
		hsm.Operation("label", func() string { return "base" }),
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineOverrideHSM",
		hsm.Attribute("flag", true),
		hsm.Operation("label", func() string { return "derived" }),
	)
	sm := hsm.Started(context.Background(), &struct{ hsm.HSM }{}, &derived)

	value, ok := hsm.Get(context.Background(), sm, "flag")
	if !ok {
		t.Fatal("flag attribute not found")
	}
	if value != true {
		t.Fatalf("flag = %v, want true", value)
	}
	result, err := hsm.Call(context.Background(), sm, "label")
	if err != nil {
		t.Fatalf("Call(label) error = %v", err)
	}
	if result != "derived" {
		t.Fatalf("Call(label) = %v, want derived", result)
	}
}

func TestRedefineReplacesBaseTransition(t *testing.T) {
	advance := hsm.Event{Name: "advance"}
	base := hsm.Define(
		"BaseRedefineTransitionHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
		hsm.State("base"),
		hsm.State("derived"),
		hsm.Transition(
			"go",
			hsm.On(advance),
			hsm.Source("idle"),
			hsm.Target("base"),
		),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineTransitionHSM",
		hsm.Transition(
			"go",
			hsm.On(advance),
			hsm.Source("idle"),
			hsm.Target("derived"),
		),
	)

	sm := hsm.Started(context.Background(), &struct{ hsm.HSM }{}, &derived)
	<-hsm.Dispatch(context.Background(), sm, advance)
	if got, want := sm.State(), "/DerivedRedefineTransitionHSM/derived"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestRedefineReplacesBaseInitial(t *testing.T) {
	base := hsm.Define(
		"BaseRedefineInitialHSM",
		hsm.Initial(hsm.Target("base")),
		hsm.State("base"),
		hsm.State("derived"),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineInitialHSM",
		hsm.Initial(hsm.Target("derived")),
	)

	sm := hsm.Started(context.Background(), &struct{ hsm.HSM }{}, &derived)
	if got, want := sm.State(), "/DerivedRedefineInitialHSM/derived"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestRedefineReplacingGeneratedTriggerRemovesStaleActivity(t *testing.T) {
	advance := hsm.Event{Name: "advance"}
	base := hsm.Define(
		"BaseRedefineGeneratedTriggerHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
		hsm.State("stale"),
		hsm.State("replacement"),
		hsm.Transition(
			"timer",
			hsm.After(func(context.Context, *THSM, hsm.Event) time.Duration {
				return time.Hour
			}),
			hsm.Source("idle"),
			hsm.Target("stale"),
		),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineGeneratedTriggerHSM",
		hsm.Transition(
			"timer",
			hsm.On(advance),
			hsm.Source("idle"),
			hsm.Target("replacement"),
		),
	)

	if got := activityCount(t, derived, "/DerivedRedefineGeneratedTriggerHSM/idle"); got != 0 {
		t.Fatalf("generated activity count = %d, want 0", got)
	}
	sm := hsm.Started(context.Background(), &THSM{}, &derived)
	<-hsm.Dispatch(context.Background(), sm, advance)
	if got, want := sm.State(), "/DerivedRedefineGeneratedTriggerHSM/replacement"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestRedefineReplacingGeneratedTriggerDeduplicatesActivity(t *testing.T) {
	base := hsm.Define(
		"BaseRedefineGeneratedTriggerDedupHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle"),
		hsm.State("base"),
		hsm.State("replacement"),
		hsm.Transition(
			"timer",
			hsm.After(func(context.Context, *THSM, hsm.Event) time.Duration {
				return time.Hour
			}),
			hsm.Source("idle"),
			hsm.Target("base"),
		),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineGeneratedTriggerDedupHSM",
		hsm.Transition(
			"timer",
			hsm.After(func(context.Context, *THSM, hsm.Event) time.Duration {
				return 2 * time.Hour
			}),
			hsm.Source("idle"),
			hsm.Target("replacement"),
		),
	)

	if got := activityCount(t, derived, "/DerivedRedefineGeneratedTriggerDedupHSM/idle"); got != 1 {
		t.Fatalf("generated activity count = %d, want 1", got)
	}
}

func TestRedefineReplacingDelayedEntryPointSelector(t *testing.T) {
	goEvent := hsm.Event{Name: "go"}
	child := hsm.Define(
		"RedefineDelayedEntryChild",
		hsm.EntryPoint("warm", hsm.Target("ready")),
		hsm.Initial(hsm.Target("ready")),
		hsm.State("ready"),
	)
	base := hsm.Define(
		"BaseRedefineDelayedEntryHSM",
		hsm.Initial(hsm.Target("outside")),
		hsm.State("outside",
			hsm.Transition(
				"enter_drive",
				hsm.On(goEvent),
				hsm.Target("../drive"),
				hsm.EntryPoint("warm"),
			),
		),
		hsm.SubmachineState("drive", child),
		hsm.State("replacement"),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineDelayedEntryHSM",
		hsm.State("outside",
			hsm.Transition(
				"enter_drive",
				hsm.On(goEvent),
				hsm.Target("../replacement"),
			),
		),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &derived)
	<-hsm.Dispatch(context.Background(), sm, goEvent)
	if got, want := sm.State(), "/DerivedRedefineDelayedEntryHSM/replacement"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestSubmachineRebasesAbsoluteChildRootReferences(t *testing.T) {
	finish := hsm.Event{Name: "finish"}
	reset := hsm.Event{Name: "reset"}
	child := hsm.Define(
		"AbsoluteChildRootHSM",
		hsm.Initial(hsm.Target("/AbsoluteChildRootHSM/idle")),
		hsm.State("idle",
			hsm.Transition(hsm.On(finish), hsm.Target("/AbsoluteChildRootHSM/done")),
		),
		hsm.State("done"),
		hsm.Transition(
			hsm.On(reset),
			hsm.Source("/AbsoluteChildRootHSM/done"),
			hsm.Target("/AbsoluteChildRootHSM/idle"),
		),
	)
	model := hsm.Define(
		"SubmachineAbsoluteRebaseParent",
		hsm.Initial(hsm.Target("drive")),
		hsm.SubmachineState("drive", child),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	if got, want := sm.State(), "/SubmachineAbsoluteRebaseParent/drive/idle"; got != want {
		t.Fatalf("initial state = %q, want %q", got, want)
	}
	<-hsm.Dispatch(context.Background(), sm, finish)
	if got, want := sm.State(), "/SubmachineAbsoluteRebaseParent/drive/done"; got != want {
		t.Fatalf("finish state = %q, want %q", got, want)
	}
	<-hsm.Dispatch(context.Background(), sm, reset)
	if got, want := sm.State(), "/SubmachineAbsoluteRebaseParent/drive/idle"; got != want {
		t.Fatalf("reset state = %q, want %q", got, want)
	}
}

func TestRedefineReplacesConnectionPoints(t *testing.T) {
	finish := hsm.Event{Name: "finish"}
	baseChild := hsm.Define(
		"BaseRedefineConnectionChild",
		hsm.EntryPoint("hot", hsm.Target("cold")),
		hsm.ExitPoint("done"),
		hsm.Initial(hsm.Target("cold")),
		hsm.State("cold"),
		hsm.State("warm",
			hsm.Transition(hsm.On(finish), hsm.Target("../done")),
		),
	)
	derivedChild := hsm.Redefine(
		baseChild,
		"DerivedRedefineConnectionChild",
		hsm.EntryPoint("hot", hsm.Target("warm")),
		hsm.ExitPoint("done"),
	)
	model := hsm.Define(
		"RedefineConnectionParent",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Transition(hsm.On(hsm.Event{Name: "start"}), hsm.Target("../drive"), hsm.EntryPoint("hot")),
		),
		hsm.SubmachineState("drive", derivedChild,
			hsm.Transition(hsm.ExitPoint("done"), hsm.Target("../complete")),
		),
		hsm.State("complete"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &model)
	<-hsm.Dispatch(context.Background(), sm, hsm.Event{Name: "start"})
	if got, want := sm.State(), "/RedefineConnectionParent/drive/warm"; got != want {
		t.Fatalf("entry-point state = %q, want %q", got, want)
	}
	<-hsm.Dispatch(context.Background(), sm, finish)
	if got, want := sm.State(), "/RedefineConnectionParent/complete"; got != want {
		t.Fatalf("exit-point state = %q, want %q", got, want)
	}
}

func TestRedefineReplacesStateLikeVertexWithFinal(t *testing.T) {
	base := hsm.Define(
		"BaseRedefineFinalReplacementHSM",
		hsm.Initial(hsm.Target("work")),
		hsm.State("work",
			hsm.Transition(hsm.On(hsm.Event{Name: "go"}), hsm.Target("../other")),
		),
		hsm.State("other"),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineFinalReplacementHSM",
		hsm.Final("work"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &derived)
	if got, want := sm.State(), "/DerivedRedefineFinalReplacementHSM/work"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestRedefineReplacingStateRemovesRootOwnedSourceTransition(t *testing.T) {
	base := hsm.Define(
		"BaseRedefineRootSourceTransitionHSM",
		hsm.Initial(hsm.Target("work")),
		hsm.State("work"),
		hsm.State("other"),
		hsm.Transition(
			"work_to_other",
			hsm.On(hsm.Event{Name: "go"}),
			hsm.Source("work"),
			hsm.Target("other"),
		),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineRootSourceTransitionHSM",
		hsm.Final("work"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &derived)
	if got, want := sm.State(), "/DerivedRedefineRootSourceTransitionHSM/work"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestRedefineReplacingStateRemovesRootOwnedTransitionMembers(t *testing.T) {
	base := hsm.Define(
		"BaseRedefineRootSourceTransitionMembersHSM",
		hsm.Initial(hsm.Target("work")),
		hsm.State("work"),
		hsm.State("other"),
		hsm.Transition(
			"work_to_other",
			hsm.On("go"),
			hsm.Source("work"),
			hsm.Target("other"),
			hsm.Guard(func(context.Context, *THSM, hsm.Event) bool { return true }),
			hsm.Effect(func(context.Context, *THSM, hsm.Event) {}),
		),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineRootSourceTransitionMembersHSM",
		hsm.Final("work"),
	)

	for name := range derived.Members() {
		if name == "/DerivedRedefineRootSourceTransitionMembersHSM/work_to_other" ||
			hsm.IsAncestor("/DerivedRedefineRootSourceTransitionMembersHSM/work_to_other", name) {
			t.Fatalf("removed transition member survived redefine: %s", name)
		}
	}
}

func TestRedefineReplacingStateRemovesRootOwnedGeneratedTrigger(t *testing.T) {
	base := hsm.Define(
		"BaseRedefineRootSourceTimerHSM",
		hsm.Initial(hsm.Target("work")),
		hsm.State("work"),
		hsm.State("stale"),
		hsm.Transition(
			"timer",
			hsm.After(func(context.Context, *THSM, hsm.Event) time.Duration {
				return time.Hour
			}),
			hsm.Source("work"),
			hsm.Target("stale"),
		),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineRootSourceTimerHSM",
		hsm.Final("work"),
	)

	sm := hsm.Started(context.Background(), &THSM{}, &derived)
	if got, want := sm.State(), "/DerivedRedefineRootSourceTimerHSM/work"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestRedefineReplacesBaseStateSubtree(t *testing.T) {
	advance := hsm.Event{Name: "advance"}
	base := hsm.Define(
		"BaseRedefineStateHSM",
		hsm.Initial(hsm.Target("idle")),
		hsm.State("idle",
			hsm.Initial(hsm.Target("baseChild")),
			hsm.State("baseChild"),
			hsm.Transition(hsm.On(advance), hsm.Target("../baseDone")),
		),
		hsm.State("baseDone"),
		hsm.State("derivedDone"),
	)
	derived := hsm.Redefine(
		base,
		"DerivedRedefineStateHSM",
		hsm.State("idle",
			hsm.Initial(hsm.Target("derivedChild")),
			hsm.State("derivedChild"),
			hsm.Transition(hsm.On(advance), hsm.Target("../derivedDone")),
		),
	)

	sm := hsm.Started(context.Background(), &struct{ hsm.HSM }{}, &derived)
	if got, want := sm.State(), "/DerivedRedefineStateHSM/idle/derivedChild"; got != want {
		t.Fatalf("initial state = %q, want %q", got, want)
	}
	<-hsm.Dispatch(context.Background(), sm, advance)
	if got, want := sm.State(), "/DerivedRedefineStateHSM/derivedDone"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}
