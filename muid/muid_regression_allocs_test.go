package muid_test

import (
	"testing"

	"github.com/stateforward/hsm.go/muid"
)

func TestMakeRegressionAllocs(t *testing.T) {
	const runs = 1000

	allocs := testing.AllocsPerRun(runs, func() {
		_ = muid.Make()
	})

	assertAllocsWithin(t, "Make", allocs, 0)
}

func TestMakeStringRegressionAllocs(t *testing.T) {
	const runs = 1000

	allocs := testing.AllocsPerRun(runs, func() {
		_ = muid.MakeString()
	})

	assertAllocsWithin(t, "MakeString", allocs, 1)
}

func assertAllocsWithin(t *testing.T, name string, got float64, max float64) {
	t.Helper()
	if got > max {
		t.Fatalf("%s allocs/run = %.2f, want <= %.0f", name, got, max)
	}
}
