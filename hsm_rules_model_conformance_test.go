package hsm_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

var modelRulesAdvanceEvent = hsm.Event{Name: "advance"}

func modelRulesNoBehavior(context.Context, *THSM, hsm.Event) {}

func modelRulesAlwaysGuard(context.Context, *THSM, hsm.Event) bool {
	return true
}

func modelRulesAfterDuration(context.Context, *THSM, hsm.Event) time.Duration {
	return time.Millisecond
}

func modelRulesWhenChannel(context.Context, *THSM, hsm.Event) <-chan struct{} {
	return make(chan struct{})
}

func TestModelRulesConformance(t *testing.T) {
	runRule := func(rule, name string, fn func(t *testing.T)) {
		t.Helper()
		t.Run(rule, func(t *testing.T) {
			t.Run(name, fn)
		})
	}

	runRule("HSM01", "top_level_initial_required", func(t *testing.T) {
		assertPanicContains(t, "HSM01/top_level_initial_required", "initial state is required for state machine", func() {
			hsm.Define(
				"BadTopLevelInitialMissing",
				hsm.State("idle"),
			)
		})
	})

	runRule("HSM03", "top_level_entry_actions_rejected", func(t *testing.T) {
		assertPanicContains(t, "HSM03/top_level_entry_actions_rejected", "entry actions are not allowed on top level state machine", func() {
			hsm.Define(
				"BadTopLevelEntryAction",
				hsm.Entry(modelRulesNoBehavior),
				hsm.State("idle"),
				hsm.Initial(hsm.Target("idle")),
			)
		})
	})

	runRule("HSM04", "top_level_exit_actions_rejected", func(t *testing.T) {
		assertPanicContains(t, "HSM04/top_level_exit_actions_rejected", "exit actions are not allowed on top level state machine", func() {
			hsm.Define(
				"BadTopLevelExitAction",
				hsm.Exit(modelRulesNoBehavior),
				hsm.State("idle"),
				hsm.Initial(hsm.Target("idle")),
			)
		})
	})

	runRule("HSM05", "initial_target_must_be_nested", func(t *testing.T) {
		assertPanicContains(t, "HSM05/initial_target_must_be_nested", "must target a nested state", func() {
			hsm.Define(
				"BadNestedInitialTarget",
				hsm.State("parent",
					hsm.State("child"),
					hsm.Initial(hsm.Target("/BadNestedInitialTarget/outside")),
				),
				hsm.State("outside"),
				hsm.Initial(hsm.Target("parent")),
			)
		})
	})

	runRule("HSM06", "initial_transition_cannot_have_guard", func(t *testing.T) {
		assertPanicContains(t, "HSM06/initial_transition_cannot_have_guard", "cannot have a guard", func() {
			hsm.Define(
				"BadInitialGuard",
				hsm.State("idle"),
				hsm.Initial(
					hsm.Target("idle"),
					hsm.Guard(modelRulesAlwaysGuard),
				),
			)
		})
	})

	runRule("HSM07", "initial_pseudostate_cannot_have_multiple_transitions", func(t *testing.T) {
		assertPanicContains(t, "HSM07/initial_pseudostate_cannot_have_multiple_transitions", "cannot have multiple transitions", func() {
			hsm.Define(
				"BadInitialMultipleTransitions",
				hsm.State("idle"),
				hsm.State("done"),
				hsm.Initial(
					hsm.Target("idle"),
					hsm.Transition(hsm.Target("done")),
				),
			)
		})
	})

	runRule("HSM08", "transition_members_must_be_declared_inside_transition", func(t *testing.T) {
		cases := []struct {
			name    string
			want    string
			element hsm.RedefinableElement
		}{
			{name: "source", want: "hsm.Source() must be called within a hsm.Transition()", element: hsm.Source("idle")},
			{name: "target", want: "Target() must be called within Transition()", element: hsm.Target("done")},
			{name: "on", want: "trigger must be called within a Transition", element: hsm.On(modelRulesAdvanceEvent)},
			{name: "onset", want: "OnSet() must be called within a Transition", element: hsm.OnSet("attr")},
			{name: "oncall", want: "OnCall() must be called within a Transition", element: hsm.OnCall("call")},
			{name: "after", want: "after must be called within a Transition", element: hsm.After(modelRulesAfterDuration)},
			{name: "every", want: "every must be called within a Transition", element: hsm.Every(modelRulesAfterDuration)},
			{name: "when", want: "when must be called within a Transition", element: hsm.When(modelRulesWhenChannel)},
			{name: "guard", want: "guard must be called within a Transition", element: hsm.Guard(modelRulesAlwaysGuard)},
			{name: "effect", want: "effect must be called within a Transition", element: hsm.Effect(modelRulesNoBehavior)},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "HSM08/"+tc.name, tc.want, func() {
					hsm.Define(
						fmt.Sprintf("BadTransitionMemberOwner_%s", tc.name),
						hsm.State("idle", tc.element),
						hsm.Initial(hsm.Target("idle")),
					)
				})
			})
		}
	})

	runRule("HSM09", "state_behaviors_must_be_declared_inside_state", func(t *testing.T) {
		cases := []struct {
			name    string
			want    string
			element hsm.RedefinableElement
		}{
			{name: "entry", want: "entry must be called within a State", element: hsm.Entry(modelRulesNoBehavior)},
			{name: "exit", want: "exit must be called within a State", element: hsm.Exit(modelRulesNoBehavior)},
			{name: "activity", want: "activity must be called within a State", element: hsm.Activity(modelRulesNoBehavior)},
			{name: "defer", want: "defer must be called within a State", element: hsm.Defer(modelRulesAdvanceEvent)},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "HSM09/"+tc.name, tc.want, func() {
					hsm.Define(
						fmt.Sprintf("BadStateBehaviorOwner_%s", tc.name),
						hsm.State("idle",
							hsm.Transition(
								hsm.On(modelRulesAdvanceEvent),
								tc.element,
								hsm.Target("../done"),
							),
						),
						hsm.State("done"),
						hsm.Initial(hsm.Target("idle")),
					)
				})
			})
		}
	})

	runRule("HSM10", "implicit_completion_transition_registers_final_event", func(t *testing.T) {
		model := hsm.Define(
			"CompletionTransitionModelRulesHSM",
			hsm.State("idle"),
			hsm.State("done"),
			hsm.Transition(
				hsm.Source("idle"),
				hsm.Target("done"),
			),
			hsm.Initial(hsm.Target("idle")),
		)
		for _, snapshot := range model.TransitionSnapshots() {
			if snapshot.Source != "/CompletionTransitionModelRulesHSM/idle" {
				continue
			}
			if len(snapshot.Events) == 1 && snapshot.Events[0] == hsm.FinalEvent.Name {
				return
			}
			t.Fatalf("completion transition events = %v, want %q", snapshot.Events, hsm.FinalEvent.Name)
		}
		t.Fatal("completion transition snapshot not found")
	})

	runRule("HSM10", "explicit_empty_on_event_list_rejected", func(t *testing.T) {
		emptyEvents := []hsm.Event{}
		assertPanicContains(t, "HSM10/explicit_empty_on_event_list_rejected", "requires at least one event", func() {
			hsm.Define(
				"BadEmptyOnEventList",
				hsm.State("idle",
					hsm.Transition(
						hsm.On(emptyEvents...),
						hsm.Target("../done"),
					),
				),
				hsm.State("done"),
				hsm.Initial(hsm.Target("idle")),
			)
		})
	})

	runRule("HSM10", "explicit_empty_defer_event_list_rejected", func(t *testing.T) {
		emptyEvents := []hsm.Event{}
		assertPanicContains(t, "HSM10/explicit_empty_defer_event_list_rejected", "requires at least one event", func() {
			hsm.Define(
				"BadEmptyDeferEventList",
				hsm.State("idle", hsm.Defer(emptyEvents...)),
				hsm.Initial(hsm.Target("idle")),
			)
		})
	})

	runRule("HSM17", "choice_requires_outgoing_transitions", func(t *testing.T) {
		assertPanicContains(t, "HSM17/choice_requires_outgoing_transitions", "you must define at least one transition for choice", func() {
			hsm.Define(
				"BadChoiceWithoutTransitions",
				hsm.State("idle",
					hsm.Choice("branch"),
				),
				hsm.Initial(hsm.Target("idle")),
			)
		})
	})

	runRule("HSM18", "choice_default_branch_must_be_last", func(t *testing.T) {
		assertPanicContains(t, "HSM18/choice_default_branch_must_be_last", "the last transition of choice state", func() {
			hsm.Define(
				"BadChoiceDefaultBranch",
				hsm.State("idle",
					hsm.Choice("branch",
						hsm.Transition(hsm.Target("../left")),
						hsm.Transition(
							hsm.Target("../right"),
							hsm.Guard(modelRulesAlwaysGuard),
						),
					),
				),
				hsm.State("left"),
				hsm.State("right"),
				hsm.Initial(hsm.Target("idle")),
			)
		})
	})

	runRule("HSM19", "source_and_target_paths_must_resolve", func(t *testing.T) {
		cases := []struct {
			name  string
			want  string
			build func()
		}{
			{
				name: "source",
				want: "missing source",
				build: func() {
					hsm.Define(
						"BadMissingSourcePath",
						hsm.State("idle"),
						hsm.State("done"),
						hsm.Transition(
							hsm.On(modelRulesAdvanceEvent),
							hsm.Source("missing"),
							hsm.Target("done"),
						),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "target",
				want: "missing target",
				build: func() {
					hsm.Define(
						"BadMissingTargetPath",
						hsm.State("idle"),
						hsm.Transition(
							hsm.On(modelRulesAdvanceEvent),
							hsm.Source("idle"),
							hsm.Target("missing"),
						),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "HSM19/"+tc.name, tc.want, tc.build)
			})
		}
	})

	runRule("HSM20", "top_level_target_requires_source", func(t *testing.T) {
		t.Run("source_only_internal_allowed", func(t *testing.T) {
			assertNoPanic(t, "HSM20/source_only_internal_allowed", func() {
				hsm.Define(
					"TopLevelInternalTransitionAllowed",
					hsm.State("idle"),
					hsm.Transition(
						hsm.On(modelRulesAdvanceEvent),
						hsm.Source("idle"),
						hsm.Effect(modelRulesNoBehavior),
					),
					hsm.Initial(hsm.Target("idle")),
				)
			})
		})

		t.Run("target_only", func(t *testing.T) {
			assertPanicContains(t, "HSM20/target_only", "top level transitions with a target must also define a source", func() {
				hsm.Define(
					"BadTopLevelTransitionTargetOnly",
					hsm.State("idle"),
					hsm.Transition(
						hsm.On(modelRulesAdvanceEvent),
						hsm.Target("idle"),
					),
					hsm.Initial(hsm.Target("idle")),
				)
			})
		})
	})

	runRule("HSM21", "internal_transition_requires_effect", func(t *testing.T) {
		cases := []struct {
			name  string
			build func()
		}{
			{
				name: "empty",
				build: func() {
					hsm.Define(
						"BadInternalTransitionWithoutEffect",
						hsm.State("idle",
							hsm.Transition(
								hsm.On(modelRulesAdvanceEvent),
							),
						),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "guard_only",
				build: func() {
					hsm.Define(
						"BadGuardOnlyInternalTransition",
						hsm.State("idle",
							hsm.Transition(
								hsm.On(modelRulesAdvanceEvent),
								hsm.Guard(modelRulesAlwaysGuard),
							),
						),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "HSM21/"+tc.name, "internal transitions require an effect", tc.build)
			})
		}
	})

	runRule("MODEL22", "non_initial_pseudostate_transitions_cannot_have_triggers", func(t *testing.T) {
		cases := []struct {
			name  string
			build func()
		}{
			{
				name: "choice",
				build: func() {
					hsm.Define(
						"BadChoiceTriggeredTransition",
						hsm.Initial(hsm.Target("parent")),
						hsm.State("parent",
							hsm.Choice("branch",
								hsm.Transition(
									hsm.On(modelRulesAdvanceEvent),
									hsm.Target("done"),
								),
							),
							hsm.State("done"),
						),
					)
				},
			},
			{
				name: "entry_point",
				build: func() {
					hsm.Define(
						"BadEntryPointTriggeredTransition",
						hsm.Initial(hsm.Target("target")),
						hsm.State("target",
							hsm.EntryPoint("warm",
								hsm.On(modelRulesAdvanceEvent),
								hsm.Target("ready"),
							),
							hsm.Initial(hsm.Target("ready")),
							hsm.State("ready"),
						),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "MODEL22/"+tc.name, "transitions cannot have triggers", tc.build)
			})
		}
	})

	runRule("HSM32", "named_model_members_must_not_be_empty", func(t *testing.T) {
		cases := []struct {
			name  string
			want  string
			build func()
		}{
			{
				name: "attribute",
				want: "attribute name cannot be empty",
				build: func() {
					hsm.Define(
						"BadEmptyAttributeName",
						hsm.Attribute(""),
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "operation",
				want: "operation name cannot be empty",
				build: func() {
					hsm.Define(
						"BadEmptyOperationName",
						hsm.Operation("", func() {}),
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "onset",
				want: "attribute name cannot be empty",
				build: func() {
					hsm.Define(
						"BadEmptyOnSetName",
						hsm.State("idle",
							hsm.Transition(
								hsm.OnSet(""),
								hsm.Target("../done"),
							),
						),
						hsm.State("done"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "oncall",
				want: "operation name cannot be empty",
				build: func() {
					hsm.Define(
						"BadEmptyOnCallName",
						hsm.Operation("run", func() {}),
						hsm.State("idle",
							hsm.Transition(
								hsm.OnCall(""),
								hsm.Target("../done"),
							),
						),
						hsm.State("done"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "HSM32/"+tc.name, tc.want, tc.build)
			})
		}
	})

	runRule("MODEL32", "named_model_members_must_not_contain_slash", func(t *testing.T) {
		child := hsm.Define(
			"SlashNameChild",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
		)
		cases := []struct {
			name  string
			build func()
		}{
			{
				name: "model",
				build: func() {
					hsm.Define(
						"Bad/ModelName",
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "state",
				build: func() {
					hsm.Define(
						"BadSlashState",
						hsm.State("bad/name"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "submachine",
				build: func() {
					hsm.Define(
						"BadSlashSubmachine",
						hsm.SubmachineState("bad/name", child),
						hsm.Initial(hsm.Target("bad")),
					)
				},
			},
			{
				name: "transition",
				build: func() {
					hsm.Define(
						"BadSlashTransition",
						hsm.State("idle",
							hsm.Transition("bad/name", hsm.On(modelRulesAdvanceEvent), hsm.Target("../done")),
						),
						hsm.State("done"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "choice",
				build: func() {
					hsm.Define(
						"BadSlashChoice",
						hsm.State("idle",
							hsm.Choice("bad/name", hsm.Transition(hsm.Target("../done"))),
						),
						hsm.State("done"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "entry_point",
				build: func() {
					hsm.Define(
						"BadSlashEntryPoint",
						hsm.State("idle", hsm.EntryPoint("bad/name", hsm.Target("idle"))),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "exit_point",
				build: func() {
					hsm.Define(
						"BadSlashExitPoint",
						hsm.State("idle", hsm.ExitPoint("bad/name")),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "shallow_history",
				build: func() {
					hsm.Define(
						"BadSlashShallowHistory",
						hsm.State("parent", hsm.ShallowHistory("bad/name")),
						hsm.Initial(hsm.Target("parent")),
					)
				},
			},
			{
				name: "deep_history",
				build: func() {
					hsm.Define(
						"BadSlashDeepHistory",
						hsm.State("parent", hsm.DeepHistory("bad/name")),
						hsm.Initial(hsm.Target("parent")),
					)
				},
			},
			{
				name: "final",
				build: func() {
					hsm.Define(
						"BadSlashFinal",
						hsm.Final("bad/name"),
						hsm.Initial(hsm.Target("bad")),
					)
				},
			},
			{
				name: "attribute",
				build: func() {
					hsm.Define(
						"BadSlashAttribute",
						hsm.Attribute("bad/name"),
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "operation",
				build: func() {
					hsm.Define(
						"BadSlashOperation",
						hsm.Operation("bad/name", func() {}),
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "onset",
				build: func() {
					hsm.Define(
						"BadSlashOnSet",
						hsm.State("idle", hsm.Transition(hsm.OnSet("bad/name"), hsm.Target("../done"))),
						hsm.State("done"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "oncall",
				build: func() {
					hsm.Define(
						"BadSlashOnCall",
						hsm.Operation("run", func() {}),
						hsm.State("idle", hsm.Transition(hsm.OnCall("bad/name"), hsm.Target("../done"))),
						hsm.State("done"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "MODEL32/"+tc.name, "must not contain", tc.build)
			})
		}
	})

	runRule("HSM33", "duplicate_attributes_and_operations_rejected", func(t *testing.T) {
		cases := []struct {
			name  string
			want  string
			build func()
		}{
			{
				name: "attribute",
				want: "attribute \"/BadDuplicateAttribute/dup\" already defined",
				build: func() {
					hsm.Define(
						"BadDuplicateAttribute",
						hsm.Attribute("dup"),
						hsm.Attribute("dup"),
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "operation",
				want: "operation \"/BadDuplicateOperation/run\" already defined",
				build: func() {
					hsm.Define(
						"BadDuplicateOperation",
						hsm.Operation("run", func() {}),
						hsm.Operation("run", func() {}),
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "HSM33/"+tc.name, tc.want, tc.build)
			})
		}
	})

	runRule("HSM34", "oncall_requires_declared_operation", func(t *testing.T) {
		assertPanicContains(t, "HSM34/oncall_requires_declared_operation", "missing operation", func() {
			hsm.Define(
				"BadMissingOnCallOperation",
				hsm.State("idle",
					hsm.Transition(
						hsm.OnCall("missing"),
						hsm.Target("../done"),
					),
				),
				hsm.State("done"),
				hsm.Initial(hsm.Target("idle")),
			)
		})
	})

	runRule("HSM42", "time_based_triggers_require_real_state_source", func(t *testing.T) {
		type builder struct {
			name string
			want string
			add  hsm.RedefinableElement
		}
		cases := []builder{
			{name: "after", want: "after can only be used on transitions where the source is a State", add: hsm.After(modelRulesAfterDuration)},
			{name: "every", want: "every can only be used on transitions where the source is a State", add: hsm.Every(modelRulesAfterDuration)},
			{name: "when", want: "when can only be used on transitions where the source is a State", add: hsm.When(modelRulesWhenChannel)},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "HSM42/"+tc.name, tc.want, func() {
					hsm.Define(
						fmt.Sprintf("BadTimeTriggerSource_%s", tc.name),
						hsm.State("parent",
							hsm.State("idle"),
							hsm.Choice("branch",
								hsm.Transition(
									tc.add,
									hsm.Target("../idle"),
								),
							),
						),
						hsm.Initial(hsm.Target("parent")),
					)
				})
			})
		}
	})

	runRule("HSM49", "final_state_cannot_have_behaviors_or_transitions", func(t *testing.T) {
		assertPanicContains(t, "HSM49/final_state_cannot_have_behaviors_or_transitions", "final state", func() {
			hsm.Define(
				"BadFinalStateTransition",
				hsm.Final("done"),
				hsm.State("idle"),
				hsm.Transition(
					hsm.On(modelRulesAdvanceEvent),
					hsm.Source("done"),
					hsm.Target("idle"),
				),
				hsm.Initial(hsm.Target("idle")),
			)
		})
	})

	runRule("HSM50", "history_pseudostates_must_be_nested_inside_state", func(t *testing.T) {
		cases := []struct {
			name  string
			want  string
			build func()
		}{
			{
				name: "shallow",
				want: "within a nested State",
				build: func() {
					hsm.Define(
						"BadTopLevelShallowHistory",
						hsm.ShallowHistory("remember"),
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
			{
				name: "deep",
				want: "within a nested State",
				build: func() {
					hsm.Define(
						"BadTopLevelDeepHistory",
						hsm.DeepHistory("remember"),
						hsm.State("idle"),
						hsm.Initial(hsm.Target("idle")),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "HSM50/"+tc.name, tc.want, tc.build)
			})
		}
	})

	runRule("MODEL51", "history_pseudostates_require_default_transition", func(t *testing.T) {
		cases := []struct {
			name  string
			build func()
		}{
			{
				name: "shallow",
				build: func() {
					hsm.Define(
						"BadShallowHistoryWithoutDefault",
						hsm.Initial(hsm.Target("parent")),
						hsm.State("parent",
							hsm.Initial(hsm.Target("ready")),
							hsm.State("ready"),
							hsm.ShallowHistory("remember"),
						),
					)
				},
			},
			{
				name: "deep",
				build: func() {
					hsm.Define(
						"BadDeepHistoryWithoutDefault",
						hsm.Initial(hsm.Target("parent")),
						hsm.State("parent",
							hsm.Initial(hsm.Target("ready")),
							hsm.State("ready"),
							hsm.DeepHistory("remember"),
						),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "MODEL51/"+tc.name, "requires a default transition", tc.build)
			})
		}
	})

	runRule("MODEL51A", "state_allows_at_most_one_history_vertex_per_kind", func(t *testing.T) {
		cases := []struct {
			name  string
			want  string
			build func()
		}{
			{
				name: "shallow",
				want: "more than one shallow history vertex",
				build: func() {
					hsm.Define(
						"BadDuplicateShallowHistory",
						hsm.Initial(hsm.Target("parent")),
						hsm.State("parent",
							hsm.Initial(hsm.Target("ready")),
							hsm.State("ready"),
							hsm.ShallowHistory("remember", hsm.Transition(hsm.Target("ready"))),
							hsm.ShallowHistory("again", hsm.Transition(hsm.Target("ready"))),
						),
					)
				},
			},
			{
				name: "deep",
				want: "more than one deep history vertex",
				build: func() {
					hsm.Define(
						"BadDuplicateDeepHistory",
						hsm.Initial(hsm.Target("parent")),
						hsm.State("parent",
							hsm.Initial(hsm.Target("ready")),
							hsm.State("ready"),
							hsm.DeepHistory("remember", hsm.Transition(hsm.Target("ready"))),
							hsm.DeepHistory("again", hsm.Transition(hsm.Target("ready"))),
						),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "MODEL51A/"+tc.name, tc.want, tc.build)
			})
		}
	})

	runRule("MODEL52", "pseudostate_transitions_require_targets", func(t *testing.T) {
		assertPanicContains(t, "MODEL52/choice_targetless_effect", "target is required", func() {
			hsm.Define(
				"BadTargetlessChoiceTransition",
				hsm.Initial(hsm.Target("parent")),
				hsm.State("parent",
					hsm.Choice("branch",
						hsm.Transition(hsm.Effect(modelRulesNoBehavior)),
					),
				),
			)
		})
	})

	runRule("MODEL53", "history_default_transition_targets_owner_region", func(t *testing.T) {
		assertPanicContains(t, "MODEL53/history_target_outside", "must target inside", func() {
			hsm.Define(
				"BadHistoryDefaultOutsideTarget",
				hsm.Initial(hsm.Target("parent")),
				hsm.State("parent",
					hsm.Initial(hsm.Target("ready")),
					hsm.State("ready"),
					hsm.ShallowHistory("remember", hsm.Transition(hsm.Target("../outside"))),
				),
				hsm.State("outside"),
			)
		})
	})

	runRule("MODEL54", "entry_point_targets_declaring_boundary", func(t *testing.T) {
		cases := []struct {
			name  string
			want  string
			build func()
		}{
			{
				name: "outside",
				want: "must target inside",
				build: func() {
					hsm.Define(
						"BadEntryPointOutsideTarget",
						hsm.Initial(hsm.Target("boundary")),
						hsm.State("boundary",
							hsm.EntryPoint("warm", hsm.Target("../outside")),
							hsm.Initial(hsm.Target("inside")),
							hsm.State("inside"),
						),
						hsm.State("outside"),
					)
				},
			},
			{
				name: "exit_point",
				want: "cannot target exit point",
				build: func() {
					hsm.Define(
						"BadEntryPointExitPointTarget",
						hsm.EntryPoint("warm", hsm.Target("done")),
						hsm.ExitPoint("done"),
						hsm.Initial(hsm.Target("idle")),
						hsm.State("idle"),
					)
				},
			},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				assertPanicContains(t, "MODEL54/"+tc.name, tc.want, tc.build)
			})
		}
	})

	runRule("MODEL55", "submachine_state_boundary_rejects_child_declarations", func(t *testing.T) {
		child := hsm.Define(
			"SubmachineBoundaryChild",
			hsm.Initial(hsm.Target("idle")),
			hsm.State("idle"),
		)
		assertPanicContains(t, "MODEL55/submachine_child_state", "submachine state cannot contain", func() {
			hsm.Define(
				"BadSubmachineBoundaryChildState",
				hsm.Initial(hsm.Target("child")),
				hsm.SubmachineState("child", child, hsm.State("extra")),
			)
		})
		assertPanicContains(t, "MODEL55/submachine_empty_model_child_state", "requires a machine model", func() {
			hsm.Define(
				"BadSubmachineBoundaryEmptyModelChildState",
				hsm.Initial(hsm.Target("child")),
				hsm.SubmachineState("child", hsm.Model{}, hsm.State("extra")),
			)
		})
	})

	runRule("MODEL56", "submachine_boundaries_require_connection_points", func(t *testing.T) {
		child := hsm.Define(
			"SubmachineBoundaryPathChild",
			hsm.Initial(hsm.Target("inside")),
			hsm.State("inside"),
		)
		assertPanicContains(t, "MODEL56/direct_internal_target", "cannot target internal state", func() {
			hsm.Define(
				"BadSubmachineInternalTarget",
				hsm.Initial(hsm.Target("outside")),
				hsm.SubmachineState("drive", child),
				hsm.State("outside"),
				hsm.Transition(
					hsm.On("go"),
					hsm.Source("outside"),
					hsm.Target("drive/inside"),
				),
			)
		})
		assertPanicContains(t, "MODEL56/direct_internal_source", "submachine internal source", func() {
			hsm.Define(
				"BadSubmachineInternalSource",
				hsm.Initial(hsm.Target("drive")),
				hsm.SubmachineState("drive", child),
				hsm.State("outside"),
				hsm.Transition(
					hsm.On("leave"),
					hsm.Source("drive/inside"),
					hsm.Target("outside"),
				),
			)
		})
		assertPanicContains(t, "MODEL56/internal_escape_target", "cannot target outside submachine boundary", func() {
			escapingChild := hsm.InlineModel(
				hsm.Initial(hsm.Target("inside")),
				hsm.State("inside",
					hsm.Transition(
						hsm.On("escape"),
						hsm.Target("../../outside"),
					),
				),
			)
			hsm.Define(
				"BadSubmachineInternalEscape",
				hsm.Initial(hsm.Target("drive")),
				hsm.SubmachineState("drive", escapingChild),
				hsm.State("outside"),
			)
		})
	})
}
