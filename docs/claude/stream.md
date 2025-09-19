# Stream Operations

Thread-safe, generic streaming for async producer-consumer patterns (`internal/utils/stream/`).

**Related Documentation:**
- [HTTP Requests](http-requests.md) - HTTP streams using `RequestAndParseStream`
- [Generic Types](generics.md) - Stream type safety patterns
- [Cache Operations](cache.md) - Pub/sub streaming patterns

## Core Concepts

The `Stream[T]` type provides buffered, type-safe communication between goroutines with backpressure control.

## Basic Usage

### Creation
```go
// Create with buffer size (max items before blocking)
s := stream.NewStream[T](128)
defer s.Close()
```

### Producer Operations
```go
// Non-blocking write (errors if full)
err := s.Write(data)

// Blocking write (waits for space)
s.WriteBlocking(data)

// Propagate errors
s.WriteError(err)

// Signal completion
s.Close()
```

### Consumer Operations
```go
// Standard iteration pattern
for s.Next() {         // Blocks waiting for data/close
    data, err := s.Read()
    if err != nil {
        if err == stream.ErrEmpty {
            continue
        }
        // Handle actual error
        break
    }
    // Process data
}
```

## Advanced Features

### Lifecycle Hooks
```go
// Cleanup on close
s.OnClose(func() {
    // Release resources
})

// Pre-close operations
s.BeforeClose(func() {
    // Finalize state
})
```

### Data Filtering
```go
// Add validation/transformation
s.Filter(func(data T) error {
    if !isValid(data) {
        return errors.New("invalid data")
    }
    return nil
})
```

### Async Processing
```go
// Process all items
err := s.Async(func(data T) {
    processItem(data)
})
```

### Status Methods
```go
s.IsClosed()    // Check if closed
s.Size()        // Current buffer size
```

## Common Patterns

### Plugin Response Streaming
```go
func handlePluginResponse(pluginOutput <-chan Chunk) *stream.Stream[Chunk] {
    response := stream.NewStream[Chunk](128)
    
    go func() {
        defer response.Close()
        for chunk := range pluginOutput {
            if err := response.Write(chunk); err != nil {
                response.WriteError(err)
                return
            }
        }
    }()
    
    return response
}
```

### SSE Response Handler
```go
func handleSSE(ctx *gin.Context, dataStream *stream.Stream[[]byte]) {
    ctx.Header("Content-Type", "text/event-stream")
    
    for dataStream.Next() {
        data, err := dataStream.Read()
        if err != nil {
            ctx.SSEvent("error", err.Error())
            return
        }
        ctx.SSEvent("data", string(data))
        ctx.Writer.Flush()
    }
}
```

### Error Propagation
```go
func processWithValidation(input *stream.Stream[Data]) *stream.Stream[Result] {
    output := stream.NewStream[Result](64)
    
    go func() {
        defer output.Close()
        for input.Next() {
            data, err := input.Read()
            if err != nil {
                output.WriteError(err)
                return
            }
            
            result, err := validate(data)
            if err != nil {
                output.WriteError(err)
                return
            }
            
            output.Write(result)
        }
    }()
    
    return output
}
```

### File Chunk Assembly
```go
files := make(map[string]*bytes.Buffer)

for response.Next() {
    chunk, _ := response.Read()
    
    if chunk.Type == "blob_chunk" {
        id := chunk.ID
        if chunk.End {
            // Complete file
            completeFile := files[id].Bytes()
            processFile(completeFile)
            delete(files, id)
        } else {
            // Accumulate chunks
            if files[id] == nil {
                files[id] = bytes.NewBuffer(nil)
            }
            files[id].Write(chunk.Data)
        }
    }
}
```

## Implementation Details

- Uses `deque` for efficient FIFO operations
- Thread-safe with mutex protection
- Signal channel for blocking consumers
- Atomic operations for close state
- Condition variable for blocking writers