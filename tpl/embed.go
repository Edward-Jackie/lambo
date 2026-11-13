package tpl

import "embed"

//go:embed create/*.tpl frontend/*.tpl
var CreateTemplateFS embed.FS
