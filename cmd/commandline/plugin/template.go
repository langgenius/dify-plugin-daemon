package plugin

import (
	_ "embed"
)

//go:embed templates/README.md
var README []byte

//go:embed templates/.env.example
var ENV_EXAMPLE []byte

//go:embed templates/PRIVACY.md
var PRIVACY []byte

//go:embed templates/.github/workflows/plugin-publish.yml
var PLUGIN_PUBLISH_WORKFLOW []byte
