package config

import "github.com/zclconf/go-cty/cty"

type ConfigFile struct {
	Configs *Config `hcl:"Config,block"`
}

type Config struct {
	Namespace string     `hcl:"namespace,label"`
	Profiles  []*Profile `hcl:"profile,block"`
}

type Profile struct {
	Name   string    `hcl:"name,label"`
	Values cty.Value `hcl:"values"`
}
