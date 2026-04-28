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

	response.Process(func(t int) {
		nums += 1
	})

	if nums != 10000 {
		t.Errorf("Expected 10000 messages, got %d", nums)
	}
}

func TestStreamBlockingWrite(t *testing.T) {
	response := NewStream[int](1)
	assert.NoError(t, response.Write(1))

	writerStarted := make(chan struct{})
	writerFinished := make(chan struct{})

	go func() {
		close(writerStarted)
		response.WriteBlocking(2)
		close(writerFinished)
	}()

	<-writerStarted

	select {
	case <-writerFinished:
		t.Fatal("WriteBlocking should block while the queue is full")
	case <-time.After(20 * time.Millisecond):
	}

	assert.True(t, response.Next())

	first, err := response.Read()
	assert.NoError(t, err)
	assert.Equal(t, 1, first)

	select {
	case <-writerFinished:
	case <-time.After(1 * time.Second):
		t.Fatal("WriteBlocking did not unblock after the queue had space")
	}

	assert.True(t, response.Next())

	second, err := response.Read()
	assert.NoError(t, err)
	assert.Equal(t, 2, second)
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

	// wait for the blocking write to happen
	time.Sleep(1 * time.Second)
	response.Close()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected the blocking write to be done")
	}
}
