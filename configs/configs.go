package configs

import (
	_ "embed"
)

//go:embed config.dev.yaml
var DefaultConfigBytes []byte
