package hsm_test

import "testing"

func TestRuntimeRulesConformance(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "HSM02/composite_initial_enters_nested_state"},
		{name: "HSM11/completion_event_drives_prioritized_follow_on_transition"},
		{name: "HSM12/explicit_events_or_any_event_only"},
		{name: "HSM13/any_event_only_fires_as_fallback"},
		{name: "HSM14/specific_event_precedes_any_event"},
		{name: "HSM15/first_passing_guard_wins_for_same_event"},
		{name: "HSM16/choice_vertex_models_conditional_branching"},
		{name: "HSM22/external_code_uses_attributes_instead_of_internal_context"},
		{name: "HSM23/get_and_set_expose_stateful_machine_data"},
		{name: "HSM24/context_carries_requests_not_durable_machine_state"},
		{name: "HSM25/external_input_arrives_via_events_or_attributes"},
		{name: "HSM26/event_payload_coordinates_request_response"},
		{name: "HSM27/run_to_completion_serializes_context_access_without_mutexes"},
		{name: "HSM28/events_and_attributes_replace_shared_memory_coordination"},
		{name: "HSM29/set_emits_attribute_reaction_when_value_changes"},
		{name: "HSM30/set_same_value_does_not_emit_onset"},
		{name: "HSM31/onset_drives_attribute_transition"},
		{name: "HSM35/call_triggers_operation_protocol_without_fake_events"},
		{name: "HSM36/hsm_context_scopes_fire_and_forget_dispatch_to_machine_lifetime"},
		{name: "HSM37/background_context_allows_dispatch_to_outlive_machine_lifetime"},
		{name: "HSM38/transient_behavior_context_is_used_only_for_intentional_cancellation"},
		{name: "HSM39/id_is_unique_and_name_is_model_name"},
		{name: "HSM40/name_is_not_used_as_unique_instance_identifier"},
		{name: "HSM41/qualified_name_and_state_report_different_paths"},
		{name: "HSM43/negative_durations_do_not_fire_time_based_triggers"},
		{name: "HSM44/activity_reserved_for_long_running_work"},
		{name: "HSM45/short_synchronous_work_stays_in_entry_exit_or_effect"},
		{name: "HSM46/activity_respects_context_cancellation_on_exit"},
		{name: "HSM47/decompose_behavior_into_separate_state_units"},
		{name: "HSM48/deferred_events_replay_after_state_exit"},
		{name: "HSM51/history_fallback_transition_overrides_parent_initial_on_first_reentry"},
		{name: "HSM52/any_event_guards_filter_internal_lifecycle_events"},
		{name: "HSM53/wait_for_returned_channels_before_post_transition_assertions"},
		{name: "HSM54/after_hooks_are_test_observers_not_production_sync"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Fatal("runtime conformance scenario not implemented yet")
		})
	}
}
