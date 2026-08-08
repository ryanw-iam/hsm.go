package hsm

import (
	"runtime"
	"testing"
)

// TestCompletionErrIsNotPoisonedByRecycledChannelAllocations guards against
// keying completionErrors by the channel's numeric address. Failed completions
// whose channels have been garbage collected must never surface their stale
// errors on later, successful completions that the allocator happens to place
// at a recycled address.
func TestCompletionErrIsNotPoisonedByRecycledChannelAllocations(t *testing.T) {
	for range 50_000 {
		_ = failedCompletion(ErrInvalidState)
	}
	runtime.GC()
	runtime.GC()

	for i := range 200_000 {
		done := make(chan struct{})
		close(done)
		if err := completion(done).Err(); err != nil {
			t.Fatalf("fresh successful completion %d reports a stale error: %v", i, err)
		}
	}
}

// TestCompletionErrSurvivesGarbageCollection pins the intended retention
// semantics of the fix: a failed completion still held by a caller must keep
// reporting its error even after garbage collection runs.
func TestCompletionErrSurvivesGarbageCollection(t *testing.T) {
	failed := failedCompletion(ErrInvalidState)
	runtime.GC()
	if err := failed.Err(); err != ErrInvalidState {
		t.Fatalf("failed completion Err = %v, want ErrInvalidState", err)
	}
	if err := failed.Wait(); err != ErrInvalidState {
		t.Fatalf("failed completion Wait = %v, want ErrInvalidState", err)
	}
}
