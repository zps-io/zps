package config

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

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

func (c *Config) ToHclFile() *hclwrite.File {
	file := hclwrite.NewEmptyFile()

	cfg := file.Body().AppendNewBlock("Config", []string{c.Namespace})

	for _, profile := range c.Profiles {
		pr := cfg.Body().AppendNewBlock("profile", []string{profile.Name})
		pr.Body().SetAttributeValue("values", profile.Values)
	}

	return file
}
