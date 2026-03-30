package parser

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestLineBasedChunking(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		maxChunkSize int
		expected     []string
		expectError  bool
	}{
		{
			name:         "simple lines",
			input:        "line1\nline2\nline3",
			maxChunkSize: 100,
			expected:     []string{"line1", "line2", "line3"},
			expectError:  false,
		},
		{
			name:         "empty lines",
			input:        "line1\n\nline3",
			maxChunkSize: 100,
			expected:     []string{"line1", "", "line3"},
			expectError:  false,
		},
		{
			name:         "single line",
			input:        "single line",
			maxChunkSize: 100,
			expected:     []string{"single line"},
			expectError:  false,
		},
		{
			name:         "line too long",
			input:        "this is a very long line that exceeds the maximum chunk size",
			maxChunkSize: 10,
			expected:     nil,
			expectError:  true,
		},
		{
			name:         "empty input",
			input:        "",
			maxChunkSize: 100,
			expected:     []string{},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))
			var result []string

			err := LineBasedChunking(reader, tt.maxChunkSize, func(data []byte) error {
				result = append(result, string(data))
				return nil
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d lines, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("line %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestLengthPrefixedChunking(t *testing.T) {
	tests := []struct {
		name        string
		data        [][]byte
		magicNumber byte
		expectError bool
	}{
		{
			name:        "valid single chunk",
			data:        [][]byte{[]byte("hello world")},
			magicNumber: 0x0f,
			expectError: false,
		},
		{
			name:        "valid multiple chunks",
			data:        [][]byte{[]byte("chunk1"), []byte("chunk2"), []byte("chunk3")},
			magicNumber: 0x0f,
			expectError: false,
		},
		{
			name:        "empty chunk",
			data:        [][]byte{[]byte("")},
			magicNumber: 0x0f,
			expectError: false,
		},
		{
			name:        "large chunk",
			data:        [][]byte{bytes.Repeat([]byte("a"), 1000)},
			magicNumber: 0x0f,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test data with proper format
			var buf bytes.Buffer
			for _, chunk := range tt.data {
				// Write magic number
				buf.WriteByte(tt.magicNumber)
				// Write reserved byte
				buf.WriteByte(0x00)
				// Write header length (0x000a in little endian)
				buf.Write([]byte{0x0a, 0x00})

				// Create header (10 bytes total)
				header := make([]byte, 10)
				// First 4 bytes are already used for data length placeholder
				// Write data length in bytes 4-7 (little endian)
				dataLen := uint32(len(chunk))
				binary.LittleEndian.PutUint32(header[:4], dataLen)
				// Remaining 6 bytes are reserved (already zero)

				buf.Write(header)
				buf.Write(chunk)
			}

			reader := bytes.NewReader(buf.Bytes())
			var result [][]byte

			err := LengthPrefixedChunking(reader, tt.magicNumber, 1024*1024, func(data []byte) error {
				// Make a copy of the data since it might be reused
				chunk := make([]byte, len(data))
				copy(chunk, data)
				result = append(result, chunk)
				return nil
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.data) {
				t.Errorf("expected %d chunks, got %d", len(tt.data), len(result))
				return
			}

			for i, expected := range tt.data {
				if !bytes.Equal(result[i], expected) {
					t.Errorf("chunk %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestLengthPrefixedChunking_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		magicNumber byte
		maxSize     uint32
		expectError string
	}{
		{
			name:        "wrong magic number",
			data:        []byte{0x10, 0x00, 0x0a, 0x00}, // wrong magic number
			magicNumber: 0x0f,
			maxSize:     1024,
			expectError: "magic number mismatch",
		},
		{
			name:        "wrong header length",
			data:        []byte{0x0f, 0x00, 0x0b, 0x00}, // wrong header length
			magicNumber: 0x0f,
			maxSize:     1024,
			expectError: "header length mismatch",
		},
		{
			name:        "incomplete header",
			data:        []byte{0x0f, 0x00, 0x0a, 0x00, 0x01, 0x00}, // incomplete header
			magicNumber: 0x0f,
			maxSize:     1024,
			expectError: "failed to read header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)

			err := LengthPrefixedChunking(reader, tt.magicNumber, tt.maxSize, func(data []byte) error {
				return nil
			})

			if err == nil {
				t.Errorf("expected error containing %q but got none", tt.expectError)
				return
			}

			if err.Error() == "" {
				t.Errorf("expected error containing %q but got empty error", tt.expectError)
			}
		})
	}
}

func TestLengthPrefixedChunking_DataTooLarge(t *testing.T) {
	var buf bytes.Buffer

	// Create a chunk that exceeds maxChunkSize
	largeDataSize := uint32(100)
	maxChunkSize := uint32(50)

	// Write magic number and reserved
	buf.WriteByte(0x0f)
	buf.WriteByte(0x00)
	// Write header length
	buf.Write([]byte{0x0a, 0x00})

	// Create header with large data size
	header := make([]byte, 10)
	binary.LittleEndian.PutUint32(header[:4], largeDataSize)

	buf.Write(header)

	reader := bytes.NewReader(buf.Bytes())

	err := LengthPrefixedChunking(reader, 0x0f, maxChunkSize, func(data []byte) error {
		return nil
	})

	if err == nil {
		t.Error("expected error for data too large but got none")
		return
	}

	if err.Error() == "" {
		t.Error("expected error message but got empty")
	}
}
