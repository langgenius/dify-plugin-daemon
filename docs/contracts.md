# HTTP contracts

[api/openapi.yaml](../api/openapi.yaml) owns the HTTP request and response shapes
for these tool management operations:

| Operation | Method | Path |
| --- | --- | --- |
| `ListTools` | `GET` | `/plugin/{tenant_id}/management/tools` |
| `GetTool` | `GET` | `/plugin/{tenant_id}/management/tool` |

This first slice includes installed tool provider declarations, credentials,
OAuth configuration, and tool parameters returned by these operations. Other
management operations, streaming invocation protocols, and generated Python or
TypeScript consumers remain separate follow-up work.

## Ownership and enforcement

The contract flows through these boundaries:

```text
api/openapi.yaml
  -> internal/server/contracts/api.gen.go
  -> tool management strict handlers
  -> explicit service/domain-to-HTTP mapping
```

[oapi-codegen.yaml](../internal/server/contracts/oapi-codegen.yaml) enables Go
models, Gin route registration, strict server interfaces, and an embedded copy
of the specification. The tool management routes use Gin request validation
against that specification before calling their strict handlers. The existing
daemon API key middleware remains responsible for authentication.

Query constraints are checked against the values parsed by the generated
binding code. This keeps decimal integer binding and validation consistent:
`page=08` means 8, while `page_size=0400` exceeds the maximum of 256. The
validator reads the constraints from the OpenAPI parameters, without copying
pagination limits into Go code.

Strict handler interfaces require generated request and response types. They do
not by themselves prove that serialized responses satisfy every schema
constraint. The `TestToolManagement` tests exercise the registered Gin handlers
and validate their actual HTTP status, headers, and response bodies against the
specification.

[Plugin domain entities](../pkg/entities/plugin_entities) retain ownership of
plugin YAML loading, validation, defaults, and runtime behavior. Generated HTTP
models do not replace those entities or their YAML tags. The controller boundary
maps service/domain values into generated response models explicitly. A field
that belongs in both the plugin manifest and the HTTP response therefore needs
an intentional mapping and a serialization test.

Generation checks keep the OpenAPI document and generated Go code in sync.
Domain changes still require reviewing the HTTP mapping: adding an optional
manifest field or a domain enum value does not necessarily fail compilation.

## Compatibility

The specification describes the existing wire behavior of these endpoints:

- Successful responses contain `code: 0`, `message: "success"`, and typed `data`.
- Daemon operation failures still return HTTP 200 with a negative `code`,
  `data: null`, and a `message` containing JSON-encoded error details as a string.
  Consumers must inspect the envelope; HTTP 200 alone does not mean success.
- Request validation failures use HTTP 400. Authentication failures use the
  existing HTTP 401 daemon error response.
- Scalar query parameters must occur once. Repeated values now return HTTP 400;
  the previous Gin binding silently selected the first value.
- Required, optional, and nullable fields are separate contract decisions.
  Preserve existing `null`, omitted fields, and empty arrays when mapping and
  serializing responses.
- Plugin-defined JSON values and schemas remain open where the specification
  permits them. Do not narrow arbitrary defaults or JSON Schema content to
  strings merely to simplify generated types.
- `tenant_id` remains a non-empty string. The contract does not introduce a UUID
  requirement or change the existing provider and plugin identifier formats.

Response tests compare representative legacy and generated serialization as
well as checking schema validity. This protects behavior that a type check alone
cannot establish, including nullability, open JSON values, and error envelopes.

## Updating the contract

Use Go 1.26. The repository pins `oapi-codegen` v2.8.0 as a Go tool dependency in
[go.mod](../go.mod); generation uses `go tool oapi-codegen` rather than a globally
installed executable. See the [Go tool dependency documentation][go-tools] for
how tool versions are recorded.

The specification uses the OpenAPI 3.1 features covered by the contract tests.
Go type overrides are limited to JSON values and nullable shared objects that
the generator cannot represent directly; schema validation still checks their
wire constraints. New schema features need both accepted and rejected payload
tests against the pinned generator and validator.

From the repository root:

```sh
go generate ./internal/server/contracts
go test ./internal/server ./internal/server/controllers -run '^Test(ToolManagement|ServerHostBinding)'
```

For a contract change:

1. Update `api/openapi.yaml` to describe the intended HTTP behavior.
2. Generate `internal/server/contracts/api.gen.go`; do not edit generated code
   directly.
3. Update the strict handlers and explicit mappings where required.
4. Add or update HTTP response and request validation cases, including relevant
   compatibility cases.
5. Commit the specification, generated output, implementation, and tests
   together.

[The contract workflow](../.github/workflows/contracts.yml) regenerates the code,
checks for modified, deleted, and untracked files in the specification and
generated contract paths, then runs the focused controller, authentication, and
server startup tests. It does not start PostgreSQL, Redis, or a plugin runtime.
The response tests supply service results at the handler boundary while
exercising the real HTTP registration, validation, mapping, and serialization.

The generation check can also be run locally after committing the expected
changes:

```sh
go generate ./internal/server/contracts
if [ -n "$(git status --porcelain --untracked-files=all -- api/openapi.yaml internal/server/contracts)" ]; then
  git status --short --untracked-files=all -- api/openapi.yaml internal/server/contracts
  git diff -- api/openapi.yaml internal/server/contracts
  exit 1
fi
```

[go-tools]: https://go.dev/doc/modules/managing-dependencies#tools
