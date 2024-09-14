package local_manager

import (
	"fmt"
	"os"
	"path"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (r *LocalPluginRuntime) InitEnvironment() error {
	if _, err := os.Stat(path.Join(r.State.AbsolutePath, ".installed")); err == nil {
		return nil
	}

	var err error
	if r.Config.Meta.Runner.Language == constants.Python {
		err = r.InitPythonEnvironment()
	} else {
		return fmt.Errorf("unsupported language: %s", r.Config.Meta.Runner.Language)
	}

	if err != nil {
		return err
	}

	// create .installed file
	f, err := os.Create(path.Join(r.State.AbsolutePath, ".installed"))
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}

func (r *LocalPluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	checksum, err := r.Checksum()
	if err != nil {
		return "", err
	}
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", r.Config.Identity(), checksum))
}
