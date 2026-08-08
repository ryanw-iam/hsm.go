# package hsm

`import "github.com/stateforward/hsm.go"`

Package hsm provides a powerful hierarchical state machine (HSM) implementation for Go.

# Overview

It enables modeling complex state-driven systems with features like hierarchical states,
entry/exit actions, guard conditions, and event-driven transitions. The implementation
follows the Stateforward HSM DSL/runtime contract, ensuring consistency across
platforms.

# Features

  - **Hierarchical States**: Support for nested states and regions.
  - **Event-Driven**: Asynchronous event processing with context propagation.
  - **Guards & Actions**: Flexible functional definitions for transition guards and state actions.
  - **Type Safe**: Generics-based implementation for state context.

# Usage

Define your state machine structure and behavior using the declarative builder pattern:

	type MyHSM struct {
	    hsm.HSM
	    counter int
	}

	model := hsm.Define(
	    "example",
	    hsm.Initial(hsm.Target("foo")),
	    hsm.State("foo"),
	    hsm.State("bar"),
	    hsm.Transition(
	        hsm.On("moveToBar"),
	        hsm.Source("foo"),
	        hsm.Target("bar"),
	    ),
	)

	// Start the state machine
	sm := hsm.Started(context.Background(), &MyHSM{}, &model)
	if err := <-hsm.Dispatch(context.Background(), sm, hsm.Event{Name: "moveToBar"}); err != nil {
		// handle the dispatch failure
	}

# Synchronization

Receive from the `Completion` returned by `Dispatch`, `Set`, `Restart`,
`Stop`, `DispatchAll`, or `DispatchTo` before asserting on post-transition
state. The received error is `nil` on success and identifies a runtime
submission failure otherwise. A completion is one-shot; receive its result
once.

Direct dispatch to a nil, unstarted, or stopped instance returns a failed
completion. Fan-out dispatch filters inactive recipients.

Use `AfterProcess`, `AfterDispatch`, `AfterEntry`, `AfterExit`, and
`AfterExecuted` for tests and deterministic observation only.

<!-- selected public API entries from package hsm -->

- Model construction: `Define`, `Redefine`, `InlineModel`, `State`, `SubmachineState`, `Initial`, `Transition`, `TransitionType`, `Source`, `Target`, `Choice`, `EntryPoint`, `ExitPoint`, `ShallowHistory`, `DeepHistory`, `Final`.
- Behavior and triggers: `Entry`, `Activity`, `Exit`, `Effect`, `Guard`, `On`, `OnSet`, `OnCall`, `After`, `At`, `Every`, `When`, `Defer`, `Attribute`, `Operation`.
- Model hooks: `Validator`, `Finalizer`, and `Observe`; custom hooks use `ModelValidator`, `ModelValidatorFunc`, `DefaultModelValidator`, `ModelFinalizer`, `ModelFinalizerFunc`, and `DefaultModelFinalizer`.
- Runtime lifecycle: `New`, `Started`, `Start`, `Stop`, `Restart`, `Dispatch`, `DispatchAll`, `DispatchTo`, `Set`, `Get`, `Call`, `TakeSnapshot`, `TakeSnapshots`; `TakeSnapshot(ctx, group)` returns ordered `[]Snapshot`.
- Groups and identity: `NewGroup`, `MakeGroup`, `Group.Instances`, `Group.States`, `Group.Snapshots`, `ID`, `QualifiedName`, `Name`, `FromContext`, `InstancesFromContext`; group snapshots preserve flattened member order.
- Deterministic test observation: `AfterProcess`, `AfterDispatch`, `AfterEntry`, `AfterExit`, and `AfterExecuted`.
- Core types: `Model`, `FinalizedModel`, `Config`, `Event`, `Completion`, `ObservationData`, `AttributeChange`, `CallData`, `Snapshot`, `TransitionSnapshot`, `EventSnapshot`, `Queue`, `Group`, `HSM`, `Instance`, `Element`, `RedefinableElement`, `OperationFunc`, and `ExpressionFunc`.
- Built-ins: `InitialEvent`, `ErrorEvent`, `AnyEvent`, `FinalEvent`, `ObservationEvent`, `InfiniteDuration`, `Keys`, `Version`, exported `*Kind` constants, and exported error sentinels including `ErrMissingHSM`, `ErrInvalidState`, `ErrUnknownAttribute`, and `ErrInvalidAttributeType`.
