package action

// TODO handle overrides (file from Zpkgfile overrides vals from original

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Manifest struct {
	Zpkg *Zpkg `hcl:"Zpkg,block"`

	Tags []*Tag `hcl:"Tag,block"`

	Requirements []*Requirement `hcl:"Requirement,block"`

	Dirs     []*Dir     `hcl:"Dir,block"`
	Files    []*File    `hcl:"File,block"`
	SymLinks []*SymLink `hcl:"SymLink,block"`
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

func (m *Manifest) Section(filters ...string) []Action {
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