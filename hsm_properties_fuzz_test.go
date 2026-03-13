package hsm_test

import "testing"

func FuzzHSMEventScriptProperties(f *testing.F) {
	f.Add("begin|finish")
	f.Add("|begin||finish|finish")

	f.Fuzz(func(t *testing.T, script string) {
		checkHSMEventScriptProperties(t, script)
	})
}

func checkHSMEventScriptProperties(t *testing.T, script string) {
	t.Helper()
	t.Fatalf("TODO: implement deterministic HSM event-script property checks for %q", script)
}
