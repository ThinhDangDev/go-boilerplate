package templates

import "embed"

//go:embed all:base all:features
var Templates embed.FS
