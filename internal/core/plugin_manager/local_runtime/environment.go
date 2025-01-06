package local_runtime

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (r *LocalPluginRuntime) InitEnvironment() error {
	var err error
	if r.Config.Meta.Runner.Language == constants.Python {
		err = r.InitPythonEnvironment()
	} else {
		return fmt.Errorf("unsupported language: %s", r.Config.Meta.Runner.Language)
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *LocalPluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	checksum, err := r.Checksum()
	if err != nil {
		return "", err
	}
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", r.Config.Identity(), checksum))
}
