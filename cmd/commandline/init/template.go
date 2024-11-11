package init

import (
	_ "embed"
)

//go:embed templates/README.md
var README []byte
