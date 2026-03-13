package muid

import "testing"

func FuzzMUIDGeneratorProperties(f *testing.F) {
	f.Add((uint64(1)<<defaultConfig.MachineIDBitLen)|0x15, uint64(0), uint64(0), int64(0))
	f.Add(uint64(7), uint64(50), uint64(10), defaultConfig.Epoch)

	f.Fuzz(func(t *testing.T, machineID, timestampBits, machineBits uint64, epoch int64) {
		checkMUIDGeneratorProperties(t, machineID, timestampBits, machineBits, epoch)
	})
}

func checkMUIDGeneratorProperties(t *testing.T, machineID, timestampBits, machineBits uint64, epoch int64) {
	t.Helper()
	t.Fatalf("TODO: implement MUID generator property checks for machineID=%d timestampBits=%d machineBits=%d epoch=%d", machineID, timestampBits, machineBits, epoch)
}
