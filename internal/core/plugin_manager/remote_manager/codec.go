package remote_manager

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

	return w.getLines(buf), nil
}

func (w *codec) getLines(data []byte) [][]byte {
	// write to buffer
	w.buf.Write(data)

	// read line by line, split by \n, remaining data will be kept in buffer
	buf := make([]byte, w.buf.Len())
	w.buf.Read(buf)
	w.buf.Reset()

	lines := bytes.Split(buf, []byte("\n"))

	// if last line is not completed, keep it in buffer
	if len(lines[len(lines)-1]) != 0 {
		w.buf.Write(lines[len(lines)-1])
		lines = lines[:len(lines)-1]
	} else if len(lines) > 0 {
		lines = lines[:len(lines)-1]
	}

	return lines
}
