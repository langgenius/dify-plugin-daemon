package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateMessage(t *testing.T) {
	message := "1234567890"
	message = truncateMessage(message)
	assert.Equal(t, "1234567890", message)

	message = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"

	message = ""
	for i := 0; i < 4096; i++ {
		message += "1"
	}
	message = truncateMessage(message)
	expected := ""
	for i := 0; i < 512; i++ {
		expected += "1"
	}
	expected += "..."
	for i := 0; i < 512; i++ {
		expected += "1"
	}
	assert.Equal(t, expected, message)
}
