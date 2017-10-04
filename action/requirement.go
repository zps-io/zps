package action

import (
	"strings"
)

type Requirement struct {
	Name      string `json:"name"`
	Method    string `json:"method"`
	Operation string `json:"operation,omitempty"`
	Version   string `json:"version,omitempty"`
}

func NewRequirement() *Requirement {
	return &Requirement{}
}

func (r *Requirement) Key() string {
	key := []string{r.Name, r.Method, r.Operation}
	return strings.Join(key, ":")
}

func (r *Requirement) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(r.Type()),
		r.Name,
		r.Method,
		r.Operation,
		r.Version,
	}, "|")
}

func (r *Requirement) Unique() string {
	key := []string{"requirement", r.Name, r.Method, r.Operation}
	return strings.Join(key, ":")
}

func (r *Requirement) Type() string {
	return "requirement"
}

func (r *Requirement) Valid() bool {
	if r.Name == "" {
		return false
	}

	if r.Method == "" {
		return false
	}

	return true
}
