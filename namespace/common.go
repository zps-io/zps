package namespace

import (
	"github.com/solvent-io/zps/action"
)

type Namespace interface {
	GetValue(name string) string
	Columns() string
	Validate(manifest *action.Manifest) error
}

func Get(t string) Namespace {
	switch t {
	case "vcs":
		return NewVcsNs()
	}

	return nil
}
