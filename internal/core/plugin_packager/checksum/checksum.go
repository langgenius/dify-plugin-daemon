package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"path"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func CalculateChecksum(plugin decoder.PluginDecoder) (string, error) {
	m := map[string]any{}

	if err := plugin.Walk(func(filename string, dir string) error {
		var err error
		m[path.Join(dir, filename)], err = plugin.ReadFile(path.Join(dir, filename))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return "", err
	}

	str := parser.MarshalJsonBytes(m)
	sha := sha256.New()
	sha.Write(str)
	return hex.EncodeToString(sha.Sum(nil)), nil
}
