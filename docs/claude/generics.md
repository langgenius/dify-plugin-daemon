# Generic Types Usage

The project uses Go generics (1.18+) extensively for type safety.

**Related Documentation:**
- [Database Operations](database.md) - Generic query builder patterns
- [Stream Operations](stream.md) - Type-safe streaming
- [Cache Operations](cache.md) - Generic cache operations
- [HTTP Requests](http-requests.md) - Generic request/response parsing

## Database Operations

```go
// Type-safe queries
plugin, _ := db.GetOne[models.Plugin](...)
plugins, _ := db.GetAll[models.Plugin](...)
count, _ := db.GetCount[models.Plugin](...)

// Generic constraints for comparisons
type genericComparableConstraint interface {
    int | int8 | int16 | int32 | int64 |
    uint | uint8 | uint16 | uint32 | uint64 |
    float32 | float64 | bool
}

func GreaterThan[T genericComparableConstraint](field string, value T)
```

## Stream Operations

```go
// Type-safe streaming
stream.NewStream[tool_entities.ToolResponseChunk](128)
stream.Stream[agent_entities.AgentStrategyResponseChunk]

// Consumer pattern
for stream.Next() {
    data, err := stream.Read()
    // data is correctly typed
}
```

## Plugin Invocation

```go
// Generic plugin invocation with request/response types
GenericInvokePlugin[RequestType, ResponseType](
    session,
    request,
    bufferSize,
)

// Example usage
response, err := GenericInvokePlugin[
    requests.RequestInvokeTool, 
    tool_entities.ToolResponseChunk,
](session, request, 128)
```

## Cache Operations

```go
// Type-safe cache operations
cache.Get[models.Plugin](key)
cache.GetMapField[node](mapKey, field)
cache.Subscribe[EventType](channel)
cache.AutoGetWithGetter[plugin_entities.PluginDeclaration](...)
```

## HTTP Request Parsing

```go
// Generic request/response handling
resp, err := http_requests.RequestAndParse[ResponseType](
    client,
    url,
    method,
    options...,
)

// Stream parsing
stream, err := http_requests.RequestAndParseStream[ChunkType](
    client,
    url,
    method,
)
```

## Service Layer Patterns

```go
// Base SSE handler with generics
baseSSEWithSession(
    func(session *Session) (*stream.Stream[T], error) {
        return plugin_daemon.InvokeTool(session, &r.Data)
    },
    accessType,
    accessAction,
    request,
    context,
    timeout,
)
```

## Benefits

1. **Compile-time type safety** - Catch type errors at build time
2. **Code reuse** - Single implementation for multiple types
3. **Better IDE support** - Auto-completion and type hints
4. **Reduced runtime errors** - No type assertions needed
5. **Clear APIs** - Types are explicit in function signatures