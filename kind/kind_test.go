package kind

import (
	"testing"
)

func resetKindCounter(t *testing.T, start uint64) {
	t.Helper()
	previous := n
	n = start
	t.Cleanup(func() {
		n = previous
	})
}

func TestKinds(t *testing.T) {
	resetKindCounter(t, 1)

	baseA := Make()
	baseB := Make()
	duplicateBases := Make(baseA, baseA, baseB)
	derived := Make(duplicateBases, baseA, baseB)
	unrelated := Make()

	t.Run("unique ids are allocated sequentially", func(t *testing.T) {
		if got, want := baseA&idMask, Kind(1); got != want {
			t.Fatalf("baseA id = %d, want %d", got, want)
		}
		if got, want := baseB&idMask, Kind(2); got != want {
			t.Fatalf("baseB id = %d, want %d", got, want)
		}
		if got, want := duplicateBases&idMask, Kind(3); got != want {
			t.Fatalf("duplicateBases id = %d, want %d", got, want)
		}
		if got, want := derived&idMask, Kind(4); got != want {
			t.Fatalf("derived id = %d, want %d", got, want)
		}
	})

	t.Run("list flattens ancestry and preserves first-seen order", func(t *testing.T) {
		if got, want := List(duplicateBases), [depthMax]Kind{1, 2}; got != want {
			t.Fatalf("List(duplicateBases) = %v, want %v", got, want)
		}
		if got, want := List(derived), [depthMax]Kind{3, 1, 2}; got != want {
			t.Fatalf("List(derived) = %v, want %v", got, want)
		}
	})

	t.Run("is matches self and transitive bases only", func(t *testing.T) {
		if !Is(derived, derived) {
			t.Fatal("derived kind should match itself")
		}
		if !Is(derived, duplicateBases) {
			t.Fatal("derived kind should match direct base")
		}
		if !Is(derived, baseA) {
			t.Fatal("derived kind should match transitive baseA")
		}
		if !Is(derived, baseB) {
			t.Fatal("derived kind should match transitive baseB")
		}
		if Is(derived, unrelated) {
			t.Fatal("derived kind should not match unrelated base")
		}
	})
}

func TestKindsIDSpaceWrapsAfterEightBits(t *testing.T) {
	resetKindCounter(t, idMask)

	last := Make()
	wrapped := Make()

	if got, want := last&idMask, Kind(idMask); got != want {
		t.Fatalf("last id before wrap = %d, want %d", got, want)
	}
	if got := wrapped & idMask; got != 0 {
		t.Fatalf("wrapped id = %d, want 0 after 8-bit overflow", got)
	}
	if got, want := List(wrapped), [depthMax]Kind{}; got != want {
		t.Fatalf("List(wrapped) = %v, want %v", got, want)
	}
	if !Is(wrapped, wrapped) {
		t.Fatal("wrapped kind should still self-match")
	}
}
