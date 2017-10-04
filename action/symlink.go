package action

import (
	"strings"
)

type SymLink struct {
	Path   string `json:"path"`
	Owner  string `json:"owner"`
	Group  string `json:"group"`
	Target string `json:"target"`
}

func NewSymLink() *SymLink {
	return &SymLink{}
}

func (s *SymLink) Key() string {
	return s.Path
}

func (s *SymLink) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(s.Type()),
		s.Owner + ":" + s.Group,
		s.Path,
		s.Target,
	}, "|")
}

func (s *SymLink) Unique() string {
	key := []string{"symlink", s.Path}
	return strings.Join(key, ":")
}

func (s *SymLink) Type() string {
	return "symlink"
}

func (s *SymLink) Valid() bool {
	if s.Path == "" {
		return false
	}

	return true
}
