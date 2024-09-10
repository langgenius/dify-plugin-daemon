package stream

import (
	"errors"
	"sync"
	"testing"
	"time"
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
