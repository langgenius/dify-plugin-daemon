# Backwards LLM invocation timeouts

Agent strategy plugins can invoke an LLM through the daemon's reverse connection
to Dify's inner API. That connection uses framed responses even when the model
request has `stream: false`.

LLM and structured-output LLM reverse calls have three transport budgets. All
values below are milliseconds. Zero means **inherit**, not unlimited.

| Variable | Default | Meaning |
| --- | --- | --- |
| `DIFY_BACKWARDS_INVOCATION_LLM_TOTAL_TIMEOUT` | `0` | Entire reverse HTTP request, including headers and response body. Inherits `DIFY_BACKWARDS_INVOCATION_READ_TIMEOUT` (normally `240000`). |
| `DIFY_BACKWARDS_INVOCATION_LLM_FIRST_RESPONSE_TIMEOUT` | `0` | Time until the first decoded response frame. Inherits the effective LLM total budget. |
| `DIFY_BACKWARDS_INVOCATION_LLM_IDLE_TIMEOUT` | `0` | Maximum read inactivity after the first decoded frame. Incoming body bytes or decoded frames reset it. Inherits `DIFY_BACKWARDS_INVOCATION_READ_TIMEOUT`. |

These budgets do not change the timeout settings for tool, embedding, storage,
or other reverse calls. They also do not change the forward model invocation
protocol. The first-response budget is not a first-token SLA: an empty model
delta can already be a valid response frame. A blocking model invocation waits
under its first-response/total budgets, not the post-response idle budget.

For example, to allow a reverse LLM call to run for up to ten minutes without
changing the existing four-minute budget for other reverse calls:

```dotenv
DIFY_BACKWARDS_INVOCATION_READ_TIMEOUT=240000
DIFY_BACKWARDS_INVOCATION_LLM_TOTAL_TIMEOUT=600000
DIFY_BACKWARDS_INVOCATION_LLM_FIRST_RESPONSE_TIMEOUT=600000
DIFY_BACKWARDS_INVOCATION_LLM_IDLE_TIMEOUT=240000
```

Keep the enclosing plugin/Agent execution budget longer than the single LLM call
budget. Check `PLUGIN_MAX_EXECUTION_TIMEOUT`, the SDK's invocation timeout,
serverless transaction limits, and API/worker HTTP read timeouts separately;
their units and semantics differ. A caller cancellation or earlier parent
deadline always wins. Multiple LLM/tool calls share the enclosing execution
budget; increasing the per-call limit does not grant unlimited Agent execution.

Timeout messages include `phase=first_response`, `phase=read_idle`, or
`phase=total`, and `timeout_ms`. They use the existing reverse-call error envelope
so older SDKs can display the reason without a protocol upgrade. The daemon
cancels the reverse HTTP request and releases its timers/body when it completes
or the consumer closes it. Cancellation of the provider's actual computation
still depends on the downstream API/runtime/SDK honoring the disconnect.

No automatic retries are added. Retrying an entire Agent execution can repeat
tool side effects and model charges.

## Long-stream regression

The optional integration test below uses a local framed HTTP producer, not a
model provider. It runs a legacy control and the LLM policy concurrently, proving
that the deployment budgets permit an active stream past the old four-minute
cutoff. Export the daemon timeout settings from the example above first.

```sh
DIFY_TEST_BACKWARDS_LLM_STREAM_SECONDS=260 go test \
  ./internal/core/dify_invocation/calldify \
  -run '^TestBackwardsLLMLongStream$' -v -count=1 -timeout=8m
```
