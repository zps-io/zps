package action

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/hashicorp/hcl2/gohcl"
)

type Manifest struct {
	Actions Actions
	Zpf     *hcl.BodyContent

	index map[string]int
}

type JsonManifest struct {
	Zpkg        []Action `json:"zpkg"`
	Requirement []Action `json:"requirement,omitempty"`
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

func (m *Manifest) Load(from string, zhcl []byte) error {
	var err error
	var diag hcl.Diagnostics
	var zpf *hcl.File

	parser := hclparse.NewParser()

	if from == "hcl" {
		zpf, diag = parser.ParseHCL(zhcl, "manifest.hcl")
		if diag.HasErrors() {
			return diag
		}
	} else if from == "json" {
		zpf, diag = parser.ParseJSON(zhcl, "manifest.hcl")
		if diag.HasErrors() {
			return diag
		}
	}

	m.Zpf, diag = zpf.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "zpkg",
			},
			{
				Type: "requirement",
			},
			{
				Type: "meta",
			},
			{
				Type: "dir",
			},
			{
				Type: "file",
			},
			{
				Type: "symLink",
			},
		},
	})

	if diag.HasErrors() {
		return errors.New("invalid manifest schema")
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
	actions := m.Zpf.Blocks.OfType("zpkg")

	if len(actions) > 0 {
		for _, item := range actions {
			var object Zpkg

			err := gohcl.DecodeBody(item.Body, nil, &object)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(&object)
			} else {
				return errors.New("invalid zpkg action")
			}
			break
		}
	}

	return nil
}

func (m *Manifest) LoadRequirement() error {
	actions := m.Zpf.Blocks.OfType("requirement")

	if len(actions) > 0 {
		for _, item := range actions {
			var object Requirement

			err := gohcl.DecodeBody(item.Body, nil, &object)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(&object)
			} else {
				return errors.New("invalid requirement action")
			}
		}
	}

	return nil
}

func (m *Manifest) LoadMeta() error {
	actions := m.Zpf.Blocks.OfType("meta")

	if len(actions) > 0 {
		for _, item := range actions {
			var object Meta

			err := gohcl.DecodeBody(item.Body, nil, &object)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(&object)
			} else {
				return errors.New("invalid meta action")
			}
		}
	}

	return nil
}

func (m *Manifest) LoadDir() error {
	actions := m.Zpf.Blocks.OfType("dir")

	if len(actions) > 0 {
		for _, item := range actions {
			var object Dir

			err := gohcl.DecodeBody(item.Body, nil, &object)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(&object)
			} else {
				return errors.New("invalid dir action")
			}
		}
	}

	return nil
}

func (m *Manifest) LoadFile() error {
	actions := m.Zpf.Blocks.OfType("file")

	if len(actions) > 0 {
		for _, item := range actions {
			var object File

			err := gohcl.DecodeBody(item.Body, nil, &object)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(&object)
			} else {
				return errors.New("invalid file action")
			}
		}
	}

	return nil
}

func (m *Manifest) LoadSymLink() error {
	actions := m.Zpf.Blocks.OfType("symlink")

	if len(actions) > 0 {
		for _, item := range actions {
			var object SymLink

			err := gohcl.DecodeBody(item.Body, nil, &object)
			if err != nil {
				return err
			}

			if object.Valid() {
				m.Add(&object)
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
