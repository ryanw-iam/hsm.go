#!/usr/bin/env bash

set -euo pipefail

section() {
  printf '\n==> %s\n' "$1"
}

run_go_test() {
  local label="$1"
  shift

  section "$label"
  go test "$@"
}

section "phase 4 workspace verification"
echo "repo: $(pwd)"

run_go_test "baseline: hsm" ./...
run_go_test "baseline: kind" ./kind/...
run_go_test "baseline: muid" ./muid/...

run_go_test "baseline: combined workspace" ./... ./kind/... ./muid/...

run_go_test "blocking: runtime adversarial suites" . -run 'TestRuntimeAdversarial|TestPublicHelperAdversarial'
run_go_test "blocking: root fuzz-seed replay" . -run 'FuzzHSMEventScriptProperties'
run_go_test "blocking: muid fuzz-seed replay" ./muid/... -run 'FuzzMUIDGeneratorProperties'
run_go_test "blocking: root allocation regression suites" . -run 'Test.*RegressionAllocs'
run_go_test "blocking: muid allocation regression suites" ./muid/... -run 'Test.*RegressionAllocs'
run_go_test "blocking: hsm benchmark smoke" . -run '^$' -bench '^BenchmarkNestedStates_NoEntryExitActivity$' -benchtime=20x -benchmem
run_go_test "blocking: muid benchmark smoke" ./muid -run '^$' -bench '^BenchmarkMUIDGeneration$' -benchtime=20x -benchmem
run_go_test "blocking: race-short workspace" -race -short ./... ./kind/... ./muid/...

section "result"
echo "phase 4 workspace verification passed"
