package stream

import (
	"errors"
	"testing"
	"time"
)

func TestStreamGenerator(t *testing.T) {
	response := NewStreamResponse[int](512)

	go func() {
		for i := 0; i < 10000; i++ {
			response.Write(i)
			time.Sleep(time.Microsecond)
		}
		response.Close()
	}()

	msg := 0

	for response.Next() {
		_, err := response.Read()
		if err != nil {
			t.Error(err)
		}
		msg += 1
	}

	if msg != 10000 {
		t.Errorf("Expected 10000 messages, got %d", msg)
	}
}

func TestStreamGeneratorErrorMessage(t *testing.T) {
	response := NewStreamResponse[int](512)

	go func() {
		for i := 0; i < 10000; i++ {
			response.Write(i)
			time.Sleep(time.Microsecond)
		}
		response.WriteError(errors.New("test error"))
		response.Close()
	}()

	for response.Next() {
		_, err := response.Read()
		if err != nil {
			if err.Error() != "test error" {
				t.Error(err)
			}
		}
	}
}
