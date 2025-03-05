package local_runtime

import (
	"testing"

	version "github.com/hashicorp/go-version"
)

func TestGetPluginSdkVersion(t *testing.T) {
	var requirements = `
dify-plugin==0.0.1b70
gunicorn==20.1.0
`
	localRuntime := &LocalPluginRuntime{}
	version, err := localRuntime.getPluginSdkVersion(requirements)
	if err != nil {
		t.Fatalf("failed to get the version of the plugin sdk: %s", err)
	}

	if version != "0.0.1b70" {
		t.Fatalf("failed to get the correct version of the plugin sdk: %s", version)
	}

	var requirements2 = `
python-dotenv==1.0.1
dify-plugin~=0.0.1b70
`
	version, err = localRuntime.getPluginSdkVersion(requirements2)
	if err != nil {
		t.Fatalf("failed to get the version of the plugin sdk: %s", err)
	}

	if version != "0.0.1b70" {
		t.Fatalf("failed to get the correct version of the plugin sdk: %s", version)
	}

	var requirements3 = `
# comment
dify_plugin==0.0.1b70
# comment
gunicorn~=20.1.0
`
	version, err = localRuntime.getPluginSdkVersion(requirements3)
	if err != nil {
		t.Fatalf("failed to get the version of the plugin sdk: %s", err)
	}

	if version != "0.0.1b70" {
		t.Fatalf("failed to get the correct version of the plugin sdk: %s", version)
	}

	var requirements4 = `
dify_plugin~=0.0.1b70
`
	version, err = localRuntime.getPluginSdkVersion(requirements4)
	if err != nil {
		t.Fatalf("failed to get the version of the plugin sdk: %s", err)
	}

	if version != "0.0.1b70" {
		t.Fatalf("failed to get the correct version of the plugin sdk: %s", version)
	}

	var requirements5 = `
dify-plugin==0.0.1
`
	version, err = localRuntime.getPluginSdkVersion(requirements5)
	if err != nil {
		t.Fatalf("failed to get the version of the plugin sdk: %s", err)
	}

	if version != "0.0.1" {
		t.Fatalf("failed to get the correct version of the plugin sdk: %s", version)
	}
}

func TestCompareVersion(t *testing.T) {
	v1, err := version.NewVersion("0.0.1b70")
	if err != nil {
		t.Fatalf("failed to create the version: %s", err)
	}
	v2, err := version.NewVersion("0.0.1")
	if err != nil {
		t.Fatalf("failed to create the version: %s", err)
	}

	if v1.GreaterThan(v2) {
		t.Fatalf("v1 should be less than v2: %s", v1)
	}

	v3, err := version.NewVersion("0.0.1b7")
	if err != nil {
		t.Fatalf("failed to create the version: %s", err)
	}

	v4, err := version.NewVersion("0.0.1b70")
	if err != nil {
		t.Fatalf("failed to create the version: %s", err)
	}

	if v3.GreaterThan(v4) {
		t.Fatalf("v3 should be less than v4: %s", v3)
	}
}
