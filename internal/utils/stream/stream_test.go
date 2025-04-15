package stream

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStreamGenerator(t *testing.T) {
	response := NewStream[int](512)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for i := 0; i < 10000; i++ {
			response.Write(i)
			time.Sleep(time.Microsecond)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			response.Write(i)
			time.Sleep(time.Microsecond)
		}
		wg.Done()
	}()

	go func() {
		wg.Wait()
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

	if msg != 20000 {
		t.Errorf("Expected 10000 messages, got %d", msg)
	}
}

func TestStreamGeneratorErrorMessage(t *testing.T) {
	response := NewStream[int](512)

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

func TestStreamGeneratorWrapper(t *testing.T) {
	response := NewStream[int](512)
	nums := 0

	go func() {
		for i := 0; i < 10000; i++ {
			response.Write(i)
			time.Sleep(time.Microsecond)
		}
		response.Close()
	}()

	response.Async(func(t int) {
		nums += 1
	})

	if nums != 10000 {
		t.Errorf("Expected 10000 messages, got %d", nums)
	}
}

func TestStreamBlockingWrite(t *testing.T) {
	response := NewStream[int](1)
	response.Write(1)

	const numWrites = 1000000

	go func() {
		for i := 0; i < numWrites; i++ {
			response.WriteBlocking(1)
			time.Sleep(time.Microsecond)
		}
		response.Close()
	}()

	received := 0
	done := make(chan bool)
	go func() {
		defer func() {
			close(done)
		}()
		// wait for the blocking write to happen
		time.Sleep(1 * time.Second)
		for response.Next() {
			_, err := response.Read()
			if err != nil {
				t.Error(err)
			}
			received += 1
		}
	}()

	<-done
	assert.Equal(t, received, numWrites+1)
}

// WriteBlocking should return directly if the stream is closed
func TestStreamCloseBlockingWrite(t *testing.T) {
	response := NewStream[int](1)
	response.Write(1)

	done := make(chan bool)

	go func() {
		response.WriteBlocking(1)
		close(done)
	}()

	response.Close()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected the blocking write to be done")
	}
}
