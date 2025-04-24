package test_utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"golang.org/x/exp/rand"
)

// FakeOpenAIResponse represents the structure of an OpenAI chat completion response
type FakeOpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Delta        Delta   `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type Delta struct {
	Content string `json:"content"`
}

// StartFakeOpenAIServer starts a fake OpenAI server that streams responses
// Returns the port number and a cancel function to stop the server
func StartFakeOpenAIServer() (int, func()) {
	port, err := network.GetRandomPort()
	if err != nil {
		panic(fmt.Sprintf("Failed to get a random port: %v", err))
	}

	// Find an available port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(fmt.Sprintf("Failed to find an available port: %v", err))
	}

	listener.Close()

	// Create a new server
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Define the chat completions endpoint
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		// Set headers for streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Generate a random ID
		id := fmt.Sprintf("chatcmpl-%d", rand.Intn(1000000))

		// Create a list of random words to return
		words := []string{
			"hello", "world", "this", "is", "a", "fake", "openai", "server",
			"that", "streams", "responses", "for", "testing", "purposes",
			"only", "it", "will", "return", "words", "every", "hundred",
			"milliseconds", "until", "it", "reaches", "the", "limit",
			"of", "four", "hundred", "tokens", "please", "use", "this",
			"for", "benchmarking", "your", "plugin", "system", "thank",
			"you", "for", "using", "our", "service", "have", "a", "nice",
			"day", "goodbye", "see", "you", "soon", "take", "care",
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Stream 400 words, one every 100ms
		for i := 0; i < 100; i++ {
			word := words[i%len(words)]

			// Add space before words (except the first one)
			if i > 0 {
				word = " " + word
			}

			response := FakeOpenAIResponse{
				ID:      id,
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   "gpt-3.5-turbo",
				Choices: []struct {
					Index        int     `json:"index"`
					Delta        Delta   `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				}{
					{
						Index: i,
						Delta: Delta{
							Content: word,
						},
						FinishReason: nil,
					},
				},
			}

			// For the last message, set finish_reason to "stop"
			if i == 99 {
				response.Choices[0].FinishReason = &[]string{"stop"}[0]
			}

			data, _ := json.Marshal(response)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()

			// Sleep for 100ms
			time.Sleep(10 * time.Millisecond)

			// Check if the client has disconnected
			select {
			case <-r.Context().Done():
				return
			default:
			}
		}

		// Send the [DONE] message
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	})

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !strings.Contains(err.Error(), "server closed") {
			fmt.Printf("Fake OpenAI server error: %v\n", err)
		}
	}()

	// Return the port and a cancel function
	return int(port), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}
}
