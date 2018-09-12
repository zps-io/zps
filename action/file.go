package action

import (
	"fmt"
)

type File struct {
	Path  string `json:"path" hcl:"path,label"`
	Owner string `json:"owner" hcl:"owner,optional"`
	Group string `json:"group" hcl:"group,optional"`
	Mode  string `json:"mode" hcl:"mode,optional"`

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

func (f *File) Type() string {
	return "File"
}

func (f *File) Id() string {
	return fmt.Sprint(f.Type(), ".", f.Key())
}

func (f *File) Condition() *bool {
	return nil
}

func (f *File) MayFail() bool {
	return false
}

func (f *File) IsValid() bool {
	if f.Path != "" && f.Owner != "" && f.Group != "" && f.Mode != "" {
		return true
	}

	return false
}
