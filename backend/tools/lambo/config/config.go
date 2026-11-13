package config

const (
	Version       = "0.1.0"
	WireCmd       = "github.com/google/wire/cmd/wire@v0.7.0"
	LamboCmd      = "github.com/Edward-Jackie/lambo/backend/tools/lambo/cmd/lambo@latest"
	RepoBase      = "https://github.com/Edward-Jackie/lambo.git"
	RunExcludeDir = ".git,.idea,.artifacts,.tmp,tmp,vendor,node_modules,frontend,tools"
	RunIncludeExt = "go,html,yaml,yml,toml,ini,json,xml,tpl,tmpl"
)
