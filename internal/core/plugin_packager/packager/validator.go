package packager

func (p *Packager) Validate() error {
	// read manifest
	_, err := p.fetchManifest()
	if err != nil {
		return err
	}

	return nil
}
