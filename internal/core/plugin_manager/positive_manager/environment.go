package positive_manager

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/checksum"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
)

func (r *PositivePluginRuntime) calculateChecksum() string {
	plugin_decoder, err := decoder.NewFSPluginDecoder(r.LocalPackagePath)
	if err != nil {
		return ""
	}

	checksum, err := checksum.CalculateChecksum(plugin_decoder)
	if err != nil {
		return ""
	}

	return checksum
}

func (r *PositivePluginRuntime) Checksum() string {
	if r.InnerChecksum == "" {
		r.InnerChecksum = r.calculateChecksum()
	}

	return r.InnerChecksum
}

func (r *PositivePluginRuntime) Cleanup() {
	os.RemoveAll(r.WorkingPath)
}
