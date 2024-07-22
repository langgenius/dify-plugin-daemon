package stdio

import (
	"os"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/tests"
)

func BenchmarkStdioBandWidth(b *testing.B) {
	// open /dev/zero for reading
	buf := make([]byte, 1024)
	zero := os.NewFile(0, "/dev/zero")
	bytes := 0

	b.Run("Read", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			zero.Read(buf)
			bytes += len(buf)
		}
	})

	b.Log("Bytes read:", tests.ReadableBytes(bytes))
}
