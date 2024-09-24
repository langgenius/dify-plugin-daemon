package backwards_invocation

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/tester"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func getTestSession() *session_manager.Session {
	return session_manager.NewSession(
		session_manager.NewSessionPayload{
			UserID:                 "test",
			TenantID:               "test",
			PluginUniqueIdentifier: plugin_entities.PluginUniqueIdentifier(""),
			ClusterID:              "test",
			InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_ENDPOINT,
			Action:                 access_types.PLUGIN_ACCESS_ACTION_GET_AI_MODEL_SCHEMAS,
			Declaration:            nil,
			BackwardsInvocation:    tester.NewMockedDifyInvocation(),
			IgnoreCache:            true,
		},
	)
}

func TestBackwardsInvocationAllPermittedPermission(t *testing.T) {
	all_permitted_runtime := plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Resource: plugin_entities.PluginResourceRequirement{
				Permission: &plugin_entities.PluginPermissionRequirement{
					Tool: &plugin_entities.PluginPermissionToolRequirement{
						Enabled: true,
					},
					Model: &plugin_entities.PluginPermissionModelRequirement{
						Enabled:       true,
						LLM:           true,
						TextEmbedding: true,
						Rerank:        true,
						Moderation:    true,
						TTS:           true,
						Speech2text:   true,
					},
					Node: &plugin_entities.PluginPermissionNodeRequirement{
						Enabled: true,
					},
					App: &plugin_entities.PluginPermissionAppRequirement{
						Enabled: true,
					},
				},
			},
		},
	}

	invoke_llm_request := NewBackwardsInvocation(
		dify_invocation.INVOKE_TYPE_LLM,
		"test",
		getTestSession(),
		nil,
		nil,
	)
	if err := checkPermission(&all_permitted_runtime, invoke_llm_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_text_embedding_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_TEXT_EMBEDDING, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_text_embedding_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_rerank_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_RERANK, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_rerank_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_tts_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_TTS, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_tts_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_speech2text_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_SPEECH2TEXT, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_speech2text_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_moderation_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_MODERATION, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_moderation_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_tool_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_TOOL, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_tool_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_node_parameter_extractor_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_NODE_PARAMETER_EXTRACTOR, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_node_parameter_extractor_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_node_question_classifier_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_NODE_QUESTION_CLASSIFIER, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_node_question_classifier_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}

	invoke_app_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_APP, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_permitted_runtime, invoke_app_request); err != nil {
		t.Errorf("checkPermission failed: %s", err.Error())
	}
}

func TestBackwardsInvocationAllDeniedPermission(t *testing.T) {
	all_denied_runtime := plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Resource: plugin_entities.PluginResourceRequirement{},
		},
	}

	invoke_llm_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_LLM, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_llm_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_text_embedding_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_TEXT_EMBEDDING, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_text_embedding_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_rerank_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_RERANK, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_rerank_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_tts_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_TTS, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_tts_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_speech2text_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_SPEECH2TEXT, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_speech2text_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_moderation_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_MODERATION, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_moderation_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_tool_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_TOOL, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_tool_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_node_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_NODE_PARAMETER_EXTRACTOR, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_node_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_node_question_classifier_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_NODE_QUESTION_CLASSIFIER, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_node_question_classifier_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}

	invoke_app_request := NewBackwardsInvocation(dify_invocation.INVOKE_TYPE_APP, "", getTestSession(), nil, nil)
	if err := checkPermission(&all_denied_runtime, invoke_app_request); err == nil {
		t.Errorf("checkPermission failed: expected error, got nil")
	}
}
