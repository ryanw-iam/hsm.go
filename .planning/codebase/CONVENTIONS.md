# hsm.go Conventions

## API Naming

- The public DSL uses exported PascalCase Go identifiers such as `Define`, `State`, `Transition`, `Initial`, `Guard`, `Effect`, `After`, `Every`, and `Call`
- Types such as `Model`, `Event`, `HSM`, `Instance`, and `Snapshot` follow normal Go export conventions
- Helper modules keep short package names: `kind` and `muid`

## Authoring Rules

- `rules.md` is an important local convention file and should be treated as part of the package contract
- Rules emphasize explicit initial states, explicit event triggers, explicit choice states, attribute-driven updates via `Set`, and disciplined use of asynchronous behavior
- The rules also discourage hidden state mutation, wildcard trigger misuse, and misuse of completion semantics

## Code Style Signals

- The package uses standard Go imports and package layout rather than internal subpackages
- `hsm.go` mixes public API, internal types, and runtime logic in one file
- Comments are extensive around public APIs and many core internals
- Error handling includes both returned errors and selected panic-based validation paths during model construction or invalid lifecycle operations

## Runtime Conventions

- Callers are expected to interact through events, attributes, and operations rather than directly mutating machine state
- Asynchronous APIs return channels and callers are expected to wait for completion before making assertions, as reinforced by `rules.md`
- `context.Context` is the main lifecycle carrier for dispatch, activities, and machine lifetime

## Testing Conventions

- Tests follow normal Go testing style with `_test.go` files
- The main package tests use external-package style in `hsm_test.go`
- Benchmark coverage is explicit in `hsm_bench_test.go` and `muid/muid_bench_test.go`
- Coverage is uneven across modules, with `kind/kind_test.go` appearing much lighter than `hsm_test.go` or the `muid` suite

## Release And CI Conventions

- GitHub Actions appears to be the current CI path
- `.goreleaser.yaml` governs release metadata
- `.travis.yml` is still present, which suggests some legacy automation has not been cleaned up

## Repo Hygiene Notes

- `rules.md` is currently untracked in git status, so local workflow guidance may not yet be consistently versioned
- The subtree is otherwise compact and conventional for a small Go library workspace
