package action

import (
	"strings"
)

type Meta struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Value     string `json:"value"`
}

func NewMeta() *Meta {
	return &Meta{}
}

func (m *Meta) Key() string {
	key := []string{m.Namespace, m.Name}
	return strings.Join(key, ":")
}

func (m *Meta) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(m.Type()),
		m.Name,
		m.Value,
	}, "|")
}

func (m *Meta) Unique() string {
	key := []string{"meta", m.Namespace, m.Name}
	return strings.Join(key, ":")
}

func (m *Meta) Type() string {
	return "meta"
}

func (m *Meta) Valid() bool {
	if m.Namespace == "" {
		return false
	}

	if m.Name == "" {
		return false
	}

	if m.Value == "" {
		return false
	}

	return true
}
