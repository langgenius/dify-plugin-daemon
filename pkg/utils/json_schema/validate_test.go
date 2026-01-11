package jsonschema

import (
	"sync"
	"testing"
)

// This test simulates concurrent writes to a map while validating it.
// Using Validate (bytes-based) must not panic with "concurrent map iteration and map write".
func TestValidateConcurrentSafety(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"k": map[string]any{"type": "string"},
		},
	}
	doc := map[string]any{"k": "v"}

	// Mutate doc concurrently while we validate
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			select {
			case <-stop:
				return
			default:
				doc["k"] = "v" // write same key repeatedly
			}
		}
	}()

	// Run validation; must not panic
	if _, err := Validate(schema, doc); err != nil {
		// We don't assert pass/fail of the schema here; only that it doesn't panic and returns result/error sanely
	}
	close(stop)
	wg.Wait()
}
