package action

// TODO handle overrides (file from Zpkgfile overrides vals from original

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Manifest struct {
	Zpkg *Zpkg `hcl:"Zpkg,block" json:"zpkg"`

	Tags []*Tag `hcl:"Tag,block" json:"tag,omitempty"`

	Requirements []*Requirement `hcl:"Requirement,block" json:"requirement,omitempty"`

	Dirs     []*Dir     `hcl:"Dir,block" json:"dir,omitempty"`
	Files    []*File    `hcl:"File,block" json:"file,omitempty"`
	SymLinks []*SymLink `hcl:"SymLink,block" json:"symlink,omitempty"`
}

func NewManifest() *Manifest {
	return &Manifest{}
}

func (m *Manifest) Add(action Action) {
	switch action.Type() {
	case "Tag":
		m.Tags = append(m.Tags, action.(*Tag))
	case "Requirement":
		m.Requirements = append(m.Requirements, action.(*Requirement))
	case "Dir":
		m.Dirs = append(m.Dirs, action.(*Dir))
	case "File":
		m.Files = append(m.Files, action.(*File))
	case "SymLink":
		m.SymLinks = append(m.SymLinks, action.(*SymLink))
	}
}

func (m *Manifest) Section(filters ...string) Actions {
	var items []Action

	for _, filter := range filters {
		switch filter {
		case "Tag":
			for _, item := range m.Tags {
				items = append(items, item)
			}
		case "Requirement":
			for _, item := range m.Requirements {
				items = append(items, item)
			}
		case "Dir":
			for _, item := range m.Dirs {
				items = append(items, item)
			}
		case "File":
			for _, item := range m.Files {
				items = append(items, item)
			}
		case "SymLink":
			for _, item := range m.SymLinks {
				items = append(items, item)
			}
		}
	}

	return items
}

func (m *Manifest) Actions() Actions {
	var actions Actions

	fs := m.Section("Dir", "File", "SymLink")
	sort.Sort(fs)

	actions = append(actions, m.Zpkg)
	actions = append(actions, m.Section("Tag")...)
	actions = append(actions, m.Section("Requirement")...)
	actions = append(actions, fs...)

	return actions
}

func (m *Manifest) Validate() error {
	var actions Actions
	actions = m.Section("Dir", "File", "SymLink")

	sort.Sort(actions)
	for index, action := range actions {
		prev := index - 1
		if prev != -1 {
			if action.Key() == actions[prev].Key() {
				return errors.New(fmt.Sprint(
					"Action Conflicts:\n",
					strings.ToUpper(actions[prev].Type()), " => ", actions[prev].Key(), "\n",
					strings.ToUpper(action.Type()), " => ", action.Key()))
			}
		}
	}

	return nil
}

func (m *Manifest) ToJson() string {
	out, _ := json.Marshal(m)

	return string(out)
}

func (m *Manifest) Load(manifest string) error {
	return json.Unmarshal([]byte(manifest), m)
}