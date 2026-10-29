package backwards_invocation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
)

type invocationJSONWriter struct {
	session *session_manager.Session
	frames  [][]byte
	done    chan struct{}
}

func (w *invocationJSONWriter) Write(event session_manager.PLUGIN_IN_STREAM_EVENT, data any) error {
	w.frames = append(w.frames, w.session.Message(event, data))
	return nil
}

func (w *invocationJSONWriter) Done() { close(w.done) }

type invocationJSONEvent struct {
	Event   RequestEvent    `json:"event"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func invokeJSON(t *testing.T, wire string, from access_types.PluginAccessType, allowed bool) ([]invocationJSONEvent, error) {
	t.Helper()
	routine.InitPool(4)
	declaration := &plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Resource: plugin_entities.PluginResourceRequirement{
				Permission: &plugin_entities.PluginPermissionRequirement{
					Tool: &plugin_entities.PluginPermissionToolRequirement{Enabled: allowed},
					Model: &plugin_entities.PluginPermissionModelRequirement{
						Enabled: allowed, LLM: allowed, TextEmbedding: allowed, Rerank: allowed,
						TTS: allowed, Speech2text: allowed, Moderation: allowed,
					},
					Node:    &plugin_entities.PluginPermissionNodeRequirement{Enabled: allowed},
					App:     &plugin_entities.PluginPermissionAppRequirement{Enabled: allowed},
					Storage: &plugin_entities.PluginPermissionStorageRequirement{Enabled: allowed, Size: 100},
				},
			},
		},
	}
	session := &session_manager.Session{ID: "session", TenantID: `trusted"tenant`, UserID: "trusted-user"}
	writer := &invocationJSONWriter{session: session, done: make(chan struct{})}
	err := InvokeDify(declaration, from, session, writer, []byte(wire))
	if err != nil {
		require.Empty(t, writer.frames)
		return nil, err
	}
	select {
	case <-writer.done:
	case <-time.After(5 * time.Second):
		t.Fatal("invocation did not send an end frame")
	}
	var events []invocationJSONEvent
	for _, frame := range writer.frames {
		var envelope struct {
			Event session_manager.PLUGIN_IN_STREAM_EVENT `json:"event"`
			Data  invocationJSONEvent                    `json:"data"`
		}
		require.NoError(t, json.Unmarshal(frame, &envelope))
		require.Equal(t, session_manager.PLUGIN_IN_STREAM_EVENT_RESPONSE, envelope.Event)
		events = append(events, envelope.Data)
	}
	require.NotEmpty(t, events)
	require.Equal(t, REQUEST_EVENT_END, events[len(events)-1].Event)
	return events, nil
}

func echoDispatch[T any](handle *BackwardsInvocation) {
	genericDispatchTask(handle, func(handle *BackwardsInvocation, request *T) {
		handle.WriteResponse("struct", request)
	})
}

func replaceDispatch(t *testing.T, kind dify_invocation.InvokeType, dispatch func(*BackwardsInvocation)) {
	t.Helper()
	previous := dispatchMapping[kind]
	dispatchMapping[kind] = dispatch
	t.Cleanup(func() { dispatchMapping[kind] = previous })
}

func invokeJSONRequest(t *testing.T, kind dify_invocation.InvokeType, request string) json.RawMessage {
	t.Helper()
	wire := fmt.Sprintf(`{"type":%q,"backwards_request_id":"request","request":%s}`, kind, request)
	events, err := invokeJSON(t, wire, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, true)
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, REQUEST_EVENT_RESPONSE, events[0].Event, events[0].Message)
	return events[0].Data
}

func TestInvokeDifyPreservesOpaqueJSON(t *testing.T) {
	replaceDispatch(t, dify_invocation.INVOKE_TYPE_LLM, echoDispatch[dify_invocation.InvokeLLMRequest])
	values := []string{
		`null`, `{}`, `[]`, `""`, `"signature"`, `false`, `true`, `0`, `9007199254740993`,
		`[49]`, `[{"id":9007199254740993}]`,
		`{"anthropic_content":[{"type":"thinking","thinking":"","signature":"original"},{"type":"tool_use","input":{"id":9007199254740993}}]}`,
		`{"responses_output":[{"type":"reasoning","encrypted_content":"original","summary":[]}]}`,
	}
	for _, stream := range []bool{false, true} {
		for _, opaque := range values {
			t.Run(fmt.Sprintf("stream=%t/%s", stream, opaque), func(t *testing.T) {
				payload := fmt.Sprintf(`{"provider":"p","model":"m","mode":"chat","stream":%t,"prompt_messages":[{"role":"assistant","content":null,"opaque_body":%s}]}`, stream, opaque)
				response := invokeJSONRequest(t, dify_invocation.INVOKE_TYPE_LLM, payload)
				var request struct {
					Stream         bool `json:"stream"`
					PromptMessages []struct {
						Content json.RawMessage `json:"content"`
						Opaque  json.RawMessage `json:"opaque_body"`
					} `json:"prompt_messages"`
				}
				require.NoError(t, json.Unmarshal(response, &request))
				require.Equal(t, stream, request.Stream)
				require.Equal(t, "null", string(request.PromptMessages[0].Content))
				// Compare JSON bytes without decoding numbers to float64.
				var compact bytes.Buffer
				require.NoError(t, json.Compact(&compact, []byte(opaque)))
				require.Equal(t, compact.String(), string(request.PromptMessages[0].Opaque))
			})
		}
	}
}

func TestInvokeDifyPreservesParametersAndTrustedIdentity(t *testing.T) {
	replaceDispatch(t, dify_invocation.INVOKE_TYPE_LLM, echoDispatch[dify_invocation.InvokeLLMRequest])
	payload := `{"provider":"p","model":"m","mode":"chat","tenant_id":"spoof","TENANT_ID":"spoof","user_id":"spoof","uſer_id":"spoof","TYPE":"tool","type":"tool","completion_params":{"seed":9007199254740993},"tools":[{"name":"f","parameters":{"enum":[9007199254740993]}}],"prompt_messages":[{"role":"user","content":[{"type":"image","url":"https://example.com/image","opaque_body":{"id":9007199254740993}}]},{"role":"assistant","tool_calls":[{"id":"c","type":"function","function":{"name":"f","arguments":"{\"id\":9007199254740993}"}}]}]}`
	response := invokeJSONRequest(t, dify_invocation.INVOKE_TYPE_LLM, payload)
	var request struct {
		TenantID string `json:"tenant_id"`
		UserID   string `json:"user_id"`
		Type     string `json:"type"`
	}
	require.NoError(t, json.Unmarshal(response, &request))
	require.Equal(t, `trusted"tenant`, request.TenantID)
	require.Equal(t, "trusted-user", request.UserID)
	require.Equal(t, "llm", request.Type)
	require.Contains(t, string(response), `"seed":9007199254740993`)
	require.Contains(t, string(response), `"enum":[9007199254740993]`)
	require.Contains(t, string(response), `"opaque_body":{"id":9007199254740993}`)
	require.Contains(t, string(response), `"arguments":"{\"id\":9007199254740993}"`)
	require.Contains(t, string(response), `"role":"assistant","content":null`)
}

func TestInvokeDifyAllDispatchTypes(t *testing.T) {
	cases := []struct {
		kind     dify_invocation.InvokeType
		payload  string
		dispatch func(*BackwardsInvocation)
	}{
		{"llm", `{"provider":"p","model":"m","mode":"chat","prompt_messages":[{"role":"user","content":"hi"}]}`, echoDispatch[dify_invocation.InvokeLLMRequest]},
		{"llm_structured_output", `{"provider":"p","model":"m","mode":"chat","prompt_messages":[{"role":"user","content":"hi"}],"structured_output_schema":{"type":"object"}}`, echoDispatch[dify_invocation.InvokeLLMWithStructuredOutputRequest]},
		{"text_embedding", `{"provider":"p","model":"m","texts":["hi"],"input_type":"query"}`, echoDispatch[dify_invocation.InvokeTextEmbeddingRequest]},
		{"multimodal_embedding", `{"provider":"p","model":"m","documents":[{"content":"hi","content_type":"text"}],"input_type":"query"}`, echoDispatch[dify_invocation.InvokeMultimodalEmbeddingRequest]},
		{"rerank", `{"provider":"p","model":"m","query":"hi","docs":["doc"],"top_n":1,"score_threshold":0.3}`, echoDispatch[dify_invocation.InvokeRerankRequest]},
		{"multimodal_rerank", `{"provider":"p","model":"m","query":{"content":"hi","content_type":"text"},"docs":[{"content":"hi","content_type":"text"}],"top_n":1}`, echoDispatch[dify_invocation.InvokeMultimodalRerankRequest]},
		{"tts", `{"provider":"p","model":"m","content_text":"hi","voice":"test"}`, echoDispatch[dify_invocation.InvokeTTSRequest]},
		{"speech2text", `{"provider":"p","model":"m","file":"abcd"}`, echoDispatch[dify_invocation.InvokeSpeech2TextRequest]},
		{"moderation", `{"provider":"p","model":"m","text":"hi"}`, echoDispatch[dify_invocation.InvokeModerationRequest]},
		{"tool", `{"provider":"p","tool_type":"builtin","tool":"test","tool_parameters":{"id":9007199254740993}}`, echoDispatch[dify_invocation.InvokeToolRequest]},
		{"app", `{"app_id":"a","inputs":{"id":9007199254740993},"response_mode":"streaming","query":"hi"}`, echoDispatch[dify_invocation.InvokeAppRequest]},
		{"node_parameter_extractor", `{"parameters":[{"name":"x","type":"number"}],"model":{"provider":"p","name":"m","mode":"chat"},"query":"hi"}`, echoDispatch[dify_invocation.InvokeParameterExtractorRequest]},
		{"node_question_classifier", `{"classes":[{"id":"x","name":"test"}],"model":{"provider":"p","name":"m","mode":"chat"},"query":"hi"}`, echoDispatch[dify_invocation.InvokeQuestionClassifierRequest]},
		{"storage", `{"opt":"set","key":"key","value":"abcd"}`, echoDispatch[dify_invocation.InvokeStorageRequest]},
		{"system_summary", `{"text":"hi"}`, echoDispatch[dify_invocation.InvokeSummaryRequest]},
		{"upload_file", `{"filename":"hi","mimetype":"text/plain"}`, echoDispatch[dify_invocation.UploadFileRequest]},
		{"fetch_app", `{"app_id":"a"}`, echoDispatch[dify_invocation.FetchAppRequest]},
	}
	require.Len(t, dispatchMapping, len(cases))
	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			require.Contains(t, dispatchMapping, tc.kind)
			replaceDispatch(t, tc.kind, tc.dispatch)
			response := invokeJSONRequest(t, tc.kind, tc.payload)
			if tc.kind == "tool" || tc.kind == "app" {
				require.Contains(t, string(response), `"id":9007199254740993`)
			}
		})
	}
}

func TestInvokeDifyRejectsInvalidRequests(t *testing.T) {
	replaceDispatch(t, dify_invocation.INVOKE_TYPE_LLM, echoDispatch[dify_invocation.InvokeLLMRequest])
	for _, wire := range []string{
		`{`, `null`, `[]`, `{}`, `{"type":null}`, `{"type":123}`, `{"type":"llm"}`,
		`{"type":"llm","backwards_request_id":null,"request":{}}`,
		`{"type":"llm","backwards_request_id":"r","request":null}`,
		`{"type":"llm","backwards_request_id":"r","request":[]}`,
		`{"type":"llm","backwards_request_id":"r","request":{}} {}`,
	} {
		t.Run(wire, func(t *testing.T) {
			_, err := invokeJSON(t, wire, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, true)
			require.Error(t, err)
		})
	}
	for _, tc := range []struct {
		name    string
		wire    string
		from    access_types.PluginAccessType
		allowed bool
		message string
	}{
		{"required fields", `{"type":"llm","backwards_request_id":"r","request":{}}`, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, true, "error validating struct"},
		{"invalid content", `{"type":"llm","backwards_request_id":"r","request":{"provider":"p","model":"m","mode":"chat","prompt_messages":[{"role":"assistant","content":123}]}}`, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, true, "content must be a string or an array"},
		{"permissions", `{"type":"llm","backwards_request_id":"r","request":{}}`, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, false, "permission denied"},
		{"model origin", `{"type":"llm","backwards_request_id":"r","request":{}}`, access_types.PLUGIN_ACCESS_TYPE_MODEL, true, "you can not invoke dify"},
		{"unsupported type", `{"type":"unsupported","backwards_request_id":"r","request":{}}`, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, true, "unsupported invoke type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events, err := invokeJSON(t, tc.wire, tc.from, tc.allowed)
			require.NoError(t, err)
			require.Len(t, events, 2)
			require.Equal(t, REQUEST_EVENT_ERROR, events[0].Event)
			require.Contains(t, events[0].Message, tc.message)
		})
	}
}
