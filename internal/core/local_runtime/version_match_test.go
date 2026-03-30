package local_runtime

import (
	"testing"

	version "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func TestGetPluginSdkVersion(t *testing.T) {
	var requirements = `
dify-plugin==0.0.1b70
gunicorn==20.1.0
`
	localRuntime := &LocalPluginRuntime{}
	version, err := localRuntime.getPluginSdkVersion(requirements)
	assert.Nil(t, err)
	assert.Equal(t, "0.0.1b70", version)

	var requirements2 = `
python-dotenv==1.0.1
dify-plugin~=0.0.1b70
`
	version, err = localRuntime.getPluginSdkVersion(requirements2)
	assert.Nil(t, err)
	assert.Equal(t, "0.0.1b70", version)

	var requirements3 = `
# comment
dify_plugin==0.0.1b70
# comment
gunicorn~=20.1.0
`
	version, err = localRuntime.getPluginSdkVersion(requirements3)
	assert.Nil(t, err)
	assert.Equal(t, "0.0.1b70", version)

	var requirements4 = `
dify_plugin~=0.0.1b70
`
	version, err = localRuntime.getPluginSdkVersion(requirements4)
	assert.Nil(t, err)
	assert.Equal(t, "0.0.1b70", version)

	var requirements5 = `
dify-plugin==0.0.1
`
	version, err = localRuntime.getPluginSdkVersion(requirements5)
	assert.Nil(t, err)
	assert.Equal(t, "0.0.1", version)

	var requirements6 = `
dify-plugin>=0.1.0,<0.2.0
`
	version, err = localRuntime.getPluginSdkVersion(requirements6)
	assert.Nil(t, err)
	assert.Equal(t, "0.2.0", version)
}

func TestCompareVersion(t *testing.T) {
	v1, err := version.NewVersion("0.0.1b70")
	assert.Nil(t, err)
	v2, err := version.NewVersion("0.0.1")
	assert.Nil(t, err)

	assert.False(t, v1.GreaterThan(v2))

	v3, err := version.NewVersion("0.0.1b7")
	assert.Nil(t, err)
	v4, err := version.NewVersion("0.0.1b70")
	assert.Nil(t, err)

	assert.True(t, v3.LessThan(v4))
}
