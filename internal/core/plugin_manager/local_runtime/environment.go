package local_runtime

import (
	"errors"
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func (r *LocalPluginRuntime) InitEnvironment(decoder *decoder.ZipPluginDecoder) error {
	// extract plugin to working path
	err := r.extractPluginToWorkingPath(decoder)
	if err != nil {
		return errors.Join(err, fmt.Errorf("extract plugin to working path error"))
	}

	// initialize environment like dependencies,
	switch r.Config.Meta.Runner.Language {
	case constants.Python:
		err = r.InitPythonEnvironment()
	default:
		return fmt.Errorf("unsupported language: %s", r.Config.Meta.Runner.Language)
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *LocalPluginRuntime) extractPluginToWorkingPath(decoder *decoder.ZipPluginDecoder) error {
	// extract the plugin to working path
	// check if working path exists and if it's empty
	if _, err := os.Stat(r.State.WorkingPath); err != nil {
		if os.IsNotExist(err) {
			// create the working path
			if err := os.MkdirAll(r.State.WorkingPath, 0755); err != nil {
				return errors.Join(err, fmt.Errorf("create working directory error"))
			}

			// extract the plugin to working path
			if err = decoder.ExtractTo(r.State.WorkingPath); err != nil {
				return errors.Join(err, fmt.Errorf("extract plugin to working directory error"))
			}
		} else {
			return errors.Join(err, fmt.Errorf("check working directory error"))
		}
	} else {
		// check if the working path is empty
		if entries, err := os.ReadDir(r.State.WorkingPath); err != nil {
			return errors.Join(err, fmt.Errorf("check working directory error"))
		} else {
			if len(entries) == 0 {
				// extract the plugin to working path
				if err = decoder.ExtractTo(r.State.WorkingPath); err != nil {
					return errors.Join(err, fmt.Errorf("extract plugin to working directory error"))
				}
			}
		}
	}

	return nil
}

// return nil if environment is valid, otherwise return error
func (r *LocalPluginRuntime) EnvironmentValidation() error {
	if r.Config.Meta.Runner.Language == constants.Python {
		_, err := r.checkPythonVirtualEnvironment()
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("unsupported language: %s", r.Config.Meta.Runner.Language)
}

func (r *LocalPluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	checksum, err := r.Checksum()
	if err != nil {
		return "", err
	}
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", r.Config.Identity(), checksum))
}
