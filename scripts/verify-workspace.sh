#!/usr/bin/env bash

set -euo pipefail

section() {
  printf '\n==> %s\n' "$1"
}

run_package_check() {
  local label="$1"
  local target="$2"

  section "baseline: ${label}"
  go test "$target"
}

section "phase 1 workspace verification"
echo "repo: $(pwd)"

run_package_check "hsm" ./...
run_package_check "kind" ./kind/...
run_package_check "muid" ./muid/...

section "baseline: combined workspace"
go test ./... ./kind/... ./muid/...

section "blocking: race-short workspace"
go test -race -short ./... ./kind/... ./muid/...

section "result"
echo "phase 1 workspace verification passed"
