package slim

import (
	"encoding/json"
	"io"
)

type outputEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type OutputWriter struct {
	enc *json.Encoder
}

func NewOutputWriter(w io.Writer) *OutputWriter {
	return &OutputWriter{enc: json.NewEncoder(w)}
}

func (o *OutputWriter) Chunk(data json.RawMessage) error {
	return o.enc.Encode(outputEvent{Event: "chunk", Data: data})
}

func (o *OutputWriter) Done() error {
	return o.enc.Encode(outputEvent{Event: "done"})
}

func (o *OutputWriter) Error(code ErrorCode, msg string) error {
	return o.enc.Encode(outputEvent{
		Event: "error",
		Data:  SlimError{Code: code, Message: msg},
	})
}

func (o *OutputWriter) Message(stage string, msg string) error {
	return o.enc.Encode(outputEvent{
		Event: "message",
		Data: map[string]string{
			"stage":   stage,
			"message": msg,
		},
	})
}
