package namespace

import (
	"github.com/solvent-io/zps/action"
)

type VcsNs struct {
	manifest *action.Manifest
}

func NewVcsNs() *VcsNs {
	vcs := &VcsNs{}
	return vcs
}

func (v *VcsNs) GetValue(name string) string {
	return v.manifest.Meta("vcs", name).Value
}

func (v *VcsNs) Columns() string {
	return v.manifest.Meta("vcs", "uri").Value
}

func (v *VcsNs) Validate(manifest *action.Manifest) error {
	return nil
}
