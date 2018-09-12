package action

import (
	"fmt"
)

type SymLink struct {
	Path   string `json:"path" hcl:"path,label"`
	Owner  string `json:"owner" hcl:"owner,optional"`
	Group  string `json:"group" hcl:"group,optional"`
	Target string `json:"target" hcl:"target"`
}

func NewSymLink() *SymLink {
	return &SymLink{}
}

func (s *SymLink) Key() string {
	return s.Path
}

func (s *SymLink) Type() string {
	return "SymLink"
}

func (s *SymLink) Id() string {
	return fmt.Sprint(f.Type(), ".", f.Key())
}

func (s *SymLink) Condition() *bool {
	return nil
}

func (s *SymLink) MayFail() bool {
	return false
}

func (s *SymLink) IsValid() bool {
	if s.Path != "" && s.Owner != "" && s.Group != "" && s.Target != "" {
		return true
	}

	return false
}
