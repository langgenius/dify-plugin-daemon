package encoding

import (
	"encoding/ascii85"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/tests"
)

func BenchmarkAscii85(b *testing.B) {
	var data = []byte("hello world")
	var dst = make([]byte, ascii85.MaxEncodedLen(len(data)))
	bytes := 0

	b.Run("Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ascii85.Encode(dst, data)
			bytes += len(data)
		}
	})

	b.Log("Bytes encoded:", tests.ReadableBytes(bytes))

	encoded := make([]byte, ascii85.MaxEncodedLen(len(data)))
	bytes = 0
	ascii85.Encode(encoded, data)
	b.Run("Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ascii85.Decode(dst, encoded, true)
			bytes += len(encoded)
		}
	})

	b.Log("Bytes decoded:", tests.ReadableBytes(bytes))
}
