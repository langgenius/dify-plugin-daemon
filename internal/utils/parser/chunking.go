package parser

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func LineBasedChunking(reader io.Reader, maxChunkSize int, processor func([]byte) error) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024), maxChunkSize)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > maxChunkSize {
			return fmt.Errorf("line is too long: %d", len(line))
		}

		if err := processor(line); err != nil {
			return err
		}
	}
	return nil
}

// We uses following format:
// All data is stored in little endian format
//
//	| Field         | Size     | Description                     |
//	|---------------|----------|---------------------------------|
//	| Magic Number  | 1 byte   | Magic number identifier         |
//	| Reserved      | 1 byte   | Reserved field                  |
//	| Header Length | 2 bytes  | Header length (usually 0xa)    |
//	| Data Length   | 4 bytes  | Length of the data              |
//	| Reserved      | 6 bytes  | Reserved fields                 |
//	| Data          | Variable | Actual data content             |
//
//	| Reserved Fields | Header   | Data     |
//	|-----------------|----------|----------|
//	| 4 bytes total   | Variable | Variable |
//
// NOTE: this function is not thread safe
func LengthPrefixedChunking(
	reader io.Reader,
	magicNumber byte,
	maxChunkSize uint32,
	processor func([]byte) error,
) error {
	// read until EOF
	buf := bytes.NewBuffer(nil)

	for {
		// read length
		length := make([]byte, 4)
		_, err := io.ReadFull(reader, length)
		if err != nil {
			if err == io.EOF {
				return nil // Normal EOF, processing complete
			}
			return errors.Join(err, fmt.Errorf("failed to read system header"))
		}

		// check magic number
		if length[0] != magicNumber {
			return fmt.Errorf("magic number mismatch: %d", length[0])
		}

		// read header length
		headerLength := binary.LittleEndian.Uint16(length[2:4])
		if headerLength != 0xa {
			return fmt.Errorf("header length mismatch: %d", headerLength)
		}

		// read header
		header := make([]byte, headerLength)
		_, err = io.ReadFull(reader, header)
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to read header"))
		}

		// decoding data length
		dataLength := binary.LittleEndian.Uint32(header[:4])
		if dataLength > maxChunkSize {
			return fmt.Errorf("data length is too long: %d", dataLength)
		}

		// Reset buffer for new data
		buf.Reset()

		// Read data into buffer
		// io.CopyN will not return io.EOF if dataLength equals to actual data length
		_, err = io.CopyN(buf, reader, int64(dataLength))
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to read data"))
		}

		// Process the data
		err = processor(buf.Bytes())
		if err != nil {
			return errors.Join(err, fmt.Errorf("failed to process data"))
		}
	}
}
