package local_runtime

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/plugin_errors"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/assert"
)

// mockReadWriteCloser implements io.ReadWriteCloser for testing
type mockReadWriteCloser struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
	mutex    sync.Mutex
	readCond *sync.Cond
}

func newMockReadWriteCloser() *mockReadWriteCloser {
	m := &mockReadWriteCloser{
		readBuf:  bytes.NewBuffer(nil),
		writeBuf: bytes.NewBuffer(nil),
		closed:   false,
	}
	m.readCond = sync.NewCond(&m.mutex)
	return m
}

func (m *mockReadWriteCloser) Read(p []byte) (n int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Wait until there's data to read or the pipe is closed
	for m.readBuf.Len() == 0 && !m.closed {
		m.readCond.Wait()
	}

	if m.closed && m.readBuf.Len() == 0 {
		return 0, io.EOF
	}

	return m.readBuf.Read(p)
}

func (m *mockReadWriteCloser) Write(p []byte) (n int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return 0, io.ErrClosedPipe
	}

	return m.writeBuf.Write(p)
}

func (m *mockReadWriteCloser) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true
	m.readCond.Broadcast() // Wake up any waiting readers
	return nil
}

func (m *mockReadWriteCloser) WriteToRead(data []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return
	}

	m.readBuf.Write(data)
	m.readCond.Broadcast() // Signal that data is available
}

func (m *mockReadWriteCloser) GetWrittenData() []byte {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.writeBuf.Bytes()
}

func (m *mockReadWriteCloser) IsClosed() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.closed
}

// TestNewStdioHolder tests the creation of a new stdioHolder
func TestNewStdioHolder(t *testing.T) {
	stdin := newMockReadWriteCloser()
	stdout := newMockReadWriteCloser()
	stderr := newMockReadWriteCloser()

	holder := newStdioHolder("test-plugin", stdin, stdout, stderr, nil)

	assert.NotNil(t, holder)
	assert.Equal(t, "test-plugin", holder.pluginUniqueIdentifier)
	assert.Equal(t, stdin, holder.writer)
	assert.Equal(t, stdout, holder.reader)
	assert.Equal(t, stderr, holder.errReader)
	assert.NotNil(t, holder.l)
	assert.NotNil(t, holder.waitControllerChanLock)
	assert.NotNil(t, holder.waitingControllerChan)
	assert.False(t, holder.waitingControllerChanClosed)
}

// TestStdioHolderSetupAndRemoveListener tests setting up and removing event listeners
func TestStdioHolderSetupAndRemoveListener(t *testing.T) {
	holder := newStdioHolder("test-plugin", nil, nil, nil, nil)

	// Test setup listener
	holder.setupStdioEventListener("session1", func(data []byte) {
	})

	assert.NotNil(t, holder.listener)
	assert.Len(t, holder.listener, 1)
	assert.Contains(t, holder.listener, "session1")

	// Test remove listener
	holder.removeStdioHandlerListener("session1")
	assert.Len(t, holder.listener, 0)
}

// TestStdioHolderWrite tests writing to the stdio
func TestStdioHolderWrite(t *testing.T) {
	stdin := newMockReadWriteCloser()
	holder := newStdioHolder("test-plugin", stdin, nil, nil, nil)

	err := holder.write([]byte("test data"))
	assert.NoError(t, err)
	assert.Equal(t, "test data", string(stdin.GetWrittenData()))
}

// TestStdioHolderError tests the Error method
func TestStdioHolderError(t *testing.T) {
	holder := newStdioHolder("test-plugin", nil, nil, nil, nil)

	// No error initially
	assert.NoError(t, holder.Error())

	// Add an error message
	holder.WriteError("test error")
	holder.lastErrMessageUpdatedAt = time.Now()

	// Should return error
	err := holder.Error()
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())

	// Set last update time to more than 60 seconds ago
	holder.lastErrMessageUpdatedAt = time.Now().Add(-61 * time.Second)
	assert.NoError(t, holder.Error())
}

// TestStdioHolderStop tests stopping the stdio holder
func TestStdioHolderStop(t *testing.T) {
	stdin := newMockReadWriteCloser()
	stdout := newMockReadWriteCloser()
	stderr := newMockReadWriteCloser()

	holder := newStdioHolder("test-plugin", stdin, stdout, stderr, nil)

	holder.Stop()

	assert.True(t, stdin.IsClosed())
	assert.True(t, stdout.IsClosed())
	assert.True(t, stderr.IsClosed())
	assert.True(t, holder.waitingControllerChanClosed)

	// Test double stop (should not panic)
	holder.Stop()
}

// TestStdioHolderWriteError tests writing error messages
func TestStdioHolderWriteError(t *testing.T) {
	holder := newStdioHolder("test-plugin", nil, nil, nil, nil)

	// Test writing a simple error
	holder.WriteError("error1")
	assert.Equal(t, "error1", holder.errMessage)
	assert.WithinDuration(t, time.Now(), holder.lastErrMessageUpdatedAt, 1*time.Second)

	// Test appending error
	holder.WriteError("error2")
	assert.Equal(t, "error1error2", holder.errMessage)

	// Test truncation of error message when it exceeds MAX_ERR_MSG_LEN
	longError := strings.Repeat("a", 1025)
	holder.WriteError(longError)
	assert.Equal(t, 1024, len(holder.errMessage))
	assert.True(t, strings.HasSuffix(holder.errMessage, "a"))
}

// TestStdioHolderStartStderr tests the StartStderr method
func TestStdioHolderStartStderr(t *testing.T) {
	stderr := newMockReadWriteCloser()
	holder := newStdioHolder("test-plugin", nil, nil, stderr, nil)

	// Write some data to stderr
	stderr.WriteToRead([]byte("stderr message"))

	// Start a goroutine to read stderr
	done := make(chan bool)
	go func() {
		holder.StartStderr()
		done <- true
	}()

	// Close stderr to end the loop
	time.Sleep(100 * time.Millisecond)
	stderr.Close()

	// Wait for StartStderr to finish
	<-done

	// Check that the error message was captured
	assert.Contains(t, holder.errMessage, "stderr message")
}

// TestStdioHolderWait tests the Wait method
func TestStdioHolderWait(t *testing.T) {
	stdin := newMockReadWriteCloser()
	stdout := newMockReadWriteCloser()
	stderr := newMockReadWriteCloser()

	holder := newStdioHolder("test-plugin", stdin, stdout, stderr, nil)
	holder.lastActiveAt = time.Now()

	// Test normal wait with stop
	go func() {
		time.Sleep(100 * time.Millisecond)
		holder.Stop()
	}()

	err := holder.Wait()
	assert.NoError(t, err)

	stdin = newMockReadWriteCloser()
	stdout = newMockReadWriteCloser()
	stderr = newMockReadWriteCloser()

	holder = newStdioHolder("test-plugin", stdin, stdout, stderr, nil)
	holder.lastActiveAt = time.Now().Add(-(MAX_HEARTBEAT_INTERVAL + 1*time.Second)) // Inactive for more than 120 seconds

	// Test timeout due to inactivity
	err = holder.Wait()
	assert.Equal(t, plugin_errors.ErrPluginNotActive, err)

	stdin = newMockReadWriteCloser()
	stdout = newMockReadWriteCloser()
	stderr = newMockReadWriteCloser()

	// Test error when already closed
	holder = newStdioHolder("test-plugin", stdin, stdout, stderr, nil)
	holder.Stop()
	err = holder.Wait()
	assert.Error(t, err)
	assert.Equal(t, "you need to start the health check before waiting", err.Error())
}

// TestStdioHolderStartStdout tests the StartStdout method
func TestStdioHolderStartStdout(t *testing.T) {
	stdin := newMockReadWriteCloser()
	stdout := newMockReadWriteCloser()
	stderr := newMockReadWriteCloser()

	holder := newStdioHolder("test-plugin", stdin, stdout, stderr, nil)

	// Setup a listener
	receivedData := make(chan []byte, 1)
	holder.setupStdioEventListener("test-session", func(data []byte) {
		receivedData <- data
	})

	// Prepare JSON data that ParsePluginUniversalEvent can handle
	jsonData := []byte(`{"type":"event","session_id":"test-session","data":"test-data"}`)
	stdout.WriteToRead(jsonData)

	// Start a goroutine to read stdout
	go func() {
		holder.StartStdout(func() {})
	}()

	// Wait for data to be processed
	time.Sleep(100 * time.Millisecond)

	// Close stdout to end the loop
	stdout.Close()

	// Check that the data was processed
	assert.True(t, holder.started)
	assert.WithinDuration(t, time.Now(), holder.lastActiveAt, 1*time.Second)

	// Test with empty data
	stdin = newMockReadWriteCloser()
	stdout = newMockReadWriteCloser()
	stderr = newMockReadWriteCloser()

	holder = newStdioHolder("test-plugin", stdin, stdout, stderr, nil)
	stdout.WriteToRead([]byte{})

	go func() {
		holder.StartStdout(func() {})
	}()

	time.Sleep(100 * time.Millisecond)
	stdout.Close()
}

// TestStdioHolderWithRealData tests the stdioHolder with realistic data flow
func TestStdioHolderWithRealData(t *testing.T) {
	stdin := newMockReadWriteCloser()
	stdout := newMockReadWriteCloser()
	stderr := newMockReadWriteCloser()

	holder := newStdioHolder("test-plugin", stdin, stdout, stderr, nil)

	// Setup listeners
	dataReceived := make(chan bool, 1)
	holder.setupStdioEventListener("session1", func(data []byte) {
		assert.Equal(t, `"hello"`, string(data))
		dataReceived <- true
	})

	// Simulate plugin sending data
	heartbeatCalled := false
	go holder.StartStdout(func() {
		heartbeatCalled = true
	})

	heartbeatData := plugin_entities.PluginUniversalEvent{
		SessionId: "session1",
		Event:     plugin_entities.PLUGIN_EVENT_HEARTBEAT,
		Data:      []byte(`{}`),
	}

	heartbeatDataBytes := parser.MarshalJsonBytes(heartbeatData)

	// Send heartbeat event
	stdout.WriteToRead(heartbeatDataBytes)
	stdout.WriteToRead([]byte("\n"))
	time.Sleep(100 * time.Millisecond)
	assert.True(t, heartbeatCalled)

	sessionData := plugin_entities.PluginUniversalEvent{
		SessionId: "session1",
		Event:     plugin_entities.PLUGIN_EVENT_SESSION,
		Data:      []byte(`"hello"`),
	}
	sessionDataBytes := parser.MarshalJsonBytes(sessionData)

	// Send data event
	stdout.WriteToRead(sessionDataBytes)
	stdout.WriteToRead([]byte("\n"))

	// Wait for data to be processed
	select {
	case <-dataReceived:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for data to be received")
	}

	// Simulate error from plugin
	stderr.WriteToRead([]byte("Error from plugin"))
	go holder.StartStderr()
	time.Sleep(100 * time.Millisecond)

	assert.Contains(t, holder.errMessage, "Error from plugin")

	// Test cleanup
	holder.Stop()
	assert.True(t, stdin.IsClosed())
	assert.True(t, stdout.IsClosed())
	assert.True(t, stderr.IsClosed())
}

func TestMultipleTransactions(t *testing.T) {
	stdin := newMockReadWriteCloser()
	stdout := newMockReadWriteCloser()
	stderr := newMockReadWriteCloser()

	holder := newStdioHolder("test-plugin", stdin, stdout, stderr, nil)
	transactionMap := make(map[string]bool)
	transactionMapLock := sync.Mutex{}

	go holder.StartStdout(func() {})

	transaction := func() {
		id := uuid.New().String()
		transactionMapLock.Lock()
		transactionMap[id] = true
		transactionMapLock.Unlock()

		holder.setupStdioEventListener(id, func(data []byte) {
			assert.Equal(t, `"hello"`, string(data))
			transactionMapLock.Lock()
			delete(transactionMap, id)
			transactionMapLock.Unlock()

			holder.removeStdioHandlerListener(id)
		})

		payload := plugin_entities.PluginUniversalEvent{
			SessionId: id,
			Event:     plugin_entities.PLUGIN_EVENT_SESSION,
			Data:      []byte(`"hello"`),
		}

		payloadBytes := parser.MarshalJsonBytes(payload)
		stdout.WriteToRead(append(payloadBytes, '\n'))
	}

	// Run 100 transactions in parallel
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				transaction()
			}
		}()
	}

	time.Sleep(1 * time.Second)

	transactionMapLock.Lock()
	assert.Equal(t, 0, len(transactionMap))
	transactionMapLock.Unlock()
}
