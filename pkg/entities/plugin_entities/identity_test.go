package plugin_entities

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/system"
)

func TestPluginUniqueIdentifier(t *testing.T) {
	i, err := NewPluginUniqueIdentifier("langgenius/test" + system.DelimiterFLag + "1.0.0@1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("NewPluginUniqueIdentifier() returned an error: %v", err)
	}
	if i.Author() != "langgenius" {
		t.Fatalf("Author() = %s; want langgenius", i.Author())
	}
	if i.PluginID() != "langgenius/test" {
		t.Fatalf("PluginID() = %s; want langgenius/test", i.PluginID())
	}
	if i.Version() != "1.0.0" {
		t.Fatalf("Version() = %s; want 1.0.0", i.Version())
	}
	if i.Checksum() != "1234567890abcdef1234567890abcdef1234567890abcdef" {
		t.Fatalf("Checksum() = %s; want 1234567890abcdef1234567890abcdef1234567890abcdef", i.Checksum())
	}

	_, err = NewPluginUniqueIdentifier("test" + system.DelimiterFLag + "1.0.0@1234567890abcdef1234567890abcdef1234567890abcdef")
	if err != nil {
		t.Fatalf("NewPluginUniqueIdentifier() returned an error: %v", err)
	}

	_, err = NewPluginUniqueIdentifier("1.0.0@1234567890abcdef1234567890abcdef1234567890abcdef")
	if err == nil {
		t.Fatalf("NewPluginUniqueIdentifier() returned nil error for invalid identifier")
	}

	_, err = NewPluginUniqueIdentifier("1234567890abcdef1234567890abcdef1234567890abcdef")
	if err == nil {
		t.Fatalf("NewPluginUniqueIdentifier() returned nil error for invalid identifier")
	}

	_, err = NewPluginUniqueIdentifier("langgenius/test" + system.DelimiterFLag + "1.0.0@123456")
	if err == nil {
		t.Fatalf("NewPluginUniqueIdentifier() returned nil error for invalid identifier")
	}

	_, err = NewPluginUniqueIdentifier("langgenius/test" + system.DelimiterFLag + "1.0.0")
	if err == nil {
		t.Fatalf("NewPluginUniqueIdentifier() returned nil error for invalid identifier")
	}
}
