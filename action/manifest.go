package action

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

type Manifest struct {
	Actions Actions
	Zpf     *ast.File

	index map[string]int
}

type JsonManifest struct {
	Zpkg        []Action `json:"zpkg"`
	Requirement []Action `json:"requirement"`
	Meta        []Action `json:"meta,omitempty"`
	Dir         []Action `json:"dir,omitempty"`
	File        []Action `json:"file,omitempty"`
	SymLink     []Action `json:"symlink,omitempty"`
}

func NewManifest() *Manifest {
	manifest := &Manifest{}
	manifest.index = make(map[string]int)
	return manifest
}

func (m *Manifest) Load(zhcl string) error {
	var err error

	m.Zpf, err = hcl.Parse(zhcl)
	if err != nil {
		return err
	}

	_, ok := m.Zpf.Node.(*ast.ObjectList)
	if !ok {
		return errors.New("Zpkgfile doesn't contain a root object")
	}

	err = m.LoadZpkg()
	if err != nil {
		return err
	}

	err = m.LoadRequirement()
	if err != nil {
		return err
	}

	err = m.LoadMeta()
	if err != nil {
		return err
	}

	err = m.LoadDir()
	if err != nil {
		return err
	}

	err = m.LoadFile()
	if err != nil {
		return err
	}

	err = m.LoadSymLink()
	if err != nil {
		return err
	}

	return err
}

func (m *Manifest) Add(action Action) {
	if index, ok := m.index[action.Unique()]; !ok {
		m.Actions = append(m.Actions, action)
		m.index[action.Unique()] = len(m.Actions) - 1
	} else {
		// Shuffle stuff around as this is an override
		action = m.Actions[index]
		m.Actions = append(m.Actions[:index], m.Actions[index+1:]...)
		m.Actions = append(m.Actions, action)
		m.index[action.Unique()] = len(m.Actions) - 1
		m.Index()
	}
}

func (m *Manifest) Meta(namespace string, name string) *Meta {
	key := strings.Join([]string{"meta", namespace, name}, ":")
	if val, ok := m.index[key]; ok {
		return m.Actions[val].(*Meta)
	}

	return &Meta{}
}

func (m *Manifest) Section(filters ...string) []Action {
	var items []Action

	for _, item := range m.Actions {
		for _, filter := range filters {
			if item.Type() == filter {
				items = append(items, item)
			}
		}
	}

	return items
}

func (m *Manifest) Sort() {
	sort.Sort(m.Actions)

	// Rebuild the index
	m.Index()
}

func (m *Manifest) Index() {
	for index, action := range m.Actions {
		m.index[action.Unique()] = index
	}
}

// Section loaders

func (m *Manifest) LoadZpkg() error {
	actions, _ := m.Zpf.Node.(*ast.ObjectList)

	if zpkg := actions.Filter("zpkg"); len(zpkg.Items) > 0 {
		for _, item := range zpkg.Items {
			var object *Zpkg

			err := hcl.DecodeObject(&object, item.Val)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(object)
			} else {
				return errors.New("invalid zpkg action")
			}
			break
		}
	}

	return nil
}

func (m *Manifest) LoadRequirement() error {
	actions, _ := m.Zpf.Node.(*ast.ObjectList)

	if requirement := actions.Filter("requirement"); len(requirement.Items) > 0 {
		for _, item := range requirement.Items {
			var object *Requirement

			err := hcl.DecodeObject(&object, item.Val)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(object)
			} else {
				return errors.New("invalid requirement action")
			}
		}
	}

	return nil
}

func (m *Manifest) LoadMeta() error {
	actions, _ := m.Zpf.Node.(*ast.ObjectList)

	if meta := actions.Filter("meta"); len(meta.Items) > 0 {
		for _, item := range meta.Items {
			var object *Meta

			err := hcl.DecodeObject(&object, item.Val)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(object)
			} else {
				return errors.New("invalid meta action")
			}
		}
	}

	return nil
}

func (m *Manifest) LoadDir() error {
	actions, _ := m.Zpf.Node.(*ast.ObjectList)

	if dir := actions.Filter("dir"); len(dir.Items) > 0 {
		for _, item := range dir.Items {
			var object *Dir

			err := hcl.DecodeObject(&object, item.Val)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(object)
			} else {
				return errors.New("invalid dir action")
			}
		}
	}

	return nil
}

func (m *Manifest) LoadFile() error {
	actions, _ := m.Zpf.Node.(*ast.ObjectList)

	if file := actions.Filter("file"); len(file.Items) > 0 {
		for _, item := range file.Items {
			var object *File

			err := hcl.DecodeObject(&object, item.Val)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(object)
			} else {
				return errors.New("invalid file action")
			}
		}
	}

	return nil
}

func (m *Manifest) LoadSymLink() error {
	actions, _ := m.Zpf.Node.(*ast.ObjectList)

	if symlink := actions.Filter("symlink"); len(symlink.Items) > 0 {
		for _, item := range symlink.Items {
			var object *SymLink

			err := hcl.DecodeObject(&object, item.Val)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(object)
			} else {
				return errors.New("invalid symlink action")
			}
		}
	}

	return nil
}

// To Json representation
func (m *Manifest) Json() string {
	manifest := &JsonManifest{}
	m.Sort()

	manifest.Zpkg = m.Section("zpkg")

	requirement := m.Section("requirement")
	if len(requirement) > 0 {
		manifest.Requirement = requirement
	}

	meta := m.Section("meta")
	if len(meta) > 0 {
		manifest.Meta = meta
	}

	dir := m.Section("dir")
	if len(dir) > 0 {
		manifest.Dir = dir
	}

	file := m.Section("file")
	if len(file) > 0 {
		manifest.File = file
	}

	symlink := m.Section("symlink")
	if len(symlink) > 0 {
		manifest.SymLink = symlink
	}

	out, _ := json.Marshal(manifest)

	return string(out)
}
