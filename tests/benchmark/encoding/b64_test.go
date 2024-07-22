package encoding

import (
	"encoding/base64"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/tests"
)

func BenchmarkBase64(b *testing.B) {
	var data = []byte("hello world")
	bytes := 0
	var dst = make([]byte, base64.StdEncoding.EncodedLen(len(data)))

	b.Run("Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			base64.StdEncoding.Encode(dst, data)
			bytes += len(data)
		}
	})

	b.Log("Bytes encoded:", tests.ReadableBytes(bytes))

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	bytes = 0
	base64.StdEncoding.Encode(encoded, data)

	b.Run("Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			base64.StdEncoding.Decode(dst, encoded)
			bytes += len(encoded)
		}
	})

	b.Log("Bytes decoded:", tests.ReadableBytes(bytes))
}
