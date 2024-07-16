package packager

type Packager struct {
	wp string // working path

	manifest string // manifest file path
}

func NewPackager(plugin_path string) *Packager {
	return &Packager{
		wp:       plugin_path,
		manifest: "manifest.yaml",
	}
}

func (p *Packager) Pack() error {
	return nil
}
