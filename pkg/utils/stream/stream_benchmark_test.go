package stream

import (
	"testing"
)

func BenchmarkStream(b *testing.B) {
	stream := NewStream[string](1)
	go func() {
		for stream.Next() {
			stream.Read()
		}
	}()

	for i := 0; i < b.N; i++ {
		stream.Write("Hello, World!")
	}

	stream.Close()
}
