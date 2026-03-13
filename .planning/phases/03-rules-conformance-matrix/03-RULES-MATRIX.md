# Phase 3 Rules Conformance Matrix

`rules.md` remains the canonical wording for the contract. This matrix adds the honest enforcement class, the planned evidence target, and a short implementation note for each `HSMxx` rule.

| Rule | Enforcement | Evidence | Notes |
| --- | --- | --- | --- |
| HSM01 | define_panic | `TestModelRulesConformance/HSM01/top_level_initial_required` | `Define(...)` panics when the top-level machine has no initial state. |
| HSM02 | runtime_semantic | `TestRuntimeRulesConformance/HSM02/composite_initial_enters_nested_state` | Auto-entry into nested substates is a runtime behavior proven by a composite state with an explicit initial transition. |
| HSM03 | define_panic | `TestModelRulesConformance/HSM03/top_level_entry_actions_rejected` | Top-level entry actions are rejected during model construction. |
| HSM04 | define_panic | `TestModelRulesConformance/HSM04/top_level_exit_actions_rejected` | Top-level exit actions are rejected during model construction. |
| HSM05 | define_panic | `TestModelRulesConformance/HSM05/initial_target_must_be_nested` | Initial pseudostates panic when their target is not nested under the owning state or model. |
| HSM06 | define_panic | `TestModelRulesConformance/HSM06/initial_transition_cannot_have_guard` | Initial transitions reject guards at definition time. |
| HSM07 | define_panic | `TestModelRulesConformance/HSM07/initial_pseudostate_cannot_have_multiple_transitions` | Initial pseudostates may only keep one outgoing transition. |
| HSM08 | define_panic | `TestModelRulesConformance/HSM08/transition_members_must_be_declared_inside_transition` | Transition members such as `Source`, `Target`, `On`, `OnSet`, `OnCall`, `After`, `Every`, `When`, `Guard`, and `Effect` panic outside `Transition(...)`. |
| HSM09 | define_panic | `TestModelRulesConformance/HSM09/state_behaviors_must_be_declared_inside_state` | `Entry`, `Exit`, `Activity`, and `Defer` panic outside `State(...)`. |
| HSM10 | define_panic | `TestModelRulesConformance/HSM10/implicit_completion_transition_rejected` | Triggerless completion transitions are not implemented and fail during transition definition. |
| HSM11 | runtime_semantic | `TestRuntimeRulesConformance/HSM11/completion_event_drives_prioritized_follow_on_transition` | Completion events have concrete runtime priority semantics that should be asserted directly. |
| HSM12 | exemplar | `TestRuntimeRulesConformance/HSM12/explicit_events_or_any_event_only` | Wildcard-pattern misuse is not rejected mechanically today, so the evidence is a positive trigger-definition example. |
| HSM13 | runtime_semantic | `TestRuntimeRulesConformance/HSM13/any_event_only_fires_as_fallback` | `AnyEvent` only runs after explicit event transitions fail to match or pass guards. |
| HSM14 | runtime_semantic | `TestRuntimeRulesConformance/HSM14/specific_event_precedes_any_event` | Specific event transitions take precedence over `AnyEvent` at runtime. |
| HSM15 | runtime_semantic | `TestRuntimeRulesConformance/HSM15/first_passing_guard_wins_for_same_event` | Transition order for the same event is observable because the first passing guard wins. |
| HSM16 | exemplar | `TestRuntimeRulesConformance/HSM16/choice_vertex_models_conditional_branching` | The library cannot reject ad hoc branching outside `Choice(...)`, so the evidence is a positive choice-based branching example. |
| HSM17 | define_panic | `TestModelRulesConformance/HSM17/choice_requires_outgoing_transitions` | `Choice(...)` panics when defined without any outgoing transitions. |
| HSM18 | define_panic | `TestModelRulesConformance/HSM18/choice_default_branch_must_be_last` | Choice pseudostates reject a guarded final branch because the last branch must be the default. |
| HSM19 | define_panic | `TestModelRulesConformance/HSM19/source_and_target_paths_must_resolve` | Invalid relative or absolute source/target paths panic during model construction. |
| HSM20 | define_panic | `TestModelRulesConformance/HSM20/top_level_transition_requires_both_ends_or_neither` | Top-level transitions reject a source-only or target-only definition. |
| HSM21 | define_panic | `TestModelRulesConformance/HSM21/internal_transition_requires_effect` | Internal transitions with no target must define an effect. |
| HSM22 | exemplar | `TestRuntimeRulesConformance/HSM22/external_code_uses_attributes_instead_of_internal_context` | External reads and writes against internal machine fields are not blocked directly, so the evidence is a positive attribute-based usage example. |
| HSM23 | exemplar | `TestRuntimeRulesConformance/HSM23/get_and_set_expose_stateful_machine_data` | Attribute access is a recommended API pattern rather than a rejected misuse case. |
| HSM24 | exemplar | `TestRuntimeRulesConformance/HSM24/context_carries_requests_not_durable_machine_state` | `context.Context` misuse is guidance-only today, so the evidence is a positive separation-of-concerns example. |
| HSM25 | exemplar | `TestRuntimeRulesConformance/HSM25_external_input_arrives_via_events_or_attributes` | External field mutation cannot be rejected generically, so the evidence demonstrates the supported event/attribute flow. |
| HSM26 | exemplar | `TestRuntimeRulesConformance/HSM26/event_payload_coordinates_request_response` | Request/response coordination through event payloads is an API-usage exemplar, not a panic-enforced rule. |
| HSM27 | exemplar | `TestRuntimeRulesConformance/HSM27/run_to_completion_serializes_context_access_without_mutexes` | Extra mutexes are architectural misuse the library does not detect, so the evidence is a positive RTC example. |
| HSM28 | exemplar | `TestRuntimeRulesConformance/HSM28/events_and_attributes_replace_shared_memory_coordination` | Preference for events/attributes over shared memory is advisory and should be shown by example only. |
| HSM29 | runtime_semantic | `TestRuntimeRulesConformance/HSM29/set_emits_attribute_reaction_when_value_changes` | `Set(...)` has concrete change-notification semantics that can be asserted with waiter-based tests. |
| HSM30 | runtime_semantic | `TestRuntimeRulesConformance/HSM30/set_same_value_does_not_emit_onset` | Same-value `Set(...)` calls intentionally do not emit `OnSet` reactions. |
| HSM31 | runtime_semantic | `TestRuntimeRulesConformance/HSM31/onset_drives_attribute_transition` | Attribute-driven behavior is observable via `OnSet("name")` transitions. |
| HSM32 | define_panic | `TestModelRulesConformance/HSM32/named_model_members_must_not_be_empty` | Empty attribute, operation, `OnSet`, and `OnCall` names panic immediately. |
| HSM33 | define_panic | `TestModelRulesConformance/HSM33/duplicate_attributes_and_operations_rejected` | Duplicate attribute and operation declarations are rejected during definition. |
| HSM34 | define_panic | `TestModelRulesConformance/HSM34/oncall_requires_declared_operation` | `OnCall(...)` panics when the named operation has not been declared. |
| HSM35 | runtime_semantic | `TestRuntimeRulesConformance/HSM35/call_triggers_operation_protocol_without_fake_events` | `Call(...)` has concrete runtime semantics for operation dispatch and result handling. |
| HSM36 | exemplar | `TestRuntimeRulesConformance/HSM36/hsm_context_scopes_fire_and_forget_dispatch_to_machine_lifetime` | Choosing `hsm.Context()` is a lifetime-management pattern the library cannot label as misuse automatically. |
| HSM37 | exemplar | `TestRuntimeRulesConformance/HSM37/background_context_allows_dispatch_to_outlive_machine_lifetime` | Background-context lifetime choice is API guidance backed by a positive example. |
| HSM38 | exemplar | `TestRuntimeRulesConformance/HSM38/transient_behavior_context_is_used_only_for_intentional_cancellation` | Whether a transient behavior context is appropriate depends on caller intent, so only a positive exemplar is honest. |
| HSM39 | runtime_semantic | `TestRuntimeRulesConformance/HSM39/id_is_unique_and_name_is_model_name` | `ID()` uniqueness and `Name()` model-name reporting are concrete observable behaviors. |
| HSM40 | exemplar | `TestRuntimeRulesConformance/HSM40/name_is_not_used_as_unique_instance_identifier` | Treating `Name()` as unique is caller misuse that the library does not reject directly. |
| HSM41 | runtime_semantic | `TestRuntimeRulesConformance/HSM41/qualified_name_and_state_report_different_paths` | `QualifiedName()` and `State()` expose different path semantics that can be asserted directly. |
| HSM42 | define_panic | `TestModelRulesConformance/HSM42/time_based_triggers_require_real_state_source` | `After`, `Every`, and `When` panic when attached to transitions whose source is not a real state. |
| HSM43 | runtime_semantic | `TestRuntimeRulesConformance/HSM43/negative_durations_do_not_fire_time_based_triggers` | Negative-duration timers have deterministic non-firing semantics. |
| HSM44 | exemplar | `TestRuntimeRulesConformance/HSM44/activity_reserved_for_long_running_work` | Appropriate `Activity(...)` granularity is design guidance, so the evidence is a positive long-running activity example. |
| HSM45 | exemplar | `TestRuntimeRulesConformance/HSM45/short_synchronous_work_stays_in_entry_exit_or_effect` | Misplacing short synchronous work in `Activity(...)` is not rejected directly today. |
| HSM46 | runtime_semantic | `TestRuntimeRulesConformance/HSM46/activity_respects_context_cancellation_on_exit` | Clean shutdown of long-running activity code is a runtime behavior with deterministic waiter evidence. |
| HSM47 | exemplar | `TestRuntimeRulesConformance/HSM47/decompose_behavior_into_separate_state_units` | Avoiding state explosion is architectural guidance rather than a mechanically enforced rule. |
| HSM48 | runtime_semantic | `TestRuntimeRulesConformance/HSM48/deferred_events_replay_after_state_exit` | Deferred-event replay is an explicit runtime semantic already suited to deterministic tests. |
| HSM49 | define_panic | `TestModelRulesConformance/HSM49/final_state_cannot_have_behaviors_or_transitions` | Final states reject transitions, activities, entry actions, and exit actions at definition time. |
| HSM50 | define_panic | `TestModelRulesConformance/HSM50/history_pseudostates_must_be_nested_inside_state` | History pseudostates panic when declared at the top level instead of inside a composite state. |
| HSM51 | runtime_semantic | `TestRuntimeRulesConformance/HSM51/history_fallback_transition_overrides_parent_initial_on_first_reentry` | History fallback behavior is a concrete runtime semantic that should be asserted with first-entry and re-entry scenarios. |
| HSM52 | runtime_semantic | `TestRuntimeRulesConformance/HSM52/any_event_guards_filter_internal_lifecycle_events` | Guarding `AnyEvent` against internal lifecycle events is an observable runtime behavior. |
| HSM53 | runtime_semantic | `TestRuntimeRulesConformance/HSM53/wait_for_returned_channels_before_post_transition_assertions` | Returned waiter channels define the deterministic synchronization contract for settled-state assertions. |
| HSM54 | exemplar | `TestRuntimeRulesConformance/HSM54/after_hooks_are_test_observers_not_production_sync` | The helper waiters exist, but avoiding them as production synchronization is caller guidance rather than a rejected misuse. |
