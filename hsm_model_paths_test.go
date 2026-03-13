package hsm_test

import (
	"testing"

	"github.com/stateforward/hsm.go"
)

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
				hsm.ShallowHistory("remember"),
				hsm.DeepHistory("deep-remember"),
			),
		)
	})
}
