package hsm

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stateforward/hsm.go/kind"
)

var modelHookElementPCs sync.Map

func apply(model *Model, stack []Element, partials ...RedefinableElement) {
	for _, partial := range partials {
		partial(model, stack)
	}
}

func markModelHookElement(element RedefinableElement) RedefinableElement {
	modelHookElementPCs.Store(reflect.ValueOf(element).Pointer(), struct{}{})
	return element
}

func isModelHookElement(element RedefinableElement) bool {
	if element == nil {
		return false
	}
	_, ok := modelHookElementPCs.Load(reflect.ValueOf(element).Pointer())
	return ok
}

func structuralElements(elements []RedefinableElement) []RedefinableElement {
	structural := make([]RedefinableElement, 0, len(elements))
	for _, element := range elements {
		if !isModelHookElement(element) {
			structural = append(structural, element)
		}
	}
	return structural
}

// Define creates a new state machine model with the given name and elements.
// The first argument can be either a string name or a RedefinableElement.
// Additional elements are added to the model in the order they are specified.
//
// Example:
//
//	model := hsm.Define(
//	    "traffic_light",
//	    hsm.State("red"),
//	    hsm.State("yellow"),
//	    hsm.State("green"),
//	    hsm.Initial(hsm.Target("red")),
//	)
func Define[T stringLike](name T, redefinableElements ...RedefinableElement) FinalizedModel {
	return redefinableModel{
		name:     string(name),
		elements: redefinableElements,
	}.redefine()
}

// InlineModel creates an anonymous model body for generated submachine
// composition. Prefer Define for named reusable models.
func InlineModel(redefinableElements ...RedefinableElement) Model {
	return Model{elements: slices.Clone(redefinableElements)}
}

// Redefine replays an existing model under the same root or under a new root
// name and applies replacement model elements.
func Redefine(model FinalizedModel, args ...any) FinalizedModel {
	name := model.Name()
	redefinableElements := make([]RedefinableElement, 0, len(args))
	if len(args) > 0 {
		if overrideName, ok := stringLikeValue(args[0]); ok {
			name = overrideName
			args = args[1:]
		}
	}
	for _, arg := range args {
		redefinableElement, ok := requireRedefinableElement(arg)
		if !ok {
			panic(fmt.Sprintf("expected string-like or RedefinableElement, got %T", arg))
		}
		redefinableElements = append(redefinableElements, redefinableElement)
	}
	elements := slices.Clone(model.elements)
	for _, redefinableElement := range redefinableElements {
		if isModelHookElement(redefinableElement) {
			elements = append(elements, redefinableElement)
		} else {
			elements = append(elements, replacementElement(redefinableElement))
		}
	}
	return redefinableModel{
		name:      name,
		elements:  elements,
		validator: model.validator,
		finalizer: model.finalizer,
	}.redefine()
}

// Validator overrides the model validator used while defining or redefining a model.
func Validator(validator any) RedefinableElement {
	modelValidator := asModelValidator(validator)
	traceback := traceback()
	return markModelHookElement(func(model *Model, stack []Element) Element {
		if modelValidator == nil {
			traceback(fmt.Errorf("validator must implement ModelValidator or be func(*Model)"))
		}
		model.validator = modelValidator
		return stack[len(stack)-1]
	})
}

func asModelValidator(validator any) ModelValidator {
	switch typed := validator.(type) {
	case ModelValidator:
		return typed
	case func(*Model):
		return ModelValidatorFunc(typed)
	default:
		return nil
	}
}

// Finalizer overrides the model finalizer used while defining or redefining a model.
func Finalizer(finalizer any) RedefinableElement {
	modelFinalizer := asModelFinalizer(finalizer)
	traceback := traceback()
	return markModelHookElement(func(model *Model, stack []Element) Element {
		if modelFinalizer == nil {
			traceback(fmt.Errorf("finalizer must implement ModelFinalizer or be func(*Model) *FinalizedModel"))
		}
		model.finalizer = modelFinalizer
		return stack[len(stack)-1]
	})
}

func asModelFinalizer(finalizer any) ModelFinalizer {
	switch typed := finalizer.(type) {
	case ModelFinalizer:
		return typed
	case func(*Model) *FinalizedModel:
		return ModelFinalizerFunc(typed)
	default:
		return nil
	}
}

// Observe attaches an observer to behavior execution and selected transition events.
func Observe[T Instance](observer func(context.Context, T, Event), targets ...any) RedefinableElement {
	traceback := traceback()
	targetNames := observationTargets(traceback, targets)
	return func(model *Model, stack []Element) Element {
		if observer == nil {
			traceback(fmt.Errorf("observer must be a function"))
		}
		name := path.Join(model.QualifiedName(), fmt.Sprintf("observation_%d", len(model.observers)))
		model.observers = append(model.observers, observation{
			element: element{kind: ObservationKind, qualifiedName: name},
			operation: func(ctx context.Context, hsm Instance, event Event) {
				typed, ok := hsm.(T)
				if !ok {
					return
				}
				observer(ctx, typed, event)
			},
			targets: targetNames,
		})
		return stack[len(stack)-1]
	}
}

func observationTargets(traceback func(error), targets []any) []string {
	if len(targets) == 0 {
		return nil
	}
	names := make([]string, 0, len(targets))
	for _, target := range targets {
		switch typed := target.(type) {
		case string:
			names = append(names, typed)
		case Event:
			names = append(names, typed.Name)
		case *Event:
			if typed != nil {
				names = append(names, typed.Name)
			}
		case Element:
			names = append(names, typed.QualifiedName())
		default:
			reflected := reflect.ValueOf(target)
			if reflected.IsValid() && reflected.Kind() == reflect.String {
				names = append(names, reflected.Convert(stringValueType).Interface().(string))
				continue
			}
			traceback(fmt.Errorf("observer target must be a string, Event, or Element"))
		}
	}
	return names
}

func replacementElement(redefinableElement RedefinableElement) RedefinableElement {
	return func(model *Model, stack []Element) Element {
		replacementScope := &element{kind: NullKind, qualifiedName: path.Join(model.QualifiedName(), ".redefine")}
		return redefinableElement(model, append(stack, replacementScope))
	}
}

func validateBuilderName(traceback func(error), label, name string) {
	if name == "" {
		traceback(fmt.Errorf("%s name cannot be empty", label))
	}
	if strings.Contains(name, "/") {
		traceback(fmt.Errorf("%s name must not contain \"/\": %s", label, name))
	}
}

func isReplacementScope(stack []Element) bool {
	for _, element := range stack {
		if element != nil && element.Kind() == NullKind && element.Name() == ".redefine" {
			return true
		}
	}
	return false
}

func removeMemberSubtree(model *Model, qualifiedName string) {
	removedTransitions := map[string]struct{}{}
	for name, element := range model.members {
		if transition, ok := element.(*transition); ok {
			if name == qualifiedName ||
				IsAncestor(qualifiedName, name) ||
				transition.source == qualifiedName ||
				IsAncestor(qualifiedName, transition.source) ||
				IsAncestor(qualifiedName, transition.target) {
				removedTransitions[name] = struct{}{}
			}
			continue
		}
	}
	for transitionName := range removedTransitions {
		removeTransitionReference(model, transitionName)
		for name := range model.members {
			if name == transitionName || IsAncestor(transitionName, name) {
				delete(model.members, name)
			}
		}
	}
	for name := range model.members {
		if name == qualifiedName || IsAncestor(qualifiedName, name) {
			delete(model.members, name)
		}
	}
}

func removeTransitionReference(model *Model, transitionName string) {
	for _, element := range model.members {
		switch vertex := element.(type) {
		case *state:
			vertex.transitions = removeString(vertex.transitions, transitionName)
		case *vertex:
			vertex.transitions = removeString(vertex.transitions, transitionName)
		}
	}
}

func removeString(values []string, target string) []string {
	if len(values) == 0 {
		return values
	}
	next := values[:0]
	for _, value := range values {
		if value != target {
			next = append(next, value)
		}
	}
	return next
}

func transitionIsLive(model *Model, candidate *transition) bool {
	if model == nil || candidate == nil {
		return false
	}
	current, ok := model.members[candidate.QualifiedName()].(*transition)
	return ok && current == candidate
}

func isReplaceableVertexKind(kindValue uint64) bool {
	return kind.Is(kindValue, StateKind, ChoiceKind, ShallowHistoryKind, DeepHistoryKind)
}

func addStateBoundary(model *Model, stack []Element, label, stateName string, stateKind uint64, traceback func(error)) (*state, []Element) {
	owner := find(stack, NamespaceKind)
	if owner == nil {
		traceback(fmt.Errorf("%s \"%s\" must be called within Define() or State()", label, stateName))
	}
	validateBuilderName(traceback, label, stateName)
	boundary := &state{
		vertex: vertex{element: element{kind: stateKind, qualifiedName: path.Join(owner.QualifiedName(), stateName)}, transitions: []string{}},
	}
	if existing := model.members[boundary.QualifiedName()]; existing != nil {
		if !isReplacementScope(stack) || !isReplaceableVertexKind(existing.Kind()) {
			traceback(fmt.Errorf("state \"%s\" already defined", boundary.QualifiedName()))
		}
		removeMemberSubtree(model, boundary.QualifiedName())
	}
	model.members[boundary.QualifiedName()] = boundary
	return boundary, append(stack, boundary)
}

func (redefinable redefinableModel) redefine() FinalizedModel {
	validateBuilderName(traceback(), "model", redefinable.name)
	definitions := structuralElements(redefinable.elements)
	model := Model{
		state: state{
			vertex: vertex{element: element{kind: StateKind, qualifiedName: path.Join("/", redefinable.name), id: redefinable.name}, transitions: []string{}},
		},
		elements:   slices.Clone(redefinable.elements),
		events:     map[string]*Event{},
		attributes: map[string]*attribute{},
		operations: map[string]*operationDef{},
		validator:  redefinable.validator,
		finalizer:  redefinable.finalizer,
	}
	model.members = map[string]Element{
		model.qualifiedName: &model.state,
	}
	model.events[InitialEvent.Name] = &InitialEvent
	model.events[ErrorEvent.Name] = &ErrorEvent
	model.events[AnyEvent.Name] = &AnyEvent
	model.events[FinalEvent.Name] = &FinalEvent
	model.events[ObservationEvent.Name] = &ObservationEvent
	stack := []Element{&model.state}
	for len(model.elements) > 0 {
		elements := model.elements
		model.elements = []RedefinableElement{}
		apply(&model, stack, elements...)
	}

	validator := model.validator
	if validator == nil {
		validator = DefaultModelValidator{}
	}
	validator.Validate(&model)
	applyObservations(&model)
	model.elements = definitions
	finalizer := model.finalizer
	if finalizer == nil {
		finalizer = DefaultModelFinalizer{}
	}
	finalized := finalizer.Finalize(&model)
	if finalized == nil || finalized.Model == nil || finalized.transitionMap == nil || finalized.deferredMap == nil || finalized.transitionPaths == nil || finalized.historyPaths == nil || finalized.historyTargets == nil {
		panic(fmt.Errorf("finalizer must return a finalized model"))
	}
	return *finalized
}

func applyObservations(model *Model) {
	if model == nil || len(model.observers) == 0 {
		return
	}
	for index := range model.observers {
		observer := &model.observers[index]
		model.members[observer.QualifiedName()] = observer
		members := make([]Element, 0, len(model.members))
		for _, member := range model.members {
			members = append(members, member)
		}
		sort.Slice(members, func(i, j int) bool {
			return members[i].QualifiedName() < members[j].QualifiedName()
		})
		for _, member := range members {
			if member == nil || isObservationMember(model, member.QualifiedName()) {
				continue
			}
			if behavior, ok := member.(observableBehavior); ok && observationMatches(observer, member.QualifiedName()) {
				behavior.wrapObservation(observer.operation)
			}
			transition, ok := member.(*transition)
			if !ok || !observationMatchesTransition(observer, transition) {
				continue
			}
			behaviorName := path.Join(observer.QualifiedName(), "event", strings.TrimPrefix(transition.QualifiedName(), "/"))
			behavior := &behavior[Instance]{
				element: element{kind: BehaviorKind, qualifiedName: behaviorName},
				operation: func(ctx context.Context, hsm Instance, event Event) {
					observer.operation(ctx, hsm, observationEvent(transition.QualifiedName(), "event", event))
				},
			}
			model.members[behavior.QualifiedName()] = behavior
			transition.effect = append([]string{behavior.QualifiedName()}, transition.effect...)
		}
	}
}

func isObservationMember(model *Model, qualifiedName string) bool {
	for index := range model.observers {
		observerName := model.observers[index].QualifiedName()
		if qualifiedName == observerName || IsAncestor(observerName, qualifiedName) {
			return true
		}
	}
	return false
}

func observationMatches(observer *observation, qualifiedName string) bool {
	if observer == nil || len(observer.targets) == 0 {
		return true
	}
	return slices.Contains(observer.targets, qualifiedName)
}

func observationMatchesTransition(observer *observation, transition *transition) bool {
	if transition == nil {
		return false
	}
	if observationMatches(observer, transition.QualifiedName()) {
		return true
	}
	for _, eventName := range transition.events {
		if observationMatches(observer, eventName) {
			return true
		}
	}
	return false
}

func validateModel(model *Model) {
	if model == nil {
		panic(fmt.Errorf("model is nil"))
	}
	if model.state.initial == "" {
		panic(fmt.Errorf("initial state is required for state machine %s", model.state.id))
	}
	if len(model.state.entry) > 0 {
		panic(fmt.Errorf("entry actions are not allowed on top level state machine %s", model.state.id))
	}
	if len(model.state.exit) > 0 {
		panic(fmt.Errorf("exit actions are not allowed on top level state machine %s", model.state.id))
	}
	for _, element := range model.members {
		if behavior, ok := element.(*behavior[Instance]); ok && behavior.operationRef != "" {
			if model.operations == nil || model.operations[behavior.operationRef] == nil {
				panic(fmt.Errorf("missing operation \"%s\" for behavior \"%s\"", behavior.operationRef, behavior.QualifiedName()))
			}
			continue
		}
		if constraint, ok := element.(*constraint[Instance]); ok && constraint.operationRef != "" {
			if model.operations == nil || model.operations[constraint.operationRef] == nil {
				panic(fmt.Errorf("missing operation \"%s\" for guard \"%s\"", constraint.operationRef, constraint.QualifiedName()))
			}
			continue
		}
		if transition, ok := element.(*transition); ok {
			validateTransition(model, transition)
			continue
		}
		if vertex, ok := element.(*vertex); ok && kind.Is(vertex.Kind(), InitialKind, EntryPointKind, ShallowHistoryKind, DeepHistoryKind) {
			if kind.Is(vertex.Kind(), InitialKind) {
				validateInitial(model, vertex)
			} else if kind.Is(vertex.Kind(), EntryPointKind) {
				validateEntryPoint(model, vertex)
			} else {
				validateHistory(model, vertex)
			}
			continue
		}
		if state, ok := element.(*state); ok {
			validateStateHistories(model, state)
			if kind.Is(state.Kind(), FinalStateKind) {
				validateFinalState(model, state)
			}
		}
	}
}

func validateTransition(model *Model, transition *transition) {
	if transition.source == "" {
		panic(fmt.Errorf("source is required for transition \"%s\"", transition.QualifiedName()))
	}
	sourceElement := model.members[transition.source]
	if sourceElement == nil {
		panic(fmt.Errorf("missing source \"%s\" for transition \"%s\"", transition.source, transition.QualifiedName()))
	}
	sourceBoundary := enclosingSubmachineBoundary(model, transition.source)
	transitionOwner := path.Dir(transition.QualifiedName())
	if sourceBoundary != "" && !isWithinBoundary(transitionOwner, sourceBoundary) && !kind.Is(sourceElement.Kind(), ExitPointKind) {
		panic(fmt.Errorf("submachine internal source \"%s\" for transition \"%s\"", transition.source, transition.QualifiedName()))
	}
	if transition.target != "" {
		targetElement, ok := model.members[transition.target]
		if !ok {
			panic(fmt.Errorf("missing target \"%s\" for transition \"%s\"", transition.target, transition.QualifiedName()))
		}
		targetBoundary := enclosingSubmachineBoundary(model, transition.target)
		if sourceBoundary != "" &&
			targetBoundary != sourceBoundary &&
			(targetBoundary == "" || !IsAncestor(sourceBoundary, targetBoundary)) &&
			!kind.Is(sourceElement.Kind(), ExitPointKind) {
			panic(fmt.Errorf("cannot target outside submachine boundary \"%s\" from transition \"%s\"", sourceBoundary, transition.QualifiedName()))
		}
		if targetBoundary != "" &&
			sourceBoundary != targetBoundary &&
			!isWithinBoundary(transitionOwner, targetBoundary) &&
			!kind.Is(targetElement.Kind(), EntryPointKind) &&
			!kind.Is(sourceElement.Kind(), ExitPointKind) {
			panic(fmt.Errorf("cannot target internal state \"%s\" from transition \"%s\"", transition.target, transition.QualifiedName()))
		}
	}
	if path.Dir(transition.QualifiedName()) == model.QualifiedName() && transition.source == model.QualifiedName() && transition.target != "" && !transition.hasEvent(FinalEvent.Name) {
		panic(fmt.Errorf("top level transitions with a target must also define a source"))
	}
	if len(transition.events) == 0 && !kind.Is(sourceElement.Kind(), PseudostateKind) {
		panic(fmt.Errorf("transition \"%s\" has no trigger", transition.QualifiedName()))
	}
	if transition.target == "" && kind.Is(sourceElement.Kind(), PseudostateKind) {
		panic(fmt.Errorf("target is required for transition \"%s\"", transition.QualifiedName()))
	}
	if kind.Is(sourceElement.Kind(), InitialKind) {
		if transition.guard != "" {
			panic(fmt.Errorf("initial \"%s\" cannot have a guard", sourceElement.QualifiedName()))
		}
		for _, event := range transition.events {
			if event != InitialEvent.Name {
				panic(fmt.Errorf("initial \"%s\" must not have a trigger \"%s\"", sourceElement.QualifiedName(), InitialEvent.Name))
			}
		}
	} else if kind.Is(sourceElement.Kind(), PseudostateKind) && len(transition.events) > 0 {
		panic(fmt.Errorf("pseudostate \"%s\" transitions cannot have triggers", sourceElement.QualifiedName()))
	}
	if kind.Is(sourceElement.Kind(), EntryPointKind) && transition.target == "" {
		panic(fmt.Errorf("entry point \"%s\" requires a target", sourceElement.QualifiedName()))
	}
	if targetElement := model.members[transition.target]; targetElement != nil && kind.Is(targetElement.Kind(), EntryPointKind) {
		boundary := model.members[targetElement.Owner()]
		if boundary == nil || !kind.Is(boundary.Kind(), SubmachineStateKind) {
			panic(fmt.Errorf("entry point target requires a submachine transition target"))
		}
		if !kind.Is(sourceElement.Kind(), ExitPointKind) && IsAncestor(targetElement.Owner(), transition.source) {
			panic(fmt.Errorf("entry point target cannot be internal"))
		}
	}
	if kind.Is(transition.kind, InternalKind) && len(transition.effect) == 0 {
		panic(fmt.Errorf("internal transitions require an effect"))
	}
	for _, eventName := range transition.events {
		event := model.events[eventName]
		if event == nil || !kind.Is(event.Kind, CallEventKind) {
			continue
		}
		if model.operations == nil || model.operations[eventName] == nil {
			panic(fmt.Errorf("missing operation \"%s\" for OnCall()", eventName))
		}
	}
}

func enclosingSubmachineBoundary(model *Model, qualifiedName string) string {
	if model == nil || qualifiedName == "" {
		return ""
	}
	for current := path.Dir(qualifiedName); current != "" && current != "." && current != "/"; {
		if member := model.members[current]; member != nil && kind.Is(member.Kind(), SubmachineStateKind) {
			return current
		}
		next := path.Dir(current)
		if next == current {
			break
		}
		current = next
	}
	return ""
}

func isWithinBoundary(qualifiedName, boundary string) bool {
	return qualifiedName == boundary || IsAncestor(boundary, qualifiedName)
}

func validateInitial(model *Model, initial *vertex) {
	if len(initial.transitions) > 1 {
		panic(fmt.Errorf("initial \"%s\" cannot have multiple transitions %v", initial.QualifiedName(), initial.transitions))
	}
	if len(initial.transitions) == 0 {
		return
	}
	transition := get[*transition](model, initial.transitions[0])
	if transition == nil {
		return
	}
	owner := initial.Owner()
	if transition.target != owner && !IsAncestor(owner, transition.target) {
		panic(fmt.Errorf("initial \"%s\" must target a nested state not \"%s\"", initial.QualifiedName(), transition.target))
	}
}

func validateEntryPoint(model *Model, entryPoint *vertex) {
	if len(entryPoint.transitions) > 1 {
		panic(fmt.Errorf("entry point \"%s\" cannot have multiple transitions", entryPoint.QualifiedName()))
	}
	if len(entryPoint.transitions) == 0 {
		return
	}
	transition := get[*transition](model, entryPoint.transitions[0])
	if transition == nil {
		return
	}
	if transition.guard != "" {
		panic(fmt.Errorf("entry point \"%s\" cannot have a guard", entryPoint.QualifiedName()))
	}
	if transition.target == "" {
		panic(fmt.Errorf("entry point \"%s\" requires a target", entryPoint.QualifiedName()))
	}
	target := model.members[transition.target]
	if target == nil {
		return
	}
	if kind.Is(target.Kind(), ExitPointKind) {
		panic(fmt.Errorf("entry point \"%s\" cannot target exit point", entryPoint.QualifiedName()))
	}
	owner := entryPoint.Owner()
	if transition.target != owner && !IsAncestor(owner, transition.target) {
		panic(fmt.Errorf("entry point \"%s\" must target inside \"%s\"", entryPoint.QualifiedName(), owner))
	}
}

func validateHistory(model *Model, history *vertex) {
	if len(history.transitions) == 0 {
		panic(fmt.Errorf("history \"%s\" requires a default transition", history.QualifiedName()))
	}
	owner := history.Owner()
	for _, transitionName := range history.transitions {
		transition := get[*transition](model, transitionName)
		if transition == nil || transition.target == "" {
			continue
		}
		if transition.target != owner && !IsAncestor(owner, transition.target) {
			panic(fmt.Errorf("history \"%s\" must target inside \"%s\"", history.QualifiedName(), owner))
		}
	}
}

func validateStateHistories(model *Model, state *state) {
	var hasShallowHistory bool
	var hasDeepHistory bool
	for _, element := range model.members {
		history, ok := element.(*vertex)
		if !ok || history.Owner() != state.QualifiedName() {
			continue
		}
		if kind.Is(history.Kind(), ShallowHistoryKind) {
			if hasShallowHistory {
				panic(fmt.Errorf("state \"%s\" has more than one shallow history vertex", state.QualifiedName()))
			}
			hasShallowHistory = true
			continue
		}
		if kind.Is(history.Kind(), DeepHistoryKind) {
			if hasDeepHistory {
				panic(fmt.Errorf("state \"%s\" has more than one deep history vertex", state.QualifiedName()))
			}
			hasDeepHistory = true
		}
	}
}

func validateFinalState(model *Model, state *state) {
	if len(state.transitions) > 0 {
		panic(fmt.Errorf("final state \"%s\" cannot have transitions", state.QualifiedName()))
	}
	for _, element := range model.members {
		transition, ok := element.(*transition)
		if ok && transition.source == state.QualifiedName() {
			panic(fmt.Errorf("final state \"%s\" cannot have transitions", state.QualifiedName()))
		}
	}
	if len(state.activities) > 0 {
		panic(fmt.Errorf("final state \"%s\" cannot have activities", state.QualifiedName()))
	}
	if len(state.entry) > 0 {
		panic(fmt.Errorf("final state \"%s\" cannot have an entry action", state.QualifiedName()))
	}
	if len(state.exit) > 0 {
		panic(fmt.Errorf("final state \"%s\" cannot have an exit action", state.QualifiedName()))
	}
}

func finalizeModel(model *Model) *FinalizedModel {
	if model == nil {
		return nil
	}
	model = cloneModel(model)
	finalized := &FinalizedModel{
		Model:           model,
		transitionMap:   map[string]map[string][]*transition{},
		deferredMap:     map[string]map[string]string{},
		transitionPaths: map[*transition]map[string]paths{},
		historyPaths:    map[string]map[string][]string{},
		historyTargets:  map[historyTargetKey]map[string]string{},
	}
	finalizeTransitionPaths(finalized)
	buildHistoryCaches(finalized)
	buildCaches(finalized)
	return finalized
}

func finalizeTransitionPaths(model *FinalizedModel) {
	for _, element := range model.members {
		transition, ok := element.(*transition)
		if !ok {
			continue
		}
		model.transitionPaths[transition] = map[string]paths{}
		finalizeTransitionPathsFor(model, transition)
	}
}

func finalizeTransitionPathsFor(model *FinalizedModel, transition *transition) {
	sourceElement := model.members[transition.source]
	if sourceElement == nil {
		return
	}
	targetElement := model.members[transition.target]
	lcaTarget := transition.target
	if targetElement != nil && kind.Is(targetElement.Kind(), EntryPointKind) {
		lcaTarget = targetElement.Owner()
	}
	lca := LCA(transition.source, lcaTarget)
	if kind.Is(transition.kind, ExternalKind) && IsAncestor(transition.source, lcaTarget) {
		lca = path.Dir(transition.source)
	}
	if kind.Is(sourceElement.Kind(), EntryPointKind) {
		lca = path.Dir(sourceElement.Owner())
	}
	if kind.Is(sourceElement.Kind(), ExitPointKind) {
		if transition.target == sourceElement.Owner() {
			lca = transition.target
		} else if kind.Is(transition.kind, SelfKind) {
			lca = path.Dir(sourceElement.Owner())
		}
	}
	enter := finalizeEnterPath(model, lca, transition.target)
	if targetElement != nil && kind.Is(targetElement.Kind(), EntryPointKind) {
		enter = []string{transition.target}
	}
	if kind.Is(transition.kind, SelfKind) && !kind.Is(sourceElement.Kind(), ExitPointKind) && (targetElement == nil || !kind.Is(targetElement.Kind(), EntryPointKind)) {
		enter = append([]string{sourceElement.QualifiedName()}, enter...)
	}
	if kind.Is(sourceElement.Kind(), InitialKind) {
		model.transitionPaths[transition][path.Dir(sourceElement.QualifiedName())] = paths{
			enter: enter,
			exit:  []string{sourceElement.QualifiedName()},
		}
		return
	}
	for qualifiedName, element := range model.members {
		if (qualifiedName != transition.source && !IsAncestor(transition.source, qualifiedName)) || !kind.Is(element.Kind(), VertexKind) {
			continue
		}
		if kind.Is(sourceElement.Kind(), ExitPointKind) && transition.target == sourceElement.Owner() {
			model.transitionPaths[transition][element.QualifiedName()] = paths{}
			continue
		}
		exit := finalizeExitPath(model, lca, element.QualifiedName(), transition)
		if kind.Is(sourceElement.Kind(), EntryPointKind) {
			exit = []string{sourceElement.QualifiedName()}
		}
		model.transitionPaths[transition][element.QualifiedName()] = paths{
			enter: enter,
			exit:  exit,
		}
	}
}

func finalizeEnterPath(model *FinalizedModel, lca, target string) []string {
	enter := []string{}
	entering := target
	for entering != lca && entering != model.qualifiedName && entering != "" {
		enter = append([]string{entering}, enter...)
		entering = path.Dir(entering)
	}
	return enter
}

func finalizeExitPath(model *FinalizedModel, lca, current string, transition *transition) []string {
	exit := []string{}
	if kind.Is(transition.kind, InternalKind) {
		return exit
	}
	exiting := current
	for exiting != lca && exiting != "" {
		exit = append(exit, exiting)
		if exiting == model.qualifiedName {
			break
		}
		exiting = path.Dir(exiting)
	}
	if kind.Is(transition.kind, SelfKind) && (len(exit) == 0 || exit[len(exit)-1] != transition.source) {
		exit = append(exit, transition.source)
	}
	return exit
}

func buildHistoryCaches(model *FinalizedModel) {
	historyByOwner := map[string][]*vertex{}
	for _, element := range model.members {
		history, ok := element.(*vertex)
		if !ok || !kind.Is(history.Kind(), ShallowHistoryKind, DeepHistoryKind) {
			continue
		}
		owner := history.Owner()
		if owner != "" {
			historyByOwner[owner] = append(historyByOwner[owner], history)
		}
	}
	for owner := range historyByOwner {
		model.historyPaths[owner] = map[string][]string{}
		for target, element := range model.members {
			if _, ok := element.(*state); !ok || target == owner || !IsAncestor(owner, target) {
				continue
			}
			model.historyPaths[owner][target] = finalizeEnterPath(model, owner, target)
		}
	}
	skipOwners := []string{""}
	for owner := range historyByOwner {
		skipOwners = append(skipOwners, owner)
	}
	for stateName, element := range model.members {
		if _, ok := element.(*state); !ok {
			continue
		}
		for _, skipOwner := range skipOwners {
			targets := historyTargetsForState(model, historyByOwner, stateName, skipOwner)
			if len(targets) > 0 {
				model.historyTargets[historyTargetKey{stateName: stateName, skipOwner: skipOwner}] = targets
			}
		}
	}
}

func historyTargetsForState(model *FinalizedModel, historyByOwner map[string][]*vertex, stateName, skipOwner string) map[string]string {
	targets := map[string]string{}
	child := stateName
	parent := path.Dir(child)
	for parent != "" && parent != "." {
		if parent == "/" {
			break
		}
		if parent != skipOwner {
			if element := model.members[parent]; element != nil && kind.Is(element.Kind(), StateKind) {
				for _, history := range historyByOwner[parent] {
					target := stateName
					if kind.Is(history.Kind(), ShallowHistoryKind) {
						target = child
					}
					targets[history.QualifiedName()] = target
				}
			}
		}
		if parent == model.qualifiedName {
			break
		}
		next := path.Dir(parent)
		if next == parent {
			break
		}
		child = parent
		parent = next
	}
	return targets
}

func buildCaches(model *FinalizedModel) {
	type Vertex interface {
		Transitions() []string
	}
	for qualifiedName, element := range model.members {
		if _, ok := element.(Vertex); !ok {
			continue
		}

		model.transitionMap[qualifiedName] = make(map[string][]*transition)
		model.deferredMap[qualifiedName] = make(map[string]string)

		var pathToState []string
		currentPath := qualifiedName
		for currentPath != "" {
			pathToState = append([]string{currentPath}, pathToState...)
			if currentPath == model.qualifiedName {
				break
			}
			currentPath = path.Dir(currentPath)
		}

		shadowedEvents := map[string]struct{}{}
		for i := len(pathToState) - 1; i >= 0; i-- {
			statePath := pathToState[i]
			currentElement := model.members[statePath]
			if currentElement == nil {
				continue
			}

			if vertex, ok := currentElement.(Vertex); ok {
				transitions := make([]*transition, 0, len(vertex.Transitions()))
				for _, transitionQualifiedName := range vertex.Transitions() {
					if trans := get[*transition](model.Model, transitionQualifiedName); trans != nil {
						if _, ok := model.transitionPaths[trans][qualifiedName]; !ok {
							continue
						}
						transitions = append(transitions, trans)
					}
				}
				sort.SliceStable(transitions, func(i, j int) bool {
					return strings.Count(transitions[i].source, "/") > strings.Count(transitions[j].source, "/")
				})
				for _, trans := range transitions {
					for _, eventName := range trans.events {
						if _, shadowed := shadowedEvents[eventName]; shadowed {
							continue
						}
						model.transitionMap[qualifiedName][eventName] = append(model.transitionMap[qualifiedName][eventName], trans)
					}
				}
			}

			if state, ok := currentElement.(*state); ok {
				for _, deferredEvent := range state.deferred {
					shadowedEvents[deferredEvent] = struct{}{}
				}
			}
		}

		for i := len(pathToState) - 1; i >= 0; i-- {
			statePath := pathToState[i]
			currentState, ok := model.members[statePath].(*state)
			if !ok {
				continue
			}
			for _, deferredEvent := range currentState.deferred {
				unguardedTransition := false
				for _, trans := range model.transitionMap[qualifiedName][deferredEvent] {
					if trans.guard == "" && transitionHandlesAtOrBelow(trans, currentState.QualifiedName(), model.qualifiedName) {
						unguardedTransition = true
						break
					}
				}
				if !unguardedTransition {
					if _, exists := model.deferredMap[qualifiedName][deferredEvent]; !exists {
						model.deferredMap[qualifiedName][deferredEvent] = currentState.QualifiedName()
					}
				}
			}
		}
	}
}

func transitionDeclaredAtOrBelow(transition *transition, owner string) bool {
	if transition == nil || owner == "" {
		return false
	}
	transitionOwner := path.Dir(transition.QualifiedName())
	return transitionOwner == owner || IsAncestor(owner, transitionOwner)
}

func transitionHandlesAtOrBelow(transition *transition, owner, modelRoot string) bool {
	if transition == nil || owner == "" {
		return false
	}
	return (transition.source == owner && path.Dir(owner) == modelRoot) ||
		IsAncestor(owner, transition.source) ||
		transitionDeclaredAtOrBelow(transition, owner)
}

func find(stack []Element, maybeKinds ...uint64) Element {
	for i := len(stack) - 1; i >= 0; i-- {
		if kind.Is(stack[i].Kind(), maybeKinds...) {
			return stack[i]
		}
	}
	return nil
}

func traceback(maybeError ...error) func(err error) {
	_, file, line, _ := runtime.Caller(2)
	fn := func(err error) {
		panic(fmt.Sprintf("%s:%d: %v", file, line, err))
	}
	if len(maybeError) > 0 {
		fn(maybeError[0])
	}
	return fn
}

func get[T Element](model *Model, name string) T {
	var zero T
	if model == nil || name == "" {
		return zero
	}
	if element, ok := model.members[name]; ok {
		typed, ok := element.(T)
		if ok {
			return typed
		}
	}
	return zero
}

func getBehavior[T Instance](model *Model, name string) *behavior[T] {
	if typedBehavior := get[*behavior[T]](model, name); typedBehavior != nil {
		return typedBehavior
	}
	if instanceBehavior := get[*behavior[Instance]](model, name); instanceBehavior != nil {
		if fn, ok := instanceBehavior.operationAny.(func(context.Context, T, Event)); ok {
			return &behavior[T]{
				element:      instanceBehavior.element,
				operation:    fn,
				operationRef: instanceBehavior.operationRef,
				operationAny: instanceBehavior.operationAny,
			}
		}
		return &behavior[T]{
			element:      instanceBehavior.element,
			operationRef: instanceBehavior.operationRef,
			operationAny: instanceBehavior.operationAny,
			operation: func(ctx context.Context, hsm T, event Event) {
				instanceBehavior.operation(ctx, hsm, event)
			},
		}
	}
	return nil
}

func getConstraint[T Instance](model *Model, name string) *constraint[T] {
	if typedConstraint := get[*constraint[T]](model, name); typedConstraint != nil {
		return typedConstraint
	}
	if instanceConstraint := get[*constraint[Instance]](model, name); instanceConstraint != nil {
		var expression ExpressionFunc[T]
		if fn, ok := instanceConstraint.expressionAny.(func(context.Context, T, Event) bool); ok {
			expression = fn
		} else if instanceConstraint.expression != nil {
			expression = func(ctx context.Context, hsm T, event Event) bool {
				return instanceConstraint.expression(ctx, hsm, event)
			}
		}
		return &constraint[T]{
			element:       instanceConstraint.element,
			expression:    expression,
			operationRef:  instanceConstraint.operationRef,
			expressionAny: instanceConstraint.expressionAny,
		}
	}
	return nil
}

func getFunctionName(fn any) string {
	if fn == nil {
		return ""
	}
	return path.Base(runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name())
}

func callOperationValue(fn reflect.Value, ctx context.Context, hsm Instance, event Event) []reflect.Value {
	return fn.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(hsm),
		reflect.ValueOf(event),
	})
}

func isReplacementOrCompositionScope(stack []Element) bool {
	if isReplacementScope(stack) {
		return true
	}
	for _, element := range stack {
		if element != nil && kind.Is(element.Kind(), SubmachineStateKind) {
			return true
		}
	}
	return false
}

func valueAssignableToType(value any, typ reflect.Type) bool {
	if typ == nil {
		return true
	}
	if value == nil {
		switch typ.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			return true
		default:
			return false
		}
	}
	valueType := reflect.TypeOf(value)
	return valueType.AssignableTo(typ)
}

func attributeDeclaration(args []any) (reflect.Type, any, bool, error) {
	switch len(args) {
	case 0:
		return nil, nil, false, nil
	case 1:
		if typ, ok := args[0].(reflect.Type); ok {
			return typ, nil, false, nil
		}
		defaultValue := args[0]
		var typ reflect.Type
		if defaultValue != nil {
			typ = reflect.TypeOf(defaultValue)
		}
		return typ, defaultValue, true, nil
	case 2:
		typ, ok := args[0].(reflect.Type)
		if !ok || typ == nil {
			return nil, nil, false, fmt.Errorf("explicit attribute type must be reflect.Type")
		}
		if !valueAssignableToType(args[1], typ) {
			return nil, nil, false, fmt.Errorf("default value does not match explicit attribute type")
		}
		return typ, args[1], true, nil
	default:
		return nil, nil, false, fmt.Errorf("attribute accepts at most type and default value")
	}
}

// State creates a new state element with the given name and optional child elements.
// States can have entry/exit actions, activities, and transitions.
//
// Example:
//
//	hsm.State("active",
//	    hsm.Entry(func(ctx context.Context, hsm *MyHSM, event Event) {
//	        log.Println("Entering active state")
//	    }),
//	    hsm.Activity(func(ctx context.Context, hsm *MyHSM, event Event) {
//	        // Long-running activity
//	    }),
//	    hsm.Exit(func(ctx context.Context, hsm *MyHSM, event Event) {
//	        log.Println("Exiting active state")
//	    })
//	)
func State[T stringLike](name T, partialElements ...RedefinableElement) RedefinableElement {
	stateName := string(name)
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		element, stack := addStateBoundary(model, stack, "state", stateName, StateKind, traceback)
		apply(model, stack, partialElements...)
		return element
	}
}

// SubmachineState composes a reusable child model under a state boundary.
func SubmachineState[T stringLike, M ComposableModel](name T, machine M, partialElements ...RedefinableElement) RedefinableElement {
	stateName := string(name)
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		compositionRoot, elements := compositionElements(machine)
		if len(elements) == 0 {
			traceback(fmt.Errorf("submachine state \"%s\" requires a machine model", stateName))
		}
		boundary, stack := addStateBoundary(model, stack, "submachine state", stateName, SubmachineStateKind, traceback)
		if compositionRoot != "" {
			if model.rebaseRefs == nil {
				model.rebaseRefs = map[string]string{}
			}
			previous, hadPrevious := model.rebaseRefs[compositionRoot]
			model.rebaseRefs[compositionRoot] = boundary.QualifiedName()
			defer func() {
				if hadPrevious {
					model.rebaseRefs[compositionRoot] = previous
				} else {
					delete(model.rebaseRefs, compositionRoot)
				}
			}()
		}
		apply(model, stack, elements...)
		beforePartials := map[string]struct{}{}
		for qualifiedName := range model.members {
			beforePartials[qualifiedName] = struct{}{}
		}
		apply(model, stack, partialElements...)
		for qualifiedName, member := range model.members {
			if _, existed := beforePartials[qualifiedName]; existed {
				continue
			}
			if member.Owner() == boundary.QualifiedName() && kind.Is(member.Kind(), StateKind, PseudostateKind) {
				traceback(fmt.Errorf("submachine state cannot contain %s", member.Name()))
			}
		}
		return boundary
	}
}

// LCA finds the Lowest Common Ancestor between two qualified state names in a hierarchical state machine.
// It takes two qualified names 'a' and 'b' as strings and returns their closest common ancestor.
//
// For example:
// - LCA("/s/s1", "/s/s2") returns "/s"
// - LCA("/s/s1", "/s/s1/s11") returns "/s/s1"
// - LCA("/s/s1", "/s/s1") returns "/s/s1"
func LCA(a, b string) string {
	// if both are the same the lca is the parent
	if a == b {
		return path.Dir(a)
	}
	// if one is empty the lca is the other
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	// if the parents are the same the lca is the parent
	if path.Dir(a) == path.Dir(b) {
		return path.Dir(a)
	}
	// if a is an ancestor of b the lca is a
	if IsAncestor(a, b) {
		return a
	}
	// if b is an ancestor of a the lca is b
	if IsAncestor(b, a) {
		return b
	}
	// otherwise the lca is the lca of the parents
	return LCA(path.Dir(a), path.Dir(b))
}

// IsAncestor checks whether current is an ancestor of target in the state hierarchy.
// It returns true if current appears in the path from the root to target's parent.
// Returns false if current equals target, or if either path is "." (relative root).
// The root path "/" is considered an ancestor of all other paths.
func IsAncestor(current, target string) bool {
	current = path.Clean(current)
	target = path.Clean(target)
	if current == target || current == "." || target == "." {
		return false
	}
	if current == "/" {
		return true
	}
	parent := path.Dir(target)
	for parent != "/" {
		if parent == current {
			return true
		}
		parent = path.Dir(parent)
	}
	return false
}

// Transition creates a new transition between states.
// Transitions can have events, guards, and effects.
//
// Example:
//
//	hsm.Transition(
//	    hsm.On("submit"),
//	    hsm.Source("draft"),
//	    hsm.Target("review"),
//	    hsm.Guard(func(ctx context.Context, hsm *MyHSM, event Event) bool {
//	        return hsm.IsValid()
//	    }),
//	    hsm.Effect(func(ctx context.Context, hsm *MyHSM, event Event) {
//	        log.Println("Transitioning from draft to review")
//	    })
//	)
func Transition[T redefinableOrString](nameOrPartialElement T, partialElements ...RedefinableElement) RedefinableElement {
	name, partialElement, hasPartialElement := normalizeRedefinableOrString(nameOrPartialElement)
	if hasPartialElement {
		partialElements = append([]RedefinableElement{partialElement}, partialElements...)
	}
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, VertexKind)
		if name == "" {
			structuralMembers := 0
			for _, member := range model.members {
				if member == nil || kind.Is(member.Kind(), AttributeKind, OperationKind) {
					continue
				}
				structuralMembers++
			}
			name = fmt.Sprintf("transition_%d", structuralMembers)
		}
		validateBuilderName(traceback, "transition", name)
		if owner == nil {
			traceback(fmt.Errorf("transition \"%s\" must be called within a State() or Define()", name))
		}
		transition := &transition{
			events: []string{},
			element: element{
				kind:          TransitionKind,
				qualifiedName: path.Join(owner.QualifiedName(), name),
			},
			source: ".",
		}
		if existing := model.members[transition.QualifiedName()]; existing != nil {
			if !isReplacementScope(stack) || !kind.Is(existing.Kind(), TransitionKind) {
				traceback(fmt.Errorf("\"%s\" already defined", transition.QualifiedName()))
			}
			removeTransitionReference(model, transition.QualifiedName())
			delete(model.members, transition.QualifiedName())
		}
		model.members[transition.QualifiedName()] = transition
		stack = append(stack, transition)
		apply(model, stack, partialElements...)
		if transition.source == "." || transition.source == "" {
			transition.source = owner.QualifiedName()
		}
		sourceElement, ok := model.members[transition.source]
		if !ok {
			traceback(fmt.Errorf("missing source \"%s\" for transition \"%s\"", transition.source, transition.QualifiedName()))
		}
		switch source := sourceElement.(type) {
		case *state:
			source.transitions = append(source.transitions, transition.QualifiedName())
		case *vertex:
			source.transitions = append(source.transitions, transition.QualifiedName())
		}
		if len(transition.events) == 0 && !kind.Is(sourceElement.Kind(), PseudostateKind) {
			transition.events = append(transition.events, FinalEvent.Name)
		}
		if transition.kind == TransitionKind {
			if transition.target == transition.source {
				transition.kind = SelfKind
			} else if transition.target == "" {
				transition.kind = InternalKind
			} else if IsAncestor(transition.source, transition.target) {
				transition.kind = LocalKind
			} else {
				transition.kind = ExternalKind
			}
		}
		return transition
	}
}

// TransitionType overrides the transition kind selected during model
// construction. kindValue must derive from TransitionKind.
func TransitionType(kindValue uint64) RedefinableElement {
	traceback := traceback()
	return func(_ *Model, stack []Element) Element {
		owner, ok := find(stack, TransitionKind).(*transition)
		if !ok {
			traceback(fmt.Errorf("TransitionType() must be called within a Transition()"))
		}
		if !kind.Is(kindValue, TransitionKind) {
			traceback(fmt.Errorf("TransitionType() requires a transition kind"))
		}
		owner.kind = kindValue
		return owner
	}
}

func rebaseReference(model *Model, qualifiedName string) string {
	if model == nil || len(model.rebaseRefs) == 0 || !path.IsAbs(qualifiedName) {
		return qualifiedName
	}
	root := ""
	boundary := ""
	for candidateRoot, candidateBoundary := range model.rebaseRefs {
		if candidateRoot == "" || candidateBoundary == "" {
			continue
		}
		if qualifiedName == candidateRoot || IsAncestor(candidateRoot, qualifiedName) {
			if len(candidateRoot) > len(root) {
				root = candidateRoot
				boundary = candidateBoundary
			}
		}
	}
	if root == "" {
		return qualifiedName
	}
	if qualifiedName == root {
		return boundary
	}
	return path.Join(boundary, strings.TrimPrefix(strings.TrimPrefix(qualifiedName, root), "/"))
}

func transitionEndpointName[T redefinableOrString](model *Model, stack []Element, owner *transition, nameOrPartialElement T, role string, traceback func(error)) string {
	if partialElement, ok := any(nameOrPartialElement).(RedefinableElement); ok {
		element := partialElement(model, stack)
		if element == nil {
			traceback(fmt.Errorf("transition \"%s\" %s is nil", owner.QualifiedName(), role))
		}
		return element.QualifiedName()
	}
	name, _, _ := normalizeRedefinableOrString(nameOrPartialElement)
	if !path.IsAbs(name) {
		if ancestor := find(stack, StateKind); ancestor != nil {
			name = path.Join(ancestor.QualifiedName(), name)
		}
		return name
	}
	name = rebaseReference(model, name)
	if !IsAncestor(model.qualifiedName, name) {
		name = path.Join(model.qualifiedName, name)
	}
	return name
}

// Source specifies the source state of a transition.
// It can be used within a Transition definition.
//
// Example:
//
//	hsm.Transition(
//	    hsm.Source("idle"),
//	    hsm.Target("running")
//	)
func Source[T redefinableOrString](nameOrPartialElement T) RedefinableElement {
	// Capture the stack depth for use in traceback
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, TransitionKind)
		if owner == nil {
			traceback(fmt.Errorf("hsm.Source() must be called within a hsm.Transition()"))
		}
		transition := owner.(*transition)
		if transition.source != "." && transition.source != "" {
			traceback(fmt.Errorf("transition \"%s\" already has a source \"%s\"", transition.QualifiedName(), transition.source))
		}
		transition.source = transitionEndpointName(model, stack, transition, nameOrPartialElement, "source", traceback)
		return owner
	}
}

// Defer schedules events to be processed after the current state is exited.
//
// Example:
//
//	hsm.Defer(hsm.Event{Name: "event_name"})
func Defer[T interface {
	string | *Event | Event
}](events ...T) RedefinableElement {
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		state, ok := stack[len(stack)-1].(*state)
		if !ok || !kind.Is(state.Kind(), StateKind) {
			traceback(fmt.Errorf("defer must be called within a State"))
		}
		if len(events) == 0 {
			traceback(fmt.Errorf("empty event array: defer requires at least one event"))
		}
		for _, event := range events {
			switch evt := any(event).(type) {
			case string:
				state.deferred = append(state.deferred, evt)
			case *Event:
				state.deferred = append(state.deferred, evt.Name)
			case Event:
				state.deferred = append(state.deferred, evt.Name)
			default:
				traceback(fmt.Errorf("defer must be called with a string, *Event, or Event"))
			}
		}
		return state
	}
}

// Target specifies the target state of a transition.
// It can be used within a Transition definition.
//
// Example:
//
//	hsm.Transition(
//	    hsm.Source("idle"),
//	    hsm.Target("running")
//	)
func Target[T redefinableOrString](nameOrPartialElement T) RedefinableElement {
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, TransitionKind)
		if owner == nil {
			traceback(fmt.Errorf("Target() must be called within Transition()"))
		}
		transition := owner.(*transition)
		if transition.target != "" {
			traceback(fmt.Errorf("transition \"%s\" already has target \"%s\"", transition.QualifiedName(), transition.target))
		}
		transition.target = transitionEndpointName(model, stack, transition, nameOrPartialElement, "target", traceback)
		return transition
	}
}

func addBehavior[T Instance](model *Model, owner Element, behaviorKind uint64, name string, fn func(context.Context, T, Event), rawFn ...any) string {
	var operationAny any
	if len(rawFn) > 0 {
		operationAny = rawFn[0]
	}
	behavior := &behavior[T]{
		element:      element{kind: behaviorKind, qualifiedName: path.Join(owner.QualifiedName(), name)},
		operation:    fn,
		operationAny: operationAny,
	}
	model.members[behavior.QualifiedName()] = behavior
	return behavior.QualifiedName()
}

func operationBehavior(model *Model, owner Element, behaviorKind uint64, operationName string) string {
	qualifiedOperationName := qualifyModelName(model.qualifiedName, operationName)
	qualifiedBehaviorName := path.Join(owner.QualifiedName(), operationName)
	behavior := &behavior[Instance]{
		element:      element{kind: behaviorKind, qualifiedName: qualifiedBehaviorName},
		operationRef: qualifiedOperationName,
		operation: func(ctx context.Context, hsm Instance, event Event) {
			invoker, ok := hsm.(interface {
				invokeOperationReference(context.Context, string, ...any) (any, error)
			})
			if !ok {
				panic(ErrMissingHSM)
			}
			if _, err := invoker.invokeOperationReference(ctx, qualifiedOperationName, event); err != nil {
				panic(err)
			}
		},
	}
	model.members[behavior.QualifiedName()] = behavior
	return behavior.QualifiedName()
}

func behaviorOperation(label string, traceback func(error), operation any) (string, func(context.Context, Instance, Event), any) {
	if fn, ok := operation.(func(context.Context, Instance, Event)); ok {
		return getFunctionName(operation), fn, operation
	}
	if dispatchable, ok := operation.(Instance); ok {
		value := reflect.ValueOf(dispatchable)
		if value.Kind() == reflect.Pointer && value.IsNil() {
			traceback(fmt.Errorf("%s requires functions, operation names, or dispatchable instances", label))
		}
		name := path.Base(ID(dispatchable))
		if name == "." || name == "/" || name == "" {
			name = path.Base(Name(dispatchable))
		}
		if name == "." || name == "/" || name == "" {
			name = "dispatch"
		}
		return name, func(ctx context.Context, _ Instance, event Event) {
			dispatchable.dispatch(ctx, event)
		}, operation
	}
	reflected := reflect.ValueOf(operation)
	if !reflected.IsValid() || reflected.Kind() != reflect.Func {
		traceback(fmt.Errorf("%s requires functions, operation names, or dispatchable instances", label))
	}
	return getFunctionName(operation), func(ctx context.Context, hsm Instance, event Event) {
		callOperationValue(reflected, ctx, hsm, event)
	}, operation
}

func stateBehavior(label string, behaviorKind uint64, operations []any, assign func(*state, string)) RedefinableElement {
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		if len(stack) == 0 {
			traceback(fmt.Errorf("%s must be called within a State", label))
			return nil
		}
		owner, ok := stack[len(stack)-1].(*state)
		if !ok || !kind.Is(owner.Kind(), StateKind) {
			traceback(fmt.Errorf("%s must be called within a State", label))
			return nil
		}
		for _, operation := range operations {
			switch typed := operation.(type) {
			case string:
				validateBuilderName(traceback, "operation", typed)
				assign(owner, operationBehavior(model, owner, behaviorKind, typed))
			default:
				name, fn, raw := behaviorOperation(label, traceback, operation)
				assign(owner, addBehavior(model, owner, behaviorKind, name, fn, raw))
			}
		}
		return owner
	}
}

// Effect defines an action to be executed during a transition.
// The effect function is called after exiting the source state and before entering the target state.
//
// Example:
//
//	hsm.Effect(func(ctx context.Context, hsm *MyHSM, event Event) {
//	    log.Printf("Transitioning with event: %s", event.Name)
//	})
func Effect(operations ...any) RedefinableElement {
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner, ok := find(stack, TransitionKind).(*transition)
		if !ok {
			traceback(fmt.Errorf("effect must be called within a Transition"))
		}
		for _, operation := range operations {
			switch typed := operation.(type) {
			case string:
				validateBuilderName(traceback, "operation", typed)
				owner.effect = append(owner.effect, operationBehavior(model, owner, BehaviorKind, typed))
			default:
				name, fn, raw := behaviorOperation("effect", traceback, operation)
				owner.effect = append(owner.effect, addBehavior(model, owner, BehaviorKind, name, fn, raw))
			}
		}
		return owner
	}
}

// Guard defines a condition that must be true for a transition to be taken.
// If multiple transitions are possible, the first one with a satisfied guard is chosen.
//
// Example:
//
//	hsm.Guard(func(ctx context.Context, hsm *MyHSM, event Event) bool {
//	    return hsm.counter > 10
//	})
func Guard(expression any) RedefinableElement {
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, TransitionKind)
		if owner == nil {
			traceback(fmt.Errorf("guard must be called within a Transition"))
		}
		var name string
		var operationRef string
		var guard func(context.Context, Instance, Event) bool
		if operationName, ok := expression.(string); ok {
			validateBuilderName(traceback, "operation", operationName)
			name = operationName
			qualifiedOperationName := qualifyModelName(model.qualifiedName, operationName)
			operationRef = qualifiedOperationName
			guard = func(ctx context.Context, hsm Instance, event Event) bool {
				invoker, ok := hsm.(interface {
					invokeOperationReference(context.Context, string, ...any) (any, error)
				})
				if !ok {
					panic(ErrMissingHSM)
				}
				result, err := invoker.invokeOperationReference(ctx, qualifiedOperationName, event)
				if err != nil {
					panic(err)
				}
				value, ok := result.(bool)
				return ok && value
			}
		} else {
			reflected := reflect.ValueOf(expression)
			if !reflected.IsValid() || reflected.Kind() != reflect.Func {
				traceback(fmt.Errorf("guard requires a function or operation name"))
			}
			name = getFunctionName(expression)
			guard = func(ctx context.Context, hsm Instance, event Event) bool {
				values := callOperationValue(reflected, ctx, hsm, event)
				if len(values) == 0 {
					return false
				}
				value, ok := values[0].Interface().(bool)
				return ok && value
			}
		}
		constraint := &constraint[Instance]{
			element:       element{kind: ConstraintKind, qualifiedName: path.Join(owner.QualifiedName(), name)},
			expression:    guard,
			operationRef:  operationRef,
			expressionAny: expression,
		}
		model.members[constraint.QualifiedName()] = constraint
		owner.(*transition).guard = constraint.QualifiedName()
		return owner
	}
}

// Attribute declares a model-level attribute with an optional default value.
// Attributes can be observed via OnSet("name") transitions and updated at runtime
// via Set / Instance.Set.
func Attribute[T stringLike](name T, args ...any) RedefinableElement {
	attributeName := string(name)
	traceback := traceback()
	attrType, defaultValue, hasDefault, attrErr := attributeDeclaration(args)
	return func(model *Model, stack []Element) Element {
		validateBuilderName(traceback, "attribute", attributeName)
		if attrErr != nil {
			traceback(fmt.Errorf("attribute \"%s\": %w", attributeName, attrErr))
		}
		qualifiedName := qualifyModelName(model.qualifiedName, attributeName)
		if model.attributes == nil {
			model.attributes = map[string]*attribute{}
		}
		if existing := model.members[qualifiedName]; existing != nil {
			if !kind.Is(existing.Kind(), AttributeKind) {
				traceback(fmt.Errorf("attribute \"%s\" conflicts with existing model member", qualifiedName))
			}
			if !isReplacementOrCompositionScope(stack) {
				traceback(fmt.Errorf("attribute \"%s\" already defined", qualifiedName))
			}
		}
		attr := &attribute{
			element: element{kind: AttributeKind, qualifiedName: qualifiedName},
			name:    qualifiedName,
			typ:     attrType,
		}
		if hasDefault {
			attr.defaultValue = defaultValue
			attr.hasDefault = hasDefault
		}
		model.attributes[qualifiedName] = attr
		model.members[qualifiedName] = attr
		return nil
	}
}

// Operation declares a named callable for Call()/OnCall().
// Supported callables include function values and method expressions.
func Operation[T stringLike](name T, maybeFn ...any) RedefinableElement {
	operationName := string(name)
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		validateBuilderName(traceback, "operation", operationName)
		qualifiedName := qualifyModelName(model.qualifiedName, operationName)
		if model.operations == nil {
			model.operations = map[string]*operationDef{}
		}
		if existing := model.members[qualifiedName]; existing != nil {
			if !kind.Is(existing.Kind(), OperationKind) {
				traceback(fmt.Errorf("operation \"%s\" conflicts with existing model member", qualifiedName))
			}
			if !isReplacementOrCompositionScope(stack) {
				traceback(fmt.Errorf("operation \"%s\" already defined", qualifiedName))
			}
		}
		var fn any
		var fnValue reflect.Value
		var fnType reflect.Type
		if len(maybeFn) > 1 {
			traceback(fmt.Errorf("operation \"%s\" accepts at most one function", qualifiedName))
		}
		if len(maybeFn) == 1 {
			fn = maybeFn[0]
			fnValue = reflect.ValueOf(fn)
			if !fnValue.IsValid() || fnValue.Kind() != reflect.Func {
				traceback(fmt.Errorf("operation \"%s\" must be a function", qualifiedName))
			}
			fnType = fnValue.Type()
		}
		model.operations[qualifiedName] = &operationDef{
			element: element{kind: OperationKind, qualifiedName: qualifiedName},
			name:    qualifiedName,
			fn:      fn,
			fnValue: fnValue,
			fnType:  fnType,
		}
		model.members[qualifiedName] = model.operations[qualifiedName]
		return nil
	}
}

// Initial defines the initial state for a composite state or the entire state machine.
// When a composite state is entered, its initial state is automatically entered.
//
// Example:
//
//	hsm.State("operational",
//	    hsm.State("idle"),
//	    hsm.State("running"),
//	    hsm.Initial(hsm.Target("idle")),
//	)
func Initial(partialElements ...RedefinableElement) RedefinableElement {
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, StateKind)
		if owner == nil {
			traceback(fmt.Errorf("initial must be called within a State or Model"))
		}
		initial := &vertex{
			element: element{kind: InitialKind, qualifiedName: path.Join(owner.QualifiedName(), ".initial")},
		}
		owner.(*state).initial = initial.QualifiedName()
		if existing := model.members[initial.QualifiedName()]; existing != nil {
			if !isReplacementScope(stack) || !kind.Is(existing.Kind(), InitialKind) {
				traceback(fmt.Errorf("initial \"%s\" state already exists for \"%s\"", initial.QualifiedName(), owner.QualifiedName()))
			}
			removeMemberSubtree(model, initial.QualifiedName())
		}
		model.members[initial.QualifiedName()] = initial
		stack = append(stack, initial)
		transition := (Transition(Source(initial.QualifiedName()), append(partialElements, On(InitialEvent))...)(model, stack)).(*transition)
		return transition
	}
}

// Choice creates a pseudo-state that enables dynamic branching based on guard conditions.
// The first transition with a satisfied guard condition is taken.
//
// Example:
//
//	hsm.Choice(
//	    hsm.Transition(
//	        hsm.Target("approved"),
//	        hsm.Guard(func(ctx context.Context, hsm *MyHSM, event Event) bool {
//	            return hsm.score > 700
//	        })
//	    ),
//	    hsm.Transition(
//	        hsm.Target("rejected")
//	    )
//	)
func Choice[T redefinableOrString](elementOrName T, partialElements ...RedefinableElement) RedefinableElement {
	name, partialElement, hasPartialElement := normalizeRedefinableOrString(elementOrName)
	if hasPartialElement {
		partialElements = append([]RedefinableElement{partialElement}, partialElements...)
	}
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, StateKind, TransitionKind)
		if owner == nil {
			traceback(fmt.Errorf("you must call Choice() within a State or Transition"))
		} else if kind.Is(owner.Kind(), TransitionKind) {
			transition := owner.(*transition)
			source := transition.source

			owner = model.members[source]
			if owner == nil {
				traceback(fmt.Errorf("transition \"%s\" targetting \"%s\" requires a source state when using Choice()", transition.QualifiedName(), transition.target))
			} else if kind.Is(owner.Kind(), PseudostateKind) {
				// pseudostates aren't a namespace, so we need to find the containing state
				owner = find(stack, StateKind)
				if owner == nil {
					traceback(fmt.Errorf("you must call Choice() within a State"))
				}
			}
		}
		if name == "" {
			name = fmt.Sprintf("choice_%d", len(model.elements))
		}
		validateBuilderName(traceback, "choice", name)
		qualifiedName := path.Join(owner.QualifiedName(), name)
		element := &vertex{
			element: element{kind: ChoiceKind, qualifiedName: qualifiedName},
		}
		if existing := model.members[qualifiedName]; existing != nil {
			if !isReplacementScope(stack) || !isReplaceableVertexKind(existing.Kind()) {
				traceback(fmt.Errorf("choice \"%s\" already defined", qualifiedName))
			}
			removeMemberSubtree(model, qualifiedName)
		}
		model.members[qualifiedName] = element
		stack = append(stack, element)
		apply(model, stack, partialElements...)
		if len(element.transitions) == 0 {
			traceback(fmt.Errorf("you must define at least one transition for choice \"%s\"", qualifiedName))
		}
		if defaultTransition := get[*transition](model, element.transitions[len(element.transitions)-1]); defaultTransition != nil {
			if defaultTransition.Guard() != "" {
				traceback(fmt.Errorf("the last transition of choice state \"%s\" cannot have a guard", qualifiedName))
			}
		}
		return element
	}
}

// EntryPoint creates or selects a named submachine entry point. Within a State,
// it creates a transient vertex with an optional outgoing transition. Within a
// Transition, it rewrites the transition target to the direct matching entry
// point on the transition target.
func EntryPoint[T stringLike](name T, partialElements ...RedefinableElement) RedefinableElement {
	entryPointName := string(name)
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		if selected, ok := find(stack, TransitionKind).(*transition); ok {
			validateBuilderName(traceback, "entry point", entryPointName)
			if targetElement := model.members[selected.target]; targetElement != nil && !kind.Is(targetElement.Kind(), SubmachineStateKind) {
				traceback(fmt.Errorf("entry point %q requires a submachine transition target", entryPointName))
			}
			target := resolveEntryPointTarget(model, selected.target, entryPointName)
			if target == "" {
				model.elements = append(model.elements, func(model *Model, _ []Element) Element {
					if !transitionIsLive(model, selected) {
						return selected
					}
					if targetElement := model.members[selected.target]; targetElement != nil && !kind.Is(targetElement.Kind(), SubmachineStateKind) {
						traceback(fmt.Errorf("entry point %q requires a submachine transition target", entryPointName))
					}
					target := resolveEntryPointTarget(model, selected.target, entryPointName)
					if target == "" {
						traceback(fmt.Errorf("state \"%s\" has no entry point \"%s\"", selected.target, entryPointName))
					}
					selected.target = target
					return selected
				})
				return selected
			}
			selected.target = target
			return selected
		}
		owner := find(stack, StateKind)
		if owner == nil {
			traceback(fmt.Errorf("entry point must be called within a State or Transition"))
		}
		validateBuilderName(traceback, "entry point", entryPointName)
		qualifiedName := path.Join(owner.QualifiedName(), entryPointName)
		if existing := model.members[qualifiedName]; existing != nil {
			if !isReplacementScope(stack) || !kind.Is(existing.Kind(), EntryPointKind) {
				traceback(fmt.Errorf("entry point \"%s\" already defined", qualifiedName))
			}
			removeMemberSubtree(model, qualifiedName)
		}
		element := &vertex{
			element: element{kind: EntryPointKind, qualifiedName: qualifiedName},
		}
		model.members[qualifiedName] = element
		if len(partialElements) > 0 {
			stack = append(stack, element)
			created := Transition(Source(qualifiedName), partialElements...)(model, stack)
			if transition, ok := created.(*transition); ok {
				transition.kind = LocalKind
			}
		}
		return element
	}
}

func resolveEntryPointTarget(model *Model, transitionTarget, entryPointName string) string {
	if model == nil || transitionTarget == "" || entryPointName == "" {
		return ""
	}
	boundary := model.members[transitionTarget]
	if boundary == nil || !kind.Is(boundary.Kind(), SubmachineStateKind) {
		return ""
	}
	direct := make([]string, 0)
	for _, member := range model.members {
		if member == nil || !kind.Is(member.Kind(), EntryPointKind) || member.Name() != entryPointName {
			continue
		}
		owner := member.Owner()
		if owner == transitionTarget {
			direct = append(direct, member.QualifiedName())
		}
	}
	sort.Strings(direct)
	if len(direct) > 0 {
		return direct[0]
	}
	return ""
}

// ExitPoint creates a transient vertex used to route a submachine outcome to
// a handler transition on the owning state.
func ExitPoint[T stringLike](name T, partialElements ...RedefinableElement) RedefinableElement {
	exitPointName := string(name)
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		if selected, ok := find(stack, TransitionKind).(*transition); ok {
			validateBuilderName(traceback, "exit point", exitPointName)
			boundary := selected.source
			if boundary == "." || boundary == "" {
				if owner := find(stack, StateKind); owner != nil {
					boundary = owner.QualifiedName()
				}
			}
			boundaryElement := model.members[boundary]
			if boundaryElement == nil || !kind.Is(boundaryElement.Kind(), SubmachineStateKind) {
				traceback(fmt.Errorf("exit point %q requires a submachine transition source", exitPointName))
			}
			source := resolveExitPointSource(model, boundary, exitPointName)
			if source == "" {
				traceback(fmt.Errorf("state \"%s\" has no exit point \"%s\"", boundary, exitPointName))
			}
			selected.source = source
			return selected
		}
		owner := find(stack, StateKind)
		if owner == nil {
			traceback(fmt.Errorf("exit point must be called within a State"))
		}
		validateBuilderName(traceback, "exit point", exitPointName)
		qualifiedName := path.Join(owner.QualifiedName(), exitPointName)
		if existing := model.members[qualifiedName]; existing != nil {
			if !isReplacementScope(stack) || !kind.Is(existing.Kind(), ExitPointKind) {
				traceback(fmt.Errorf("exit point \"%s\" already defined", qualifiedName))
			}
			removeMemberSubtree(model, qualifiedName)
		}
		element := &vertex{
			element: element{kind: ExitPointKind, qualifiedName: qualifiedName},
		}
		model.members[qualifiedName] = element
		if len(partialElements) > 0 {
			stack = append(stack, element)
			created := Transition(
				Source(qualifiedName),
				append([]RedefinableElement{Target(owner.QualifiedName())}, partialElements...)...,
			)(model, stack)
			if transition, ok := created.(*transition); ok {
				transition.kind = LocalKind
			}
		}
		return element
	}
}

func resolveExitPointSource(model *Model, boundary, exitPointName string) string {
	if model == nil || boundary == "" || exitPointName == "" {
		return ""
	}
	direct := make([]string, 0)
	nested := make([]string, 0)
	for _, member := range model.members {
		if member == nil || !kind.Is(member.Kind(), ExitPointKind) || member.Name() != exitPointName {
			continue
		}
		if member.QualifiedName() != boundary && !IsAncestor(boundary, member.QualifiedName()) {
			continue
		}
		owner := member.Owner()
		if owner == boundary || path.Dir(owner) == boundary {
			direct = append(direct, member.QualifiedName())
		} else {
			nested = append(nested, member.QualifiedName())
		}
	}
	sort.Strings(direct)
	if len(direct) > 0 {
		return direct[0]
	}
	sort.Strings(nested)
	if len(nested) > 0 {
		return nested[0]
	}
	return ""
}

// ShallowHistory creates a shallow history pseudostate within a composite state.
// If no history is available, any transitions defined on this pseudostate are used;
// otherwise the parent state's initial is used.
func ShallowHistory[T redefinableOrString](elementOrName T, partialElements ...RedefinableElement) RedefinableElement {
	name, partialElement, hasPartialElement := normalizeRedefinableOrString(elementOrName)
	if hasPartialElement {
		partialElements = append([]RedefinableElement{partialElement}, partialElements...)
	}
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, StateKind)
		if owner == nil || owner.QualifiedName() == model.qualifiedName {
			traceback(fmt.Errorf("you must call ShallowHistory() within a nested State"))
		}
		if name == "" {
			name = fmt.Sprintf("shallow_history_%d", len(model.elements))
		}
		validateBuilderName(traceback, "shallow history", name)
		qualifiedName := path.Join(owner.QualifiedName(), name)
		element := &vertex{
			element: element{kind: ShallowHistoryKind, qualifiedName: qualifiedName},
		}
		if existing := model.members[qualifiedName]; existing != nil {
			if !isReplacementScope(stack) || !isReplaceableVertexKind(existing.Kind()) {
				traceback(fmt.Errorf("shallow history \"%s\" already defined", qualifiedName))
			}
			removeMemberSubtree(model, qualifiedName)
		}
		model.members[qualifiedName] = element
		stack = append(stack, element)
		apply(model, stack, partialElements...)
		return element
	}
}

// DeepHistory creates a deep history pseudostate within a composite state.
// If no history is available, any transitions defined on this pseudostate are used;
// otherwise the parent state's initial is used.
func DeepHistory[T redefinableOrString](elementOrName T, partialElements ...RedefinableElement) RedefinableElement {
	name, partialElement, hasPartialElement := normalizeRedefinableOrString(elementOrName)
	if hasPartialElement {
		partialElements = append([]RedefinableElement{partialElement}, partialElements...)
	}
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, StateKind)
		if owner == nil || owner.QualifiedName() == model.qualifiedName {
			traceback(fmt.Errorf("you must call DeepHistory() within a nested State"))
		}
		if name == "" {
			name = fmt.Sprintf("deep_history_%d", len(model.elements))
		}
		validateBuilderName(traceback, "deep history", name)
		qualifiedName := path.Join(owner.QualifiedName(), name)
		element := &vertex{
			element: element{kind: DeepHistoryKind, qualifiedName: qualifiedName},
		}
		if existing := model.members[qualifiedName]; existing != nil {
			if !isReplacementScope(stack) || !isReplaceableVertexKind(existing.Kind()) {
				traceback(fmt.Errorf("deep history \"%s\" already defined", qualifiedName))
			}
			removeMemberSubtree(model, qualifiedName)
		}
		model.members[qualifiedName] = element
		stack = append(stack, element)
		apply(model, stack, partialElements...)
		return element
	}
}

// Entry defines an action to be executed when entering a state.
// The entry action is executed before any internal activities are started.
//
// Example:
//
//	hsm.Entry(func(ctx context.Context, hsm *MyHSM, event Event) {
//	    log.Printf("Entering state with event: %s", event.Name)
//	})
func Entry(operations ...any) RedefinableElement {
	return stateBehavior("entry", BehaviorKind, operations, func(owner *state, qualifiedName string) {
		owner.entry = append(owner.entry, qualifiedName)
	})
}

// Activity defines a long-running action that is executed while in a state.
// The activity is started after the entry action and stopped before the exit action.
//
// Example:
//
//	hsm.Activity(func(ctx context.Context, hsm *MyHSM, event Event) {
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return
//	        case <-time.After(time.Second):
//	            log.Println("Activity tick")
//	        }
//	    }
//	})
func Activity(operations ...any) RedefinableElement {
	return stateBehavior("activity", ConcurrentKind, operations, func(owner *state, qualifiedName string) {
		owner.activities = append(owner.activities, qualifiedName)
	})
}

// Exit defines an action to be executed when exiting a state.
// The exit action is executed after any internal activities are stopped.
//
// Example:
//
//	hsm.Exit(func(ctx context.Context, hsm *MyHSM, event Event) {
//	    log.Printf("Exiting state with event: %s", event.Name)
//	})
func Exit(operations ...any) RedefinableElement {
	return stateBehavior("exit", BehaviorKind, operations, func(owner *state, qualifiedName string) {
		owner.exit = append(owner.exit, qualifiedName)
	})
}

// On defines the events that can cause a transition.
// Multiple events can be specified for a single transition.
//
// Example:
//
//	hsm.Transition(
//	    hsm.On("start", "resume"),
//	    hsm.Source("idle"),
//	    hsm.Target("running")
//	)
func On[T interface{ *Event | Event | ~string }](events ...T) RedefinableElement {
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, TransitionKind)
		if owner == nil {
			traceback(fmt.Errorf("trigger must be called within a Transition"))
		}
		if len(events) == 0 {
			traceback(fmt.Errorf("empty event array: On() requires at least one event"))
		}
		transition := owner.(*transition)
		for _, eventOrName := range events {
			var name string
			var event *Event
			switch e := any(eventOrName).(type) {
			case Event:
				name = e.Name
				event = &e
			case *Event:
				name = e.Name
				event = e
			case string:
				name = e
				event = &Event{Name: name, Kind: EventKind}
			default:
				reflected := reflect.ValueOf(eventOrName)
				if reflected.Kind() == reflect.String {
					name = reflected.Convert(stringValueType).Interface().(string)
					event = &Event{Name: name, Kind: EventKind}
				}
			}
			transition.events = append(transition.events, name)
			registerEvent(traceback, model, event)
		}
		return owner
	}
}

// OnSet creates an attribute-change trigger for the given attribute name.
// It is the attribute-based equivalent of On(...) and is driven by Set.
func OnSet[T stringLike](name T) RedefinableElement {
	attributeName := string(name)
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, TransitionKind)
		if owner == nil {
			traceback(fmt.Errorf("OnSet() must be called within a Transition"))
		}
		validateBuilderName(traceback, "attribute", attributeName)
		qualifiedName := qualifyModelName(model.qualifiedName, attributeName)
		transition := owner.(*transition)
		eventName := qualifiedName
		transition.events = append(transition.events, eventName)
		registerEvent(traceback, model, &Event{
			Kind:   ChangeEventKind,
			Name:   eventName,
			Source: qualifiedName,
		})
		if model.attributes == nil {
			model.attributes = map[string]*attribute{}
		}
		if _, exists := model.attributes[qualifiedName]; !exists {
			if existing := model.members[qualifiedName]; existing != nil && !kind.Is(existing.Kind(), AttributeKind) {
				traceback(fmt.Errorf("attribute \"%s\" conflicts with existing model member", qualifiedName))
			}
			attr := &attribute{
				element: element{kind: AttributeKind, qualifiedName: qualifiedName},
				name:    qualifiedName,
			}
			model.attributes[qualifiedName] = attr
			model.members[qualifiedName] = attr
		}
		return owner
	}
}

func stringArgument(value any) (string, bool) {
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() || reflected.Kind() != reflect.String {
		return "", false
	}
	return reflected.Convert(stringValueType).Interface().(string), true
}

func requireFunction(traceback func(error), label string, expr any) reflect.Value {
	value := reflect.ValueOf(expr)
	if !value.IsValid() || value.Kind() != reflect.Func {
		traceback(fmt.Errorf("%s() requires a function or attribute name", label))
	}
	return value
}

func requireTimerFunction(traceback func(error), label string, expr any, result reflect.Type) reflect.Value {
	value := requireFunction(traceback, label, expr)
	fnType := value.Type()
	contextType := reflect.TypeFor[context.Context]()
	eventType := reflect.TypeFor[Event]()
	if fnType.NumIn() != 3 || !contextType.AssignableTo(fnType.In(0)) || !eventType.AssignableTo(fnType.In(2)) || fnType.NumOut() != 1 || fnType.Out(0) != result {
		traceback(fmt.Errorf("%s() requires func(context.Context, T, Event) %s or an attribute name", label, result))
	}
	return value
}

func invokeTimerFunction[T any](fn reflect.Value, ctx context.Context, hsm Instance, event Event) T {
	values := fn.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(hsm),
		reflect.ValueOf(event),
	})
	return values[0].Interface().(T)
}

func durationAttribute(name string) func(context.Context, Instance, Event) time.Duration {
	return func(ctx context.Context, hsm Instance, _ Event) time.Duration {
		value, ok := Get(ctx, hsm, name)
		if !ok {
			return -1
		}
		duration, ok := value.(time.Duration)
		if !ok {
			return -1
		}
		return duration
	}
}

func timeAttribute(name string) func(context.Context, Instance, Event) time.Time {
	return func(ctx context.Context, hsm Instance, _ Event) time.Time {
		value, ok := Get(ctx, hsm, name)
		if !ok {
			return time.Time{}
		}
		timepoint, ok := value.(time.Time)
		if !ok {
			return time.Time{}
		}
		return timepoint
	}
}

func generatedTriggerActivity(callName string, name string, operation func(Event, *state) func(context.Context, Instance, Event)) RedefinableElement {
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner, ok := find(stack, TransitionKind).(*transition)
		if !ok {
			traceback(fmt.Errorf("%s must be called within a Transition", callName))
		}
		qualifiedName := path.Join(owner.QualifiedName(), name)
		for index := 1; slices.Contains(owner.events, qualifiedName); index++ {
			qualifiedName = path.Join(owner.QualifiedName(), fmt.Sprintf("%s_%d", name, index))
		}
		event := Event{
			Kind: TimeEventKind,
			Name: qualifiedName,
		}
		owner.events = append(owner.events, qualifiedName)
		registerEvent(traceback, model, &event)
		model.push(func(model *Model, stack []Element) Element {
			if !transitionIsLive(model, owner) {
				return owner
			}
			maybeSource, ok := model.members[owner.source]
			if !ok {
				traceback(fmt.Errorf("source \"%s\" for transition \"%s\" not found", owner.source, owner.QualifiedName()))
			}
			source, ok := maybeSource.(*state)
			if !ok {
				traceback(fmt.Errorf("%s can only be used on transitions where the source is a State, not \"%s\"", callName, maybeSource.QualifiedName()))
			}
			activity := &behavior[Instance]{
				element:   element{kind: ConcurrentKind, qualifiedName: path.Join(source.QualifiedName(), "activity", qualifiedName)},
				operation: operation(event, source),
			}
			model.members[activity.QualifiedName()] = activity
			source.activities = append(source.activities, activity.QualifiedName())
			return owner
		})
		return owner
	}
}

func dispatchTimerEvent(ctx context.Context, hsm Instance, event Event, duration time.Duration) {
	timer := hsm.Clock().NewTimer(duration)
	select {
	case <-timer.C:
		timer.Stop()
		hsm.dispatch(hsm.Context(), event)
		return
	case <-ctx.Done():
		timer.Stop()
		return
	}
}

// OnCall creates a trigger for Call() invocations of the named operation.
func OnCall[T stringLike](name T) RedefinableElement {
	operationName := string(name)
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, TransitionKind)
		if owner == nil {
			traceback(fmt.Errorf("OnCall() must be called within a Transition"))
		}
		validateBuilderName(traceback, "operation", operationName)
		qualifiedName := qualifyModelName(model.qualifiedName, operationName)
		transition := owner.(*transition)
		transition.events = append(transition.events, qualifiedName)
		registerEvent(traceback, model, &Event{
			Kind:   CallEventKind,
			Name:   qualifiedName,
			Source: qualifiedName,
		})
		return owner
	}
}

// After creates a time-based transition that occurs after a specified duration.
// The duration can be dynamically computed based on the state machine's context.
//
// Example:
//
//	hsm.Transition(
//	    hsm.After(func(ctx context.Context, hsm *MyHSM, event Event) time.Duration {
//	        return time.Second * 30
//	    }),
//	    hsm.Source("active"),
//	    hsm.Target("timeout")
//	)
func After(expr any) RedefinableElement {
	traceback := traceback()
	name := "duration"
	duration := durationAttribute("")
	if attributeName, ok := stringArgument(expr); ok {
		duration = durationAttribute(attributeName)
	} else {
		fn := requireTimerFunction(traceback, "After", expr, reflect.TypeFor[time.Duration]())
		duration = func(ctx context.Context, hsm Instance, event Event) time.Duration {
			return invokeTimerFunction[time.Duration](fn, ctx, hsm, event)
		}
	}
	return generatedTriggerActivity("after", name, func(event Event, source *state) func(context.Context, Instance, Event) {
		return func(ctx context.Context, hsm Instance, _ Event) {
			duration := duration(ctx, hsm, event)
			if duration < 0 {
				return
			}
			dispatchTimerEvent(ctx, hsm, event, duration)
		}
	})
}

// At creates a time-based transition that occurs at a specific timestamp.
// The timestamp can be dynamically computed based on the state machine's context.
//
// Example:
//
//	hsm.Transition(
//	    hsm.At(func(ctx context.Context, hsm *MyHSM, event Event) time.Time {
//	        return time.Now().Add(time.Minute)
//	    }),
//	    hsm.Source("active"),
//	    hsm.Target("timeout")
//	)
func At(expr any) RedefinableElement {
	traceback := traceback()
	name := "timepoint"
	timepoint := timeAttribute("")
	if attributeName, ok := stringArgument(expr); ok {
		timepoint = timeAttribute(attributeName)
	} else {
		fn := requireTimerFunction(traceback, "At", expr, reflect.TypeFor[time.Time]())
		timepoint = func(ctx context.Context, hsm Instance, event Event) time.Time {
			return invokeTimerFunction[time.Time](fn, ctx, hsm, event)
		}
	}
	return generatedTriggerActivity("at", name, func(event Event, source *state) func(context.Context, Instance, Event) {
		return func(ctx context.Context, hsm Instance, _ Event) {
			dispatchTimerEvent(ctx, hsm, event, time.Until(timepoint(ctx, hsm, event)))
		}
	})
}

// Every schedules events to be processed on an interval.
//
// Example:
//
//	hsm.Every(func(ctx context.Context, hsm T, event Event) time.Duration {
//	    return time.Second * 30
//	})
func Every(expr any) RedefinableElement {
	traceback := traceback()
	name := "duration"
	durationExpression := durationAttribute("")
	if attributeName, ok := stringArgument(expr); ok {
		durationExpression = durationAttribute(attributeName)
	} else {
		fn := requireTimerFunction(traceback, "Every", expr, reflect.TypeFor[time.Duration]())
		durationExpression = func(ctx context.Context, hsm Instance, event Event) time.Duration {
			return invokeTimerFunction[time.Duration](fn, ctx, hsm, event)
		}
	}
	return generatedTriggerActivity("every", name, func(event Event, source *state) func(context.Context, Instance, Event) {
		return func(ctx context.Context, hsm Instance, _ Event) {
			duration := durationExpression(ctx, hsm, event)
			if duration <= 0 {
				return
			}
			for {
				timer := hsm.Clock().NewTimer(duration)
				select {
				case <-timer.C:
					timer.Stop()
					dispatched := hsm.dispatch(hsm.Context(), event)
					select {
					case <-dispatched:
					case <-ctx.Done():
						return
					}
					select {
					case <-ctx.Done():
						return
					default:
					}
					currentState := hsm.State()
					if currentState != source.QualifiedName() && !IsAncestor(source.QualifiedName(), currentState) {
						return
					}
					duration = durationExpression(ctx, hsm, event)
					if duration <= 0 {
						return
					}
				case <-ctx.Done():
					timer.Stop()
					return
				}
			}
		}
	})
}

func When(expr any) RedefinableElement {
	traceback := traceback()
	if attributeName, ok := stringArgument(expr); ok {
		return OnSet(attributeName)
	}
	fn := requireFunction(traceback, "When", expr)
	fnType := fn.Type()
	contextType := reflect.TypeFor[context.Context]()
	eventType := reflect.TypeFor[Event]()
	channelType := reflect.TypeFor[<-chan struct{}]()
	if fnType.NumIn() != 3 || !contextType.AssignableTo(fnType.In(0)) || !eventType.AssignableTo(fnType.In(2)) || fnType.NumOut() != 1 || !fnType.Out(0).AssignableTo(channelType) {
		traceback(fmt.Errorf("When() requires func(context.Context, T, Event) <-chan struct{} or an attribute name"))
	}
	name := getFunctionName(expr)
	return generatedTriggerActivity("when", name, func(event Event, source *state) func(context.Context, Instance, Event) {
		return func(ctx context.Context, hsm Instance, _ Event) {
			values := fn.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(hsm),
				reflect.ValueOf(event),
			})
			ch := values[0].Interface().(<-chan struct{})
			for {
				select {
				case _, ok := <-ch:
					dispatched := hsm.dispatch(hsm.Context(), event)
					select {
					case <-dispatched:
					case <-ctx.Done():
						return
					}
					if !ok {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}
	})
}

// Final creates a final state that represents the completion of a composite state or the entire state machine.
// When a final state is entered, a completion event is generated.
//
// Example:
//
//	hsm.State("process",
//	    hsm.State("working"),
//	    hsm.Final("done"),
//	    hsm.Transition(
//	        hsm.Source("working"),
//	        hsm.Target("done")
//	    )
//	)
func Final[T stringLike](name T) RedefinableElement {
	finalName := string(name)
	traceback := traceback()
	return func(model *Model, stack []Element) Element {
		owner := find(stack, NamespaceKind)
		if owner == nil {
			traceback(fmt.Errorf("final \"%s\" must be called within Define() or State()", finalName))
		}
		validateBuilderName(traceback, "final", finalName)
		state := &state{
			vertex: vertex{element: element{kind: FinalStateKind, qualifiedName: path.Join(owner.QualifiedName(), finalName)}, transitions: []string{}},
		}
		if existing := model.members[state.QualifiedName()]; existing != nil {
			if !isReplacementScope(stack) || !kind.Is(existing.Kind(), StateKind) {
				traceback(fmt.Errorf("final \"%s\" already defined", state.QualifiedName()))
			}
			removeMemberSubtree(model, state.QualifiedName())
		}
		model.members[state.QualifiedName()] = state
		return state
	}
}

// Match provides a simple interface, handling basic cases directly
// and delegating complex matching to the match function.
func Match[V stringLike, P stringLike](value V, patterns ...P) bool {
	matchValue := string(value)
	for _, pattern := range patterns {
		matchPattern := string(pattern)
		// fast path for exact match
		if matchPattern == matchValue {
			return true
		}
		// fast path for pure wildcard match
		if matchPattern == "*" {
			return true
		}
		patternLen := len(matchPattern)
		// fast path for empty pattern
		if patternLen == 0 {
			return matchValue == ""
		}
		// fast path for long strings with a pattern that ends with "*
		if matchPattern[patternLen-1] == '*' && strings.HasPrefix(matchValue, matchPattern[:patternLen-1]) {
			return true
		}
		// parse the value and pattern to check for a match
		if parse(matchValue, matchPattern) {
			return true
		}
	}
	return false
}

// parse implements wildcard matching using a goto-based iterative approach.
// It supports the '*' wildcard, which matches zero or more characters.
func parse(value, pattern string) bool {
	valueIndex, patternIndex := 0, 0
	valueLen, patternLen := len(value), len(pattern)
	// patternStarIndex: index of the last '*' encountered in the pattern p.
	// valueStarIndex: index in the value string v corresponding to the position *after* the characters matched by the last '*'.
	patternStarIndex, valueStarIndex := -1, -1

LOOP_START:
	// Check if the current pattern character is '*'
	if patternIndex < patternLen && pattern[patternIndex] == '*' {
		patternStarIndex = patternIndex // Remember the position of this '*'
		patternIndex++                  // Advance the pattern index past the '*'
		valueStarIndex = valueIndex     // Remember the value index where '*' matching might backtrack to
		// If '*' is the last character in the pattern, it matches the rest of the value
		if patternIndex == patternLen {
			return true
		}
		// Continue processing, effectively trying to match zero characters with '*' first
		goto LOOP_START
	}

	// Check if current characters match
	if valueIndex < valueLen && patternIndex < patternLen && pattern[patternIndex] == value[valueIndex] {
		valueIndex++    // Advance value index
		patternIndex++  // Advance pattern index
		goto LOOP_START // Continue matching the next characters
	}

	// Check if we have reached the end of both strings
	if valueIndex == valueLen && patternIndex == patternLen {
		return true // Both strings are exhausted, successful match
	}

	// Check if we reached the end of the value string, but the pattern string remains
	if valueIndex == valueLen && patternIndex < patternLen {
		// Consume any trailing '*' characters in the pattern
		for patternIndex < patternLen && pattern[patternIndex] == '*' {
			patternIndex++
		}
		// If the pattern is now exhausted, it's a match
		return patternIndex == patternLen
	}

	// Mismatch occurred, or end of pattern reached while value string still has characters.
	// Try backtracking if a '*' was previously encountered.
	if patternStarIndex != -1 {
		// Backtrack: Advance the value index associated with the last '*'
		valueStarIndex++
		// If the backtracking value index goes beyond the value string length, matching failed
		if valueStarIndex > valueLen {
			return false
		}
		valueIndex = valueStarIndex         // Reset the current value index to the new backtrack position
		patternIndex = patternStarIndex + 1 // Reset the pattern index to the character immediately after the last '*'
		goto LOOP_START                     // Retry matching from the new state
	}
	return false
}
