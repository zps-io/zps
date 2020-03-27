package config

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type TplConfigFile struct {
	Templates *TplConfig `hcl:"Template,block"`
}

type TplConfig struct {
	Name     string `hcl:"name,label"`
	Register string `hcl:"register"`
	Source   string `hcl:"source"`
	Output   string `hcl:"output"`

	Owner string `hcl:"owner,optional"`
	Group string `hcl:"group,optional"`
	Mode  string `hcl:"mode,optional"`
}

func (t TplConfig) ToHclFile() *hclwrite.File {
	file := hclwrite.NewEmptyFile()

	tpl := file.Body().AppendNewBlock("Template", []string{t.Name})
	tpl.Body().SetAttributeValue("register", cty.StringVal(t.Register))
	tpl.Body().SetAttributeValue("source", cty.StringVal(t.Source))
	tpl.Body().SetAttributeValue("output", cty.StringVal(t.Output))

	if t.Owner != "" {
		tpl.Body().SetAttributeValue("owner", cty.StringVal(t.Owner))
	}
	if t.Group != "" {
		tpl.Body().SetAttributeValue("group", cty.StringVal(t.Group))
	}
	if t.Mode != "" {
		tpl.Body().SetAttributeValue("mode", cty.StringVal(t.Mode))
	}

	return file
}
