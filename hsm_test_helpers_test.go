package hsm_test

import (
	"slices"
	"sync"
	"testing"

	"github.com/stateforward/hsm.go"
)

type Trace struct {
	sync  []string
	async []string
	mutex *sync.Mutex
}

func (t *Trace) reset() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.sync = []string{}
	t.async = []string{}
}

func (t *Trace) matches(expected Trace) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if expected.sync != nil && !slices.Equal(t.sync, expected.sync) {
		return false
	}
	if expected.async != nil && !slices.Equal(t.async, expected.async) {
		return false
	}
	return true
}

func (t *Trace) contains(expected Trace) bool {
	if expected.sync != nil && slices.ContainsFunc(t.sync, func(s string) bool {
		return slices.Contains(expected.sync, s)
	}) {
		return true
	}
	if expected.async != nil && slices.ContainsFunc(t.async, func(s string) bool {
		return slices.Contains(expected.async, s)
	}) {
		return true
	}
	return false
}

type Event struct{}

type THSM struct {
	hsm.HSM
	foo int
}

type AttrHSM struct {
	hsm.HSM
}

type CallOrderHSM struct {
	hsm.HSM
	mu    sync.Mutex
	order []string
}

func (sm *CallOrderHSM) record(step string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.order = append(sm.order, step)
}

func (sm *CallOrderHSM) orderSnapshot() []string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return append([]string(nil), sm.order...)
}

type CallSigHSM struct {
	hsm.HSM
	mu   sync.Mutex
	hits []string
}

func (sm *CallSigHSM) record(hit string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.hits = append(sm.hits, hit)
}

func (sm *CallSigHSM) hitsSnapshot() []string {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return append([]string(nil), sm.hits...)
}

func (sm *CallSigHSM) methodExpr(arg string) string {
	sm.record("methodExpr")
	return "method:" + arg
}

func assertPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for %s", name)
		}
	}()
	fn()
}

func assertNoPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic for %s: %v", name, r)
		}
	}()
	fn()
}
