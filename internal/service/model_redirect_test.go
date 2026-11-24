package service

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/stretchr/testify/require"
)

func TestVerifiedModelRedirectProviderRequiresPublisherAndExactProvider(t *testing.T) {
	declaration := plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{Author: "langgenius"},
		Verified:                               true,
		Model:                                  &plugin_entities.ModelProviderDeclaration{Provider: "destination"},
	}
	official := &decoder.Verification{AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS}
	community := &decoder.Verification{AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_COMMUNITY}
	require.True(t, isVerifiedModelRedirectProvider(&declaration, official, "destination"))
	require.True(t, isVerifiedModelRedirectProvider(&declaration, nil, "destination"))
	require.False(t, isVerifiedModelRedirectProvider(&declaration, community, "destination"))
	require.False(t, isVerifiedModelRedirectProvider(&declaration, official, "other"))
	declaration.Verified = false
	require.False(t, isVerifiedModelRedirectProvider(&declaration, official, "destination"))
}

func TestModelRedirectPackageRejectsUnverifiedManifestAndCategory(t *testing.T) {
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for filename, content := range map[string]string{
		"manifest.yaml":           "author: langgenius\nname: destination\nversion: 1.0.0\nverified: true\nplugins:\n  models:\n    - provider.yaml\n",
		"provider.yaml":           "provider: destination\n",
		".verification.dify.json": `{"authorized_category":"langgenius"}`,
	} {
		entry, err := writer.Create(filename)
		require.NoError(t, err)
		_, err = entry.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	const maxPackageSize = 1 << 20
	pkg, err := decoder.NewZipPluginDecoderWithSizeLimit(archive.Bytes(), maxPackageSize)
	require.NoError(t, err)
	defer pkg.Close()
	uid, err := pkg.UniqueIdentity()
	require.NoError(t, err)
	require.False(t, verifyModelRedirectPackage(maxPackageSize, archive.Bytes(), uid, "destination"))
}
