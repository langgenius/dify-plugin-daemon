package entities

import "errors"

func (c *PluginConfiguration) Validate() error {
	if c.Module == "" {
		return errors.New("exec is required")
	}

	if c.Name == "" {
		return errors.New("name is required")
	}

	if c.Version == "" {
		return errors.New("version is required")
	}

	return nil
}
