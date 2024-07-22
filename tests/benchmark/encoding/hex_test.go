package encoding

import (
	"encoding/hex"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/tests"
)

func BenchmarkHex(b *testing.B) {
	var data = []byte("hello world")
	var dst = make([]byte, len(data)*2)
	bytes := 0

	b.Run("Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			hex.Encode(dst, data)
			bytes += len(data)
		}
	})

	b.Log("Bytes encoded:", tests.ReadableBytes(bytes))

	encoded := make([]byte, len(data)*2)
	bytes = 0
	hex.Encode(encoded, data)
	b.Run("Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			hex.Decode(dst, encoded)
			bytes += len(encoded)
		}
	})

	b.Log("Bytes decoded:", tests.ReadableBytes(bytes))
}
