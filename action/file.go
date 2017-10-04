package action

import (
	"strings"
)

type File struct {
	Path  string `json:"path"`
	Owner string `json:"owner"`
	Group string `json:"group"`
	Mode  string `json:"mode"`

	Hash   string `json:"hash"`
	Offset int    `json:"offset"`
	Csize  int    `json:"csize"`
	Size   int    `json:"size"`
}

func NewFile() *File {
	return &File{}
}

func (f *File) Key() string {
	return f.Path
}

func (f *File) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(f.Type()),
		f.Mode,
		f.Owner + ":" + f.Group,
		f.Path,
	}, "|")
}

func (f *File) Unique() string {
	key := []string{"file", f.Path}
	return strings.Join(key, ":")
}

func (f *File) Type() string {
	return "file"
}

func (f *File) Valid() bool {
	if f.Path == "" {
		return false
	}

	return true
}
