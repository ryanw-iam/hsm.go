package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/stateforward/hsm.go"
)

type TrafficLight struct {
	hsm.HSM
	maintenanceMode bool
	carsWaiting     int
	timer           int
}

var (
	TimerEvent        = hsm.Event{Name: "TimerEvent"}
	CarArrival        = hsm.Event{Name: "CarArrival"}
	MaintenanceSwitch = hsm.Event{Name: "MaintenanceSwitch"}
	PedestrianButton  = hsm.Event{Name: "PedestrianButton"}
	Tick              = hsm.Event{Name: "Tick"}
)

func resetCars(_ context.Context, t *TrafficLight, _ hsm.Event) {
	t.carsWaiting = 0
}

func addCar(_ context.Context, t *TrafficLight, _ hsm.Event) {
	t.carsWaiting++
}

func noCarsWaiting(_ context.Context, t *TrafficLight, _ hsm.Event) bool {
	return t.carsWaiting == 0
}

func isMaintenance(_ context.Context, t *TrafficLight, _ hsm.Event) bool {
	return t.maintenanceMode
}

func isNotMaintenance(_ context.Context, t *TrafficLight, _ hsm.Event) bool {
	return !t.maintenanceMode
}

func checkCarsForChoice(_ context.Context, t *TrafficLight, _ hsm.Event) bool {
	return t.carsWaiting > 10
}

func setTimerExtended(_ context.Context, t *TrafficLight, _ hsm.Event) {
	t.timer = 60
}

func setTimerStandard(_ context.Context, t *TrafficLight, _ hsm.Event) {
	t.timer = 40
}

func maintenanceTick(_ context.Context, t *TrafficLight, _ hsm.Event) {
	t.timer++
}

var model = hsm.Define("TrafficLight",
	hsm.Initial(hsm.Target("operational")),

	hsm.State("operational",
		hsm.Transition(
			hsm.On(MaintenanceSwitch),
			hsm.Guard(isMaintenance),
			hsm.Target("../maintenance"),
		),
		hsm.Initial(hsm.Target("red")),

		hsm.State("red",
			hsm.Transition(
				hsm.On(TimerEvent),
				hsm.Guard(checkCarsForChoice),
				hsm.Effect(setTimerExtended),
				hsm.Target("../green"),
			),
			hsm.Transition(
				hsm.On(TimerEvent),
				hsm.Effect(setTimerStandard),
				hsm.Target("../green"),
			),
			hsm.Transition(
				hsm.On(CarArrival),
				hsm.Effect(addCar),
			),
		),

		hsm.State("green",
			hsm.Transition(
				hsm.On(TimerEvent),
				hsm.Target("../yellow"),
			),
			hsm.Transition(
				hsm.On(PedestrianButton),
				hsm.Guard(noCarsWaiting),
				hsm.Target("../yellow"),
			),
		),

		hsm.State("yellow",
			hsm.Defer(CarArrival.Name),
			hsm.Transition(
				hsm.On(TimerEvent),
				hsm.Target("../red"),
			),
		),
	),

	hsm.State("maintenance",
		hsm.Entry(resetCars),
		hsm.Transition(
			hsm.On(Tick),
			hsm.Effect(maintenanceTick),
		),
		hsm.Transition(
			hsm.On(MaintenanceSwitch),
			hsm.Guard(isNotMaintenance),
			hsm.Target("../operational"),
		),
	),
)

type BenchmarkResult struct {
	Language            string  `json:"language"`
	Iterations          int     `json:"iterations"`
	DurationMs          int64   `json:"duration_ms"`
	MemoryMb            float64 `json:"memory_mb"`
	ThroughputOpsPerSec int     `json:"throughput_ops_per_sec"`
}

func envInt(name string, defaultValue int) int {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}

func envBool(name string) bool {
	value, ok := os.LookupEnv(name)
	if !ok {
		return false
	}
	return value != "" && value != "0" && value != "false" && value != "False"
}

func assertTrafficLight(sm *TrafficLight, state string, carsWaiting int, timer int, step string) {
	if sm.State() != state {
		panic(fmt.Sprintf("%s: state %q, expected %q", step, sm.State(), state))
	}
	if sm.carsWaiting != carsWaiting {
		panic(fmt.Sprintf("%s: carsWaiting %d, expected %d", step, sm.carsWaiting, carsWaiting))
	}
	if sm.timer != timer {
		panic(fmt.Sprintf("%s: timer %d, expected %d", step, sm.timer, timer))
	}
}

func validateTrafficLight(ctx context.Context) {
	light := &TrafficLight{}
	sm := hsm.Started(ctx, light, &model)
	if sm == nil {
		panic("validation HSM not started")
	}
	assertTrafficLight(sm, "/TrafficLight/operational/red", 0, 0, "initial")

	completion := hsm.Dispatch(ctx, sm, CarArrival)
	if completion == nil {
		panic("dispatch did not return a blockable completion")
	}
	<-completion
	assertTrafficLight(sm, "/TrafficLight/operational/red", 1, 0, "after CarArrival")

	<-hsm.Dispatch(ctx, sm, TimerEvent)
	assertTrafficLight(sm, "/TrafficLight/operational/green", 1, 40, "after first TimerEvent")

	<-hsm.Dispatch(ctx, sm, TimerEvent)
	assertTrafficLight(sm, "/TrafficLight/operational/yellow", 1, 40, "after second TimerEvent")

	<-hsm.Dispatch(ctx, sm, TimerEvent)
	assertTrafficLight(sm, "/TrafficLight/operational/red", 1, 40, "after third TimerEvent")
}

func dispatchBatch(ctx context.Context, sm *TrafficLight, cycles int) {
	for i := 0; i < cycles; i++ {
		<-hsm.Dispatch(ctx, sm, CarArrival)
		<-hsm.Dispatch(ctx, sm, TimerEvent)
		<-hsm.Dispatch(ctx, sm, TimerEvent)
		<-hsm.Dispatch(ctx, sm, TimerEvent)
	}
}

func calibrateBatch(ctx context.Context, sm *TrafficLight) int {
	const targetBatch = 10 * time.Millisecond
	cycles := 1
	for {
		start := time.Now()
		dispatchBatch(ctx, sm, cycles)
		if time.Since(start) >= targetBatch || cycles >= 1<<20 {
			return cycles
		}
		cycles *= 2
	}
}

func main() {
	warmupMs := envInt("HSM_BENCH_WARMUP_MS", 250)
	durationMs := envInt("HSM_BENCH_DURATION_MS", 2000)
	ctx := context.Background()
	if envBool("HSM_BENCH_VALIDATE") {
		validateTrafficLight(ctx)
	}

	lightWarmup := &TrafficLight{}
	mWarmup := hsm.Started(ctx, lightWarmup, &model)
	if mWarmup == nil {
		panic("warmup HSM not started")
	}
	batchCycles := calibrateBatch(ctx, mWarmup)
	warmupDeadline := time.Now().Add(time.Duration(warmupMs) * time.Millisecond)
	for time.Now().Before(warmupDeadline) {
		dispatchBatch(ctx, mWarmup, batchCycles)
	}

	light := &TrafficLight{}
	m := hsm.Started(ctx, light, &model)
	if m == nil {
		panic("benchmark HSM not started")
	}

	runtime.GC()

	start := time.Now()
	deadline := start.Add(time.Duration(durationMs) * time.Millisecond)
	completedCycles := 0
	for time.Now().Before(deadline) {
		dispatchBatch(ctx, m, batchCycles)
		completedCycles += batchCycles
	}
	duration := time.Since(start)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	totalDispatches := completedCycles * 4
	durationSecs := duration.Seconds()
	opsPerSec := 0
	if durationSecs > 0 {
		opsPerSec = int(float64(totalDispatches) / durationSecs)
	}

	memMb := float64(memStats.Alloc) / (1024 * 1024)

	res := BenchmarkResult{
		Language:            "Go",
		Iterations:          totalDispatches,
		DurationMs:          duration.Milliseconds(),
		MemoryMb:            memMb,
		ThroughputOpsPerSec: opsPerSec,
	}

	out, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	os.Exit(0)
}
