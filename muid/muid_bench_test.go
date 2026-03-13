package muid_test

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"testing"

	"github.com/aidarkhanov/nanoid/v2"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/stateforward/hsm.go/muid"
)

// BenchmarkMUIDGeneration benchmarks the generation of MUIDs
func BenchmarkMUIDGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = muid.Make()
	}
}

// BenchmarkMUIDStringGeneration benchmarks the direct generation of MUID strings
func BenchmarkMUIDStringGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = muid.MakeString()
	}
}

// BenchmarkUUIDv4Generation benchmarks the generation of UUID v4
func BenchmarkUUIDv4Generation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uuid.New()
	}
}

// BenchmarkULIDGeneration benchmarks the generation of ULIDs
func BenchmarkULIDGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ulid.Make()
	}
}

// BenchmarkNanoIDGeneration benchmarks the generation of NanoIDs
func BenchmarkNanoIDGeneration(b *testing.B) {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = nanoid.GenerateString(alphabet, 21) // 21 chars to match ULID length
	}
}

// BenchmarkRandomUint64Generation benchmarks random uint64 generation
func BenchmarkRandomUint64Generation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var b [8]byte
		rand.Read(b[:])
		_ = binary.BigEndian.Uint64(b[:])
	}
}

// BenchmarkMUIDString benchmarks string conversion for MUIDs
func BenchmarkMUIDString(b *testing.B) {
	ids := make([]muid.MUID, b.N)
	for i := range ids {
		ids[i] = muid.Make()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ids[i].String()
	}
}

// BenchmarkUUIDString benchmarks string conversion for UUIDs
func BenchmarkUUIDString(b *testing.B) {
	ids := make([]uuid.UUID, b.N)
	for i := range ids {
		ids[i] = uuid.New()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ids[i].String()
	}
}

// BenchmarkULIDString benchmarks string conversion for ULIDs
func BenchmarkULIDString(b *testing.B) {
	ids := make([]ulid.ULID, b.N)
	for i := range ids {
		ids[i] = ulid.Make()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ids[i].String()
	}
}

// BenchmarkNanoIDString benchmarks string conversion for NanoIDs (no-op since it's already a string)
func BenchmarkNanoIDString(b *testing.B) {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ids := make([]string, b.N)
	for i := range ids {
		ids[i], _ = nanoid.GenerateString(alphabet, 21)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ids[i] // Already a string
	}
}

// BenchmarkConcurrentMUID benchmarks concurrent MUID generation
func BenchmarkConcurrentMUID(b *testing.B) {
	var wg sync.WaitGroup
	workers := 10
	idsPerWorker := b.N / workers

	b.ResetTimer()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < idsPerWorker; i++ {
				_ = muid.Make()
			}
		}()
	}
	wg.Wait()
}

// BenchmarkConcurrentUUID benchmarks concurrent UUID generation
func BenchmarkConcurrentUUID(b *testing.B) {
	var wg sync.WaitGroup
	workers := 10
	idsPerWorker := b.N / workers

	b.ResetTimer()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < idsPerWorker; i++ {
				_ = uuid.New()
			}
		}()
	}
	wg.Wait()
}

// BenchmarkMemoryMUID benchmarks memory allocations for MUID generation
func BenchmarkMemoryMUID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = muid.Make()
	}
}

// BenchmarkMemoryUUID benchmarks memory allocations for UUID generation
func BenchmarkMemoryUUID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = uuid.New()
	}
}

// BenchmarkMemoryULID benchmarks memory allocations for ULID generation
func BenchmarkMemoryULID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ulid.Make()
	}
}

// BenchmarkMemoryNanoID benchmarks memory allocations for NanoID generation
func BenchmarkMemoryNanoID(b *testing.B) {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = nanoid.GenerateString(alphabet, 21)
	}
}
