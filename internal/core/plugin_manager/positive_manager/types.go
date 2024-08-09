package positive_manager

import (
	"errors"
)

type PositivePluginRuntime struct {
	LocalPath string
}

func (r *PositivePluginRuntime) DockerImage() (string, error) {
	return "", errors.New("not implemented")
}
