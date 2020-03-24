package config

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
