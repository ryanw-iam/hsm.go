package muid

import (
	"strconv"
	"testing"
	"time"
)

func timestampOf(id MUID, g *Generator) uint64 {
	return uint64(id) >> g.timestampBitShift
}

func counterOf(id MUID, g *Generator) uint64 {
	return uint64(id) & g.counterBitMask
}

func TestNewGeneratorAppliesDefaultsMasksMachineIDAndIsMonotonic(t *testing.T) {
	config := Config{
		MachineID:       (uint64(1) << defaultConfig.MachineIDBitLen) | 0x15,
		TimestampBitLen: 0,
		MachineIDBitLen: 0,
		Epoch:           0,
	}

	generator := NewGenerator(config, 0, 0)

	if got, want := generator.timestampBitLen, defaultConfig.TimestampBitLen; got != want {
		t.Fatalf("timestampBitLen = %d, want %d", got, want)
	}
	if got, want := generator.machineIdBitLen, defaultConfig.MachineIDBitLen; got != want {
		t.Fatalf("machineIdBitLen = %d, want %d", got, want)
	}
	if got, want := generator.epoch, defaultConfig.Epoch; got != want {
		t.Fatalf("epoch = %d, want %d", got, want)
	}

	wantMachineID := config.MachineID & ((uint64(1) << generator.machineIdBitLen) - 1)
	if got := generator.machineID; got != wantMachineID {
		t.Fatalf("machineID = %d, want %d", got, wantMachineID)
	}

	first := generator.ID()
	second := generator.ID()
	if second <= first {
		t.Fatalf("generator IDs must increase, got %d then %d", first, second)
	}
}

func TestGeneratorIDClampsClockRegressionAndAdvancesOnCounterOverflow(t *testing.T) {
	generator := NewGenerator(Config{
		MachineID:       7,
		TimestampBitLen: defaultConfig.TimestampBitLen,
		MachineIDBitLen: defaultConfig.MachineIDBitLen,
		Epoch:           defaultConfig.Epoch,
	}, 0, 0)

	futureTimestamp := uint64(time.Now().UnixMilli()-generator.epoch) + 5

	generator.state.Store((futureTimestamp << generator.counterBitLen) | 7)
	regressed := generator.ID()
	if got := timestampOf(regressed, generator); got != futureTimestamp {
		t.Fatalf("regressed timestamp = %d, want %d", got, futureTimestamp)
	}
	if got := counterOf(regressed, generator); got != 8 {
		t.Fatalf("regressed counter = %d, want 8", got)
	}

	generator.state.Store((futureTimestamp << generator.counterBitLen) | generator.counterBitMask)
	overflow := generator.ID()
	if got := timestampOf(overflow, generator); got != futureTimestamp+1 {
		t.Fatalf("overflow timestamp = %d, want %d", got, futureTimestamp+1)
	}
	if got := counterOf(overflow, generator); got != 1 {
		t.Fatalf("overflow counter = %d, want 1", got)
	}
}

func TestDefaultMakeAndMakeStringAreUniqueAndMonotonic(t *testing.T) {
	const total = 512

	lastID := Make()
	seenIDs := map[MUID]struct{}{lastID: {}}
	for i := 1; i < total; i++ {
		next := Make()
		if next <= lastID {
			t.Fatalf("Make() must be strictly increasing, got %d after %d", next, lastID)
		}
		if _, ok := seenIDs[next]; ok {
			t.Fatalf("Make() collision detected for %d", next)
		}
		seenIDs[next] = struct{}{}
		lastID = next
	}

	var lastStringValue uint64
	seenStrings := make(map[string]struct{}, total)
	for i := 0; i < total; i++ {
		value := MakeString()
		if _, ok := seenStrings[value]; ok {
			t.Fatalf("MakeString() collision detected for %q", value)
		}
		seenStrings[value] = struct{}{}

		parsed, err := strconv.ParseUint(value, 32, 64)
		if err != nil {
			t.Fatalf("MakeString() returned invalid base32 %q: %v", value, err)
		}
		if i > 0 && parsed <= lastStringValue {
			t.Fatalf("MakeString() must be strictly increasing, got %d after %d", parsed, lastStringValue)
		}
		lastStringValue = parsed
	}
}
