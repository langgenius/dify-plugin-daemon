package decoder_test

import (
	_ "embed"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/langgenius-neko_0.0.2.difypkg
var pluginPackageWithoutVerificationField []byte

/*
Formerly, the plugin is all signed by langgenius but has no authorized category
*/
func TestVerifyPluginWithoutVerificationField(t *testing.T) {
	decoder, err := decoder.NewZipPluginDecoder(pluginPackageWithoutVerificationField)
	assert.NoError(t, err)

	verification, err := decoder.Verification(false)
	assert.NoError(t, err)
	assert.Nil(t, verification)

	verified := decoder.Verified()
	assert.True(t, verified)
}
