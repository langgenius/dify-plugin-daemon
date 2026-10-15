# Model dispatch redirect v1

An authenticated Dify caller can execute a model operation on another installed
model plugin without running the original plugin. The caller decides when to
redirect, maps model IDs and parameters, and supplies the target credentials.
The daemon does not inspect subscription, credit, or billing state.

## Discovery and request

Both endpoints require the existing `X-Api-Key` server authentication:

- `GET /plugin/{tenant_id}/dispatch/redirect/v1/capabilities` returns the usual
  success envelope with `data.protocols = ["model-redirect/v1"]` and an explicit
  `data.operations` array.
- `POST /plugin/{tenant_id}/dispatch/redirect/v1/{operation}` accepts the request
  below. `operation` is an existing relative model action, such as `llm/invoke`.

```json
{
  "user_id": "user-id",
  "app_id": "app-id",
  "data": {
    "provider": "source-provider",
    "model": "source-model",
    "model_type": "llm",
    "credentials": {},
    "prompt_messages": [{"role": "user", "content": "Hello"}],
    "stream": true
  },
  "redirect": {
    "version": 1,
    "route_id": "33333333-3333-4333-8333-333333333333",
    "target": {
      "plugin_id": "example/target",
      "provider": "target-provider",
      "model_type": "llm",
      "model": "target-model",
      "credentials": {"api_key": "EXAMPLE_ONLY"}
    }
  }
}
```

`app_id` and `target.expected_unique_identifier` are optional strings. If supplied,
the expected identifier must exactly equal the tenant's installed identifier.
It does not select a package, runtime, or node. The URI tenant is authoritative.
The source plugin need not be installed, and a source `X-Plugin-ID` header is not
required. The daemon replaces it with the target installation before routing.

The envelope, `redirect`, and `target` reject unknown fields. Source credentials
must be an empty object; target credentials must be a nonempty object. Target
tenant IDs, runtime URLs, and recursive redirects are not accepted. Endpoint
fields inside the target's credentials remain subject to that plugin's existing
rules. `data` also passes the existing action-specific typed validation, and
source/target model types must match the operation. The source authentication
type is not inherited by the target.

## Supported actions

V1 advertises only these implemented actions:

- `model/schema`, `model/validate_model_credentials`
- `llm/invoke`, `llm/num_tokens`
- `text_embedding/invoke`, `text_embedding/num_tokens`
- `rerank/invoke`
- `tts/invoke`, `tts/model/voices`
- `speech2text/invoke`, `moderation/invoke`
- `model/polling/start`, `model/polling/check`

Model types are `llm`, `text-embedding`, `rerank`, `tts`, `speech2text`, and
`moderation`. Tool, agent, provider-credential validation, and multimodal actions
are not advertised by this version. A TTS data tenant, when supplied, must match
the URI tenant; otherwise the daemon fills it from the URI.

## Execution and failures

The order is authentication, envelope and typed data validation, same-tenant
installation lookup, optional version comparison, target cluster routing, then
the existing typed model service. Credentials are never attached to plugin
Context, Gin context, session metadata, logs, or trace attributes. The original
request body is preserved for a remote node to repeat installation validation.
Only the target model data is delivered to its plugin.

Streaming chunks, tool deltas, usage and target errors use the existing model
protocol. Polling retains its existing non-streaming response. Cross-node SSE
responses flush incrementally and propagate caller cancellation. The daemon does
not retry an unknown invocation result or fall back to the source plugin. Each
accepted HTTP request dispatches once; `route_id` is correlation metadata, not
cross-request deduplication.

Pre-dispatch errors have a string `code` and a generic `message`:

| HTTP | Code | Meaning |
| --- | --- | --- |
| 404 | `redirect_protocol_unsupported` | Unsupported version or action |
| 404 | `redirect_target_not_installed` | No matching target installation |
| 409 | `redirect_target_version_mismatch` | Installed version changed |
| 403 | `redirect_invalid_tenant` | Tenant isolation check failed |
| 400 | `redirect_invalid_request` | Invalid envelope or action data |

Installation lookup infrastructure errors return HTTP 500 with a sanitized
message. Once streaming begins, target errors use the existing SSE error frame.

Deploy support across the daemon fleet before enabling callers. Old daemon
versions return 404 on the new path; callers must not fall back to old dispatch
URLs. Capability discovery is intended for deployment/preparation validation,
not a random-pod probe before each model request. Existing dispatch URLs continue
to use their original request format and behavior.
