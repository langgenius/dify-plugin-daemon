package stream

import (
	"errors"
	"sync"
	"sync/atomic"
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

func TestStreamNextReturnsWhenClosed(t *testing.T) {
	response := NewStream[int](1)
	nextReturned := make(chan bool, 1)
	go func() {
		nextReturned <- response.Next()
	}()

	response.Close()

	select {
	case hasNext := <-nextReturned:
		assert.False(t, hasNext)
	case <-time.After(time.Second):
		t.Fatal("Next did not return after Close")
	}
}

func TestStreamCloseWakesAllWaitingReaders(t *testing.T) {
	response := NewStream[int](1)
	const readerCount = 3
	started := make(chan struct{}, readerCount)
	nextReturned := make(chan bool, readerCount)

	for range readerCount {
		go func() {
			started <- struct{}{}
			nextReturned <- response.Next()
		}()
	}
	for range readerCount {
		<-started
	}
	time.Sleep(10 * time.Millisecond)

	response.Close()

	for range readerCount {
		select {
		case hasNext := <-nextReturned:
			assert.False(t, hasNext)
		case <-time.After(time.Second):
			t.Fatal("not all waiting readers returned after Close")
		}
	}
}

func TestStreamErrorWakesNext(t *testing.T) {
	response := NewStream[int](1)
	nextReturned := make(chan bool, 1)
	go func() {
		nextReturned <- response.Next()
	}()

	expectedErr := errors.New("stream failed")
	response.WriteError(expectedErr)

	select {
	case hasNext := <-nextReturned:
		assert.True(t, hasNext)
	case <-time.After(time.Second):
		t.Fatal("Next did not return after WriteError")
	}
	_, err := response.Read()
	assert.ErrorIs(t, err, expectedErr)
}

func TestStreamConcurrentWriteAndClose(t *testing.T) {
	for range 1_000 {
		response := NewStream[int](1)
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_ = response.Write(1)
		}()
		go func() {
			defer wg.Done()
			<-start
			response.Close()
		}()
		close(start)
		wg.Wait()
	}
}

func TestStreamFilterCanCloseStream(t *testing.T) {
	response := NewStream[int](1)
	expectedErr := errors.New("rejected")
	response.Filter(func(int) error {
		response.Close()
		return expectedErr
	})
	assert.NoError(t, response.Write(1))

	readResult := make(chan error, 1)
	go func() {
		_, err := response.Read()
		readResult <- err
	}()

	select {
	case err := <-readResult:
		assert.ErrorIs(t, err, expectedErr)
	case <-time.After(time.Second):
		t.Fatal("filter deadlocked while closing the stream")
	}
	assert.True(t, response.IsClosed())
}

func TestStreamOnCloseAddedAfterCloseRuns(t *testing.T) {
	response := NewStream[int](1)
	response.Close()

	called := false
	response.OnClose(func() {
		called = true
	})

	assert.True(t, called)
}

func TestStreamNextWaitsForOnCloseCallbacks(t *testing.T) {
	response := NewStream[int](1)
	callbackStarted := make(chan struct{})
	releaseCallback := make(chan struct{})
	response.OnClose(func() {
		close(callbackStarted)
		<-releaseCallback
	})

	go response.Close()
	<-callbackStarted

	nextReturned := make(chan struct{})
	go func() {
		response.Next()
		close(nextReturned)
	}()

	select {
	case <-nextReturned:
		t.Fatal("Next returned before OnClose callback completed")
	case <-time.After(10 * time.Millisecond):
	}

	close(releaseCallback)
	select {
	case <-nextReturned:
	case <-time.After(time.Second):
		t.Fatal("Next did not return after OnClose callback completed")
	}
}

func TestStreamCloseWithErrorWaitsForOnCloseCallbacks(t *testing.T) {
	response := NewStream[int](1)
	expectedErr := errors.New("test error")
	callbackStarted := make(chan struct{})
	releaseCallback := make(chan struct{})
	response.OnClose(func() {
		close(callbackStarted)
		<-releaseCallback
	})

	go response.CloseWithError(expectedErr)
	<-callbackStarted

	nextReturned := make(chan bool, 1)
	go func() {
		nextReturned <- response.Next()
	}()

	select {
	case <-nextReturned:
		t.Fatal("Next returned before OnClose callback completed")
	case <-time.After(10 * time.Millisecond):
	}

	close(releaseCallback)
	select {
	case hasNext := <-nextReturned:
		assert.True(t, hasNext)
	case <-time.After(time.Second):
		t.Fatal("Next did not return after OnClose callback completed")
	}
	_, err := response.Read()
	assert.ErrorIs(t, err, expectedErr)
	assert.False(t, response.Next())
}

func TestStreamReadHidesCloseErrorUntilOnCloseCallbacksComplete(t *testing.T) {
	response := NewStream[int](1)
	expectedErr := errors.New("test error")
	callbackStarted := make(chan struct{})
	releaseCallback := make(chan struct{})
	response.OnClose(func() {
		close(callbackStarted)
		<-releaseCallback
	})

	go response.CloseWithError(expectedErr)
	<-callbackStarted

	_, err := response.Read()
	assert.ErrorIs(t, err, ErrEmpty)

	close(releaseCallback)
	assert.True(t, response.Next())
	_, err = response.Read()
	assert.ErrorIs(t, err, expectedErr)
}

func TestStreamConcurrentOnCloseAndClose(t *testing.T) {
	for range 1_000 {
		response := NewStream[int](1)
		start := make(chan struct{})
		var (
			calls atomic.Int32
			wg    sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			response.OnClose(func() {
				calls.Add(1)
			})
		}()
		go func() {
			defer wg.Done()
			<-start
			response.Close()
		}()
		close(start)
		wg.Wait()
		assert.EqualValues(t, 1, calls.Load())
	}
}
