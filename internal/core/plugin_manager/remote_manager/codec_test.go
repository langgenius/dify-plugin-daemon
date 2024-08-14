package remote_manager

import (
	"testing"
)

func TestCodec(t *testing.T) {
	codec := &codec{}
	liens := codec.getLines([]byte("test\n"))
	if len(liens) != 1 {
		t.Error("getLines failed")
	}

	liens = codec.getLines([]byte("test\ntest"))
	if len(liens) == 2 {
		t.Error("getLines failed")
	}

	liens = codec.getLines([]byte("\n"))
	if len(liens) != 1 {
		t.Error("getLines failed")
	}
}

func TestCodec2(t *testing.T) {
	codec := &codec{}

	msg := "9c3df1b4-6daf-4cb4-bcaa-3f05a2dbc3a1\n{\"version\":\"1.0.0\",\"type\":\"plugin\",\"author\":\"Yeuoly\",\"name\":\"ci_test\",\"created_at\":\"2024-08-14T19:48:04.867581+08:00\",\"resource\":{\"memory\":1,\"storage\":1,\"permission\":null},\"plugins\":[\"test\"],\"execution\":{\"install\":\"echo 'hello'\",\"launch\":\"echo 'hello'\"},\"meta\":{\"version\":\"0.0.1\",\"arch\":[\"amd64\"],\"runner\":{\"language\":\"python\",\"version\":\"3.12\",\"entrypoint\":\"main\"}}}"

	lines := codec.getLines([]byte(msg))
	if len(lines) != 1 {
		if string(lines[0]) != msg[:len(lines[0])] {
			t.Error("getLines failed")
		}
	}
}
