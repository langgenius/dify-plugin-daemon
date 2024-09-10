package stream

import (
	"sync"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func BenchmarkStreamResponse(b *testing.B) {
	b.Run("Read", func(b *testing.B) {
		wg_started := sync.WaitGroup{}
		wg_started.Add(8)
		resp := stream.NewStream[int](1024)

		for i := 0; i < 8; i++ {
			go func() {
				wg_started.Done()
				for !resp.IsClosed() {
					resp.Write(1)
				}
			}()
		}

		// wait for the first element to be written
		resp.Next()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp.Next()
			resp.Read()
		}
		defer resp.Close()
	})
}
