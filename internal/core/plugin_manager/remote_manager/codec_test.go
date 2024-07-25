package remote_manager

import "testing"

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
