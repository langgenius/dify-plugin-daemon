package decoder

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/langgenius-neko_0.0.2.difypkg
var pluginPackageWithoutVerificationField []byte

/*
Formerly, the plugin is all signed by langgenius but has no authorized category
*/
func TestVerifyPluginWithoutVerificationField(t *testing.T) {
	decoder, err := NewZipPluginDecoder(pluginPackageWithoutVerificationField)
	assert.NoError(t, err)

	verification, err := decoder.Verification(false)
	assert.NoError(t, err)
	assert.Nil(t, verification)

	verified := decoder.Verified()
	assert.True(t, verified)
}
