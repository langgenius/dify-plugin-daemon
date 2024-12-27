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
