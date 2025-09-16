# HTTP Requests

The codebase provides utilities for HTTP operations (`internal/utils/http_requests/`).

**Related Documentation:**
- [Stream Operations](stream.md) - Stream handling for HTTP responses
- [Generic Types](generics.md) - Type-safe HTTP parsing
- [Database Operations](database.md) - HTTP-based backwards invocation to Dify API

## Basic Requests

```go
// JSON request with typed response
resp, err := http_requests.RequestAndParse[ResponseType](
    client,
    "https://api.example.com/endpoint",
    "POST",
    http_requests.HttpHeader(map[string]string{
        "Authorization": "Bearer token",
    }),
    http_requests.HttpPayloadJson(map[string]any{
        "key": "value",
    }),
    http_requests.HttpWriteTimeout(30),
    http_requests.HttpReadTimeout(60),
)
```

## Stream Responses

Returns a [Stream](stream.md) for SSE or chunked responses:

```go
// For SSE or chunked responses
stream, err := http_requests.RequestAndParseStream[ChunkType](
    client,
    url,
    "GET",
    http_requests.HttpUsingLengthPrefixed(true),
)

for stream.Next() {
    chunk, err := stream.Read()
    // Process chunk
}
```

## HTTP Client Setup

```go
client := &http.Client{
    Transport: &http.Transport{
        Dial: (&net.Dialer{
            Timeout:   5 * time.Second,
            KeepAlive: 120 * time.Second,
        }).Dial,
        IdleConnTimeout: 120 * time.Second,
    },
}
```

## Request Options

### Headers and Parameters
```go
HttpHeader(map[string]string)     // Set headers
HttpParams(map[string]string)     // Query parameters
HttpDirectReferer()                // Set referer to request URL
```

### Payload Types
```go
HttpPayloadJson(any)               // JSON body
HttpPayloadText(string)            // Plain text
HttpPayloadMultipart(files, data)  // Multipart form
HttpPayloadReader(io.ReadCloser)   // Custom reader
```

### Timeouts
```go
HttpWriteTimeout(seconds)          // Write timeout
HttpReadTimeout(seconds)           // Read timeout
```

### Special Options
```go
HttpUsingLengthPrefixed(bool)      // For length-prefixed protocols
```

## Backwards Invocation Example

From `internal/core/dify_invocation/real/`:

```go
func Request[T any](i *RealBackwardsInvocation, method, path string, options ...HttpOptions) (*T, error) {
    options = append(options,
        HttpHeader(map[string]string{
            "X-Inner-Api-Key": i.difyInnerApiKey,
        }),
        HttpWriteTimeout(i.writeTimeout),
        HttpReadTimeout(i.readTimeout),
    )
    
    req, err := http_requests.RequestAndParse[BaseResponse[T]](
        i.client,
        i.difyPath(path),
        method,
        options...,
    )
    
    if err != nil {
        return nil, err
    }
    
    return req.Data, nil
}
```