# hsm.go Architecture

## High-Level Shape

`hsm.go` is organized as one primary runtime package plus two helper modules:

1. `hsm.go` for the DSL, model, runtime, dispatch, and observability helpers
2. `kind` for bit-packed kind/type hierarchy logic
3. `muid` for monotonic unique IDs

## Main Architectural Pattern

The core package in `hsm.go` combines several layers in a single large file:

- element and model definitions
- DSL/builders such as `Define`, `State`, `Transition`, `Initial`, `Guard`, and `After`
- runtime instance implementation
- event dispatch and queue processing
- attribute and operation support
- observer/wait helper functions such as `AfterProcess`, `AfterDispatch`, `AfterEntry`, `AfterExit`, and `AfterExecuted`

## Model And DSL

- The package follows a declarative model-building style described in `README.md`
- `rules.md` adds semantic constraints on how callers should structure models
- The DSL is centered on explicit transitions, explicit initial states, explicit attribute changes, and event-driven interaction

## Runtime Design

- The runtime uses `context.Context` and standard synchronization primitives
- The codebase exposes asynchronous event-dispatch APIs that return channels for completion waiting
- Group and multi-instance behavior is supported through helpers like `DispatchAll` and `DispatchTo`
- The runtime also supports attribute-driven transitions via `Set` and operation-triggered behavior via `Call`

## Helper Modules

### `kind`

- `kind/kind.go` isolates kind encoding and inheritance checks
- This keeps low-level type-hierarchy logic out of the main runtime file

### `muid`

- `muid/muid.go` isolates ID generation logic
- The main runtime can depend on stable unique IDs without embedding ID algorithm details in `hsm.go`

## Architectural Strengths

- The subtree is easy to navigate at a package level because there are only three modules
- The main package presents a coherent surface area for users of the Go API
- Helper concerns with clear boundaries were extracted into `kind` and `muid`

## Architectural Tradeoffs

- Most runtime and DSL behavior is concentrated in `hsm.go`, so internal changes can have broad blast radius
- Reflection and panic-recovery behavior live in the same codebase as core dispatch logic, which raises complexity in runtime-critical paths
- There is no deeper internal package split inside the main module, so architecture is logically layered but physically monolithic
