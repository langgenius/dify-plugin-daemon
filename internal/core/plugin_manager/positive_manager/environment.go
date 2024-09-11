package positive_manager

import (
	"os"
)

func (r *PositivePluginRuntime) calculateChecksum() (string, error) {
	checksum, err := r.Decoder.Checksum()
	if err != nil {
		return "", err
	}

	return checksum, nil
}

func (r *PositivePluginRuntime) Checksum() (string, error) {
	if r.InnerChecksum == "" {
		checksum, err := r.calculateChecksum()
		if err != nil {
			return "", err
		}
		r.InnerChecksum = checksum
	}

	return r.InnerChecksum, nil
}

func (r *PositivePluginRuntime) Cleanup() {
	os.RemoveAll(r.WorkingPath)
}
