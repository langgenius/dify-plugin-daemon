package http_requests

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockReader struct {
	chunks []string
	index  int
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.index >= len(m.chunks) {
		return 0, io.EOF
	}
	n = copy(p, m.chunks[m.index])
	m.index++
	if m.index == len(m.chunks) {
		return n, io.EOF
	}
	return n, nil
}

func TestParseJsonBody(t *testing.T) {
	t.Run("multiple chunks with newlines", func(t *testing.T) {
		chunks := []string{
			`{"name": "John",`,
			"\n",
			`"age": 30}`,
			"\n",
		}
		reader := &mockReader{chunks: chunks}
		resp := &http.Response{Body: io.NopCloser(reader)}

		var result map[string]interface{}
		err := parseJsonBody(resp, &result)
		assert.Nil(t, err)

		assert.Equal(t, "John", result["name"])
		assert.Equal(t, 30, int(result["age"].(float64)))
	})

	t.Run("chunks without newlines", func(t *testing.T) {
		chunks := []string{
			`{"name": "Alice",`,
			`"age": 25}`,
		}
		reader := &mockReader{chunks: chunks}
		resp := &http.Response{Body: io.NopCloser(reader)}

		var result map[string]interface{}
		err := parseJsonBody(resp, &result)
		assert.Nil(t, err)
		assert.Equal(t, "Alice", result["name"])
		assert.Equal(t, 25, int(result["age"].(float64)))
	})

	t.Run("chunks with mixed newlines", func(t *testing.T) {
		chunks := []string{
			`{"name": "Bob",`,
			"\n",
			`"age": 35`,
			`,"city": "New York"}`,
		}
		reader := &mockReader{chunks: chunks}
		resp := &http.Response{Body: io.NopCloser(reader)}

		var result map[string]interface{}
		err := parseJsonBody(resp, &result)
		assert.Nil(t, err)
		assert.Equal(t, "Bob", result["name"])
		assert.Equal(t, 35, int(result["age"].(float64)))
		assert.Equal(t, "New York", result["city"])
	})

	t.Run("last chunk without newline", func(t *testing.T) {
		chunks := []string{
			`{"name": "Eve",`,
			"\n",
			`"age": 28}`,
		}
		reader := &mockReader{chunks: chunks}
		resp := &http.Response{Body: io.NopCloser(reader)}

		var result map[string]interface{}
		err := parseJsonBody(resp, &result)
		assert.Nil(t, err)
		assert.Equal(t, "Eve", result["name"])
		assert.Equal(t, 28, int(result["age"].(float64)))
	})

	t.Run("empty chunks", func(t *testing.T) {
		chunks := []string{
			"",
			"\n",
			"",
			`{"name": "Charlie"}`,
			"\n",
		}
		reader := &mockReader{chunks: chunks}
		resp := &http.Response{Body: io.NopCloser(reader)}

		var result map[string]interface{}
		err := parseJsonBody(resp, &result)
		assert.Nil(t, err)
		assert.Equal(t, "Charlie", result["name"])
	})

	t.Run("invalid JSON", func(t *testing.T) {
		chunks := []string{
			`{"name": "Invalid`,
			"\n",
			`"age": }`,
		}
		reader := &mockReader{chunks: chunks}
		resp := &http.Response{Body: io.NopCloser(reader)}

		var result map[string]interface{}
		err := parseJsonBody(resp, &result)
		assert.NotNil(t, err)
	})

	t.Run("large JSON split across multiple chunks", func(t *testing.T) {
		largeJSON := strings.Repeat(`{"key": "value"},`, 1000) // Create a large JSON array
		largeJSON = "[" + largeJSON[:len(largeJSON)-1] + "]"   // Remove last comma and wrap in array brackets

		chunkSize := 100
		chunks := make([]string, 0, len(largeJSON)/chunkSize+1)
		for i := 0; i < len(largeJSON); i += chunkSize {
			end := i + chunkSize
			if end > len(largeJSON) {
				end = len(largeJSON)
			}
			chunks = append(chunks, largeJSON[i:end])
		}

		reader := &mockReader{chunks: chunks}
		resp := &http.Response{Body: io.NopCloser(reader)}

		var result []map[string]string
		err := parseJsonBody(resp, &result)
		assert.Nil(t, err)
		assert.Equal(t, 1000, len(result))
	})
}
