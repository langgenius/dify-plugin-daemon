package plugin_manager

import (
	"bytes"
	"errors"

	"github.com/panjf2000/gnet/v2"
)

type codec struct {
	buf bytes.Buffer
}

func (w *codec) Decode(c gnet.Conn) ([][]byte, error) {
	size := c.InboundBuffered()
	buf := make([]byte, size)
	read, err := c.Read(buf)

	if err != nil {
		return nil, err
	}

	if read < size {
		return nil, errors.New("read less than size")
	}

	// use \ as escape character, as for \ itself, it should be escaped as well
	var start int
	var result [][]byte = make([][]byte, 0)
	var current_line []byte = make([]byte, 0)
	for i := 0; i < size; i++ {
		if buf[i] == '\\' {
			// write to current line
			current_line = append(current_line, buf[start:i]...)
			start = i + 1
			i++
			continue
		}

		if buf[i] == '\n' {
			// write to current line
			current_line = append(current_line, buf[start:i]...)
			result = append(result, current_line)
			current_line = make([]byte, 0)
			start = i + 1
		}
	}

	// for the last line, write it to buffer
	if start < size {
		w.buf.Write(buf[start:size])
	}

	return result, nil
}
