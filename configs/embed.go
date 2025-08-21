// Package configs provides embedded configuration files for the CM application.
package configs

import _ "embed"

// DefaultConfigYAML contains the default configuration file content.
//
//go:embed default.yaml
var DefaultConfigYAML []byte
