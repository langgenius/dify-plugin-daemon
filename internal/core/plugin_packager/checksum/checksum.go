package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"slices"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
)

func CalculateChecksum(plugin decoder.PluginDecoder) (string, error) {
	m := map[string][]byte{}

	sha256 := func(data []byte) []byte {
		sha := sha256.New()
		sha.Write(data)
		return sha.Sum(nil)
	}

	if err := plugin.Walk(func(filename string, dir string) error {
		var err error
		content, err := plugin.ReadFile(path.Join(dir, filename))
		if err != nil {
			return err
		}
		m[path.Join(dir, filename)] = sha256(content)
		return nil
	}); err != nil {
		return "", err
	}

	// sort the keys, ensure the order is consistent
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	data := make([]byte, 0, len(m)*(32+32))
	for _, k := range keys {
		data = append(data, sha256([]byte(k))...)
		data = append(data, m[k]...)
	}

	return hex.EncodeToString(sha256(data)), nil
}
