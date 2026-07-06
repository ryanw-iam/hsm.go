package hsm

import "testing"

func TestCustomValidatorOverridesBuiltInValidator(t *testing.T) {
	called := false
	model := Define(
		"CustomValidatorOverrideHSM",
		Validator(func(model *Model) {
			called = true
		}),
		Finalizer(func(model *Model) *FinalizedModel {
			return &FinalizedModel{
				Model:           model,
				transitionMap:   map[string]map[string][]*transition{},
				deferredMap:     map[string]map[string]string{},
				transitionPaths: map[*transition]map[string]paths{},
				historyPaths:    map[string]map[string][]string{},
				historyTargets:  map[historyTargetKey]map[string]string{},
			}
		}),
		State("idle"),
	)

	if !called {
		t.Fatal("custom validator did not run")
	}
	if model.QualifiedName() != "/CustomValidatorOverrideHSM" {
		t.Fatalf("model = %q, want custom finalized model", model.QualifiedName())
	}
}

type defaultCloneBox struct {
	Values []string
}

type defaultCloneHSM struct {
	HSM
}

func TestRuntimeAttributeDefaultsAreClonedPerInstance(t *testing.T) {
	model := Define(
		"RuntimeDefaultCloneHSM",
		Attribute("box", &defaultCloneBox{Values: []string{"original"}}),
		Initial(Target("idle")),
		State("idle"),
	)
	first := New(&defaultCloneHSM{}, &model)
	second := New(&defaultCloneHSM{}, &model)

	firstRuntime := first.HSM.instance.(*hsm[*defaultCloneHSM])
	secondRuntime := second.HSM.instance.(*hsm[*defaultCloneHSM])
	firstValue, _ := firstRuntime.attributes.Load("/RuntimeDefaultCloneHSM/box")
	secondValue, _ := secondRuntime.attributes.Load("/RuntimeDefaultCloneHSM/box")
	firstBox := firstValue.(*defaultCloneBox)
	secondBox := secondValue.(*defaultCloneBox)
	if firstBox == secondBox {
		t.Fatal("runtime instances share attribute default pointer")
	}
	firstBox.Values[0] = "mutated"
	if secondBox.Values[0] != "original" {
		t.Fatalf("second runtime saw first runtime mutation: %q", secondBox.Values[0])
	}
}
