package action

import (
	"strings"
)

type Dir struct {
	Path  string `json:"path"`
	Owner string `json:"owner"`
	Group string `json:"group"`
	Mode  string `json:"mode"`
}

func NewDir() *Dir {
	return &Dir{}
}

func (d *Dir) Key() string {
	return d.Path
}

func (d *Dir) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(d.Type()),
		d.Mode,
		d.Owner + ":" + d.Group,
		d.Path,
	}, "|")
}

func (d *Dir) Unique() string {
	key := []string{"dir", d.Path}
	return strings.Join(key, ":")
}

func (d *Dir) Type() string {
	return "dir"
}

func (d *Dir) Valid() bool {
	if d.Path == "" {
		return false
	}

	return true
}
