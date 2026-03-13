package muid

import (
	"strconv"
	"testing"
	"time"
)

func FuzzMUIDGeneratorProperties(f *testing.F) {
	f.Add((uint64(1)<<defaultConfig.MachineIDBitLen)|0x15, uint64(0), uint64(0), int64(0))
	f.Add(uint64(7), uint64(50), uint64(10), defaultConfig.Epoch)

	f.Fuzz(func(t *testing.T, machineID, timestampBits, machineBits uint64, epoch int64) {
		checkMUIDGeneratorProperties(t, machineID, timestampBits, machineBits, epoch)
	})
}

func checkMUIDGeneratorProperties(t *testing.T, machineID, timestampBits, machineBits uint64, epoch int64) {
	t.Helper()

	config := normalizeMUIDPropertyConfig(t, machineID, timestampBits, machineBits, epoch)
	generator := NewGenerator(config, 0, 0)

	if got, want := generator.timestampBitLen, config.TimestampBitLen; got != want {
		t.Fatalf("timestampBitLen = %d, want %d", got, want)
	}
	if got, want := generator.machineIdBitLen, config.MachineIDBitLen; got != want {
		t.Fatalf("machineIdBitLen = %d, want %d", got, want)
	}
	if got, want := generator.epoch, config.Epoch; got != want {
		t.Fatalf("epoch = %d, want %d", got, want)
	}

	expectedMachineID := config.MachineID
	if expectedMachineID == 0 {
		expectedMachineID = defaultConfig.MachineID
	}
	expectedMachineID &= (uint64(1) << config.MachineIDBitLen) - 1
	if got := generator.machineID; got != expectedMachineID {
		t.Fatalf("machineID = %d, want %d", got, expectedMachineID)
	}

	first := generator.ID()
	second := generator.ID()
	if second <= first {
		t.Fatalf("generator IDs must increase, got %d then %d", first, second)
	}
	assertMUIDRoundTrip(t, first)
	assertMUIDRoundTrip(t, second)

	now := time.Now().UnixMilli()
	if config.Epoch > now {
		t.Skip()
	}
	futureTimestamp := uint64(now-config.Epoch) + 5

	regressedCounter := uint64(1)
	if generator.counterBitMask <= 1 {
		regressedCounter = 0
	}
	generator.state.Store((futureTimestamp << generator.counterBitLen) | regressedCounter)
	regressed := generator.ID()
	if got, want := timestampOf(regressed, generator), futureTimestamp; got != want {
		t.Fatalf("regressed timestamp = %d, want %d", got, want)
	}
	if got, want := counterOf(regressed, generator), regressedCounter+1; got != want {
		t.Fatalf("regressed counter = %d, want %d", got, want)
	}
	assertMUIDRoundTrip(t, regressed)

	generator.state.Store((futureTimestamp << generator.counterBitLen) | generator.counterBitMask)
	overflow := generator.ID()
	if got, want := timestampOf(overflow, generator), futureTimestamp+1; got != want {
		t.Fatalf("overflow timestamp = %d, want %d", got, want)
	}
	if got, want := counterOf(overflow, generator), uint64(1); got != want {
		t.Fatalf("overflow counter = %d, want %d", got, want)
	}
	assertMUIDRoundTrip(t, overflow)
}

func normalizeMUIDPropertyConfig(t *testing.T, machineID, timestampBits, machineBits uint64, epoch int64) Config {
	t.Helper()

	if timestampBits > 62 || machineBits > 62 {
		t.Skip()
	}

	config := Config{
		MachineID:       machineID,
		TimestampBitLen: int(timestampBits),
		MachineIDBitLen: int(machineBits),
		Epoch:           epoch,
	}
	if config.TimestampBitLen <= 0 {
		config.TimestampBitLen = defaultConfig.TimestampBitLen
	}
	if config.MachineIDBitLen <= 0 {
		config.MachineIDBitLen = defaultConfig.MachineIDBitLen
	}
	if config.Epoch <= 0 {
		config.Epoch = defaultConfig.Epoch
	}
	if config.TimestampBitLen+config.MachineIDBitLen >= 63 {
		t.Skip()
	}

	return config
}

func assertMUIDRoundTrip(t *testing.T, id MUID) {
	t.Helper()

	parsed, err := strconv.ParseUint(id.String(), 32, 64)
	if err != nil {
		t.Fatalf("failed to parse %q: %v", id.String(), err)
	}
	if got, want := MUID(parsed), id; got != want {
		t.Fatalf("round-trip mismatch: got %d, want %d", got, want)
	}
}
