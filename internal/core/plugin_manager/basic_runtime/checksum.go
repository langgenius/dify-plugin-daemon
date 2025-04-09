package basic_runtime

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

type BasicChecksum struct {
	MediaTransport

	WorkingPath string
	// plugin decoder used to manage the plugin
	Decoder decoder.PluginDecoder

	InnerChecksum string
}

func (r *BasicChecksum) calculateChecksum() (string, error) {
	checksum, err := r.Decoder.Checksum()
	if err != nil {
		return "", err
	}

	return checksum, nil
}

func (r *BasicChecksum) Checksum() (string, error) {
	if r.InnerChecksum == "" {
		checksum, err := r.calculateChecksum()
		if err != nil {
			return "", err
		}
		r.InnerChecksum = checksum
	}

	return r.InnerChecksum, nil
}

func (r *BasicChecksum) Cleanup() {
	os.RemoveAll(r.WorkingPath)
}
