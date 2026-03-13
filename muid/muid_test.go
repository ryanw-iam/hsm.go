package muid

import (
	"strconv"
	"sync"
	"testing"
)

func TestMakeConcurrentUniqueness(t *testing.T) {
	const (
		workers   = 8
		perWorker = 2048
		total     = workers * perWorker
	)

	ch := make(chan MUID, total)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				ch <- Make()
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	seen := make(map[MUID]struct{}, total)
	for id := range ch {
		if _, ok := seen[id]; ok {
			t.Fatalf("collision detected for %d", id)
		}
		seen[id] = struct{}{}
	}
}

func TestMUIDStringRoundTrip(t *testing.T) {
	samples := []MUID{
		1,
		32,
		1024,
		MUID(1<<20 + 17),
		Make(),
	}

	for _, sample := range samples {
		parsed, err := strconv.ParseUint(sample.String(), 32, 64)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", sample.String(), err)
		}
		if got, want := MUID(parsed), sample; got != want {
			t.Fatalf("round-trip mismatch: got %d, want %d", got, want)
		}
	}
}

func BenchmarkMUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Make()
	}
}
