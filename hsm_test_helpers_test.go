package hsm_test

import (
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stateforward/hsm.go"
)

const (
	waiterDeadline                 = 2 * time.Second
	waiterShouldRemainPendingFor   = 25 * time.Millisecond
	deterministicTimerParkingDelay = 24 * time.Hour
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

type deterministicClockHarness struct {
	mutex         sync.Mutex
	registrations []*timerRegistration
	pending       chan *timerRegistration
}

type timerRegistration struct {
	kind      string
	requested time.Duration
	timer     *time.Timer
	after     chan time.Time
}

type startBarrier struct {
	ready sync.WaitGroup
	start chan struct{}
}

func newStartBarrier(participants int) *startBarrier {
	barrier := &startBarrier{
		start: make(chan struct{}),
	}
	barrier.ready.Add(participants)
	return barrier
}

func (b *startBarrier) participantReadyAndWait() {
	b.ready.Done()
	<-b.start
}

func (b *startBarrier) release(t *testing.T, description string) {
	t.Helper()
	ready := make(chan struct{})
	go func() {
		b.ready.Wait()
		close(ready)
	}()
	awaitWaiter(t, description, ready)
	close(b.start)
}

type runtimeTrace struct {
	mutex   sync.Mutex
	entries []string
}

func (t *runtimeTrace) record(entry string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.entries = append(t.entries, entry)
}

func (t *runtimeTrace) snapshot() []string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return append([]string(nil), t.entries...)
}

func newDeterministicClockHarness() *deterministicClockHarness {
	return &deterministicClockHarness{
		pending: make(chan *timerRegistration, 32),
	}
}

func (h *deterministicClockHarness) Clock() hsm.Clock {
	return hsm.Clock{
		After:    h.After,
		NewTimer: h.NewTimer,
	}
}

func (h *deterministicClockHarness) After(duration time.Duration) <-chan time.Time {
	registration := &timerRegistration{
		kind:      "after",
		requested: duration,
		after:     make(chan time.Time, 1),
	}
	h.register(registration)
	return registration.after
}

func (h *deterministicClockHarness) NewTimer(duration time.Duration) *time.Timer {
	registration := &timerRegistration{
		kind:      "timer",
		requested: duration,
		timer:     time.NewTimer(deterministicTimerParkingDelay),
	}
	h.register(registration)
	return registration.timer
}

func (h *deterministicClockHarness) register(registration *timerRegistration) {
	h.mutex.Lock()
	h.registrations = append(h.registrations, registration)
	h.mutex.Unlock()
	h.pending <- registration
}

func (h *deterministicClockHarness) awaitRegistration(t *testing.T, description string) *timerRegistration {
	t.Helper()
	select {
	case registration := <-h.pending:
		return registration
	case <-time.After(waiterDeadline):
		t.Fatalf("timed out waiting for %s timer registration", description)
		return nil
	}
}

func (h *deterministicClockHarness) assertNoRegistration(t *testing.T, description string) {
	t.Helper()
	select {
	case registration := <-h.pending:
		t.Fatalf("unexpected %s timer registration for %s with duration %v", registration.kind, description, registration.requested)
	case <-time.After(waiterShouldRemainPendingFor):
	}
}

func (h *deterministicClockHarness) registrationCount() int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return len(h.registrations)
}

func (r *timerRegistration) trigger(t *testing.T) {
	t.Helper()
	switch r.kind {
	case "after":
		select {
		case r.after <- time.Now():
		default:
			t.Fatalf("after registration for duration %v already fired", r.requested)
		}
	case "timer":
		deadline := time.After(waiterDeadline)
		for {
			if r.timer.Stop() {
				r.timer.Reset(0)
				return
			}
			select {
			case <-deadline:
				t.Fatalf("timed out waiting to trigger timer with duration %v", r.requested)
			default:
				runtime.Gosched()
			}
		}
	default:
		t.Fatalf("unknown timer registration kind %q", r.kind)
	}
}

func awaitWaiter(t *testing.T, description string, waiter <-chan struct{}) {
	t.Helper()
	select {
	case <-waiter:
	case <-time.After(waiterDeadline):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func assertWaiterPending(t *testing.T, description string, waiter <-chan struct{}) {
	t.Helper()
	select {
	case <-waiter:
		t.Fatalf("expected %s to remain pending", description)
	case <-time.After(waiterShouldRemainPendingFor):
	}
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

func assertPanicContains(t *testing.T, name string, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic for %s containing %q", name, want)
		}
		if got := fmt.Sprint(r); !strings.Contains(got, want) {
			t.Fatalf("panic for %s = %q, want substring %q", name, got, want)
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
