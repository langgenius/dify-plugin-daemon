package packager

import (
	"errors"
	"fmt"
)

func (p *Packager) Validate() error {
	// read manifest
	_, err := p.fetchManifest()
	if err != nil {
		return err
	}

	// check assets valid
	err = p.decoder.CheckAssetsValid()
	if err != nil {
		return errors.Join(err, fmt.Errorf("assets invalid"))
	}

	return nil
}
