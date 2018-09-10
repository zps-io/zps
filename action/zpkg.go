package action

import (
	"errors"
	"strings"
)

type Zpkg struct {
	Name        string `json:"name,omitempty" hcl:"name"`
	Version     string `json:"version" hcl:"version"`
	Publisher   string `json:"publisher,omitempty" hcl:"publisher"`
	Category    string `json:"category,omitempty" hcl:"category"`
	Uri         string `json:"uri"`
	Arch        string `json:"arch" hcl:"arch"`
	Os          string `json:"os" hcl:"os"`
	Summary     string `json:"summary" hcl:"summary"`
	Description string `json:"description" hcl:"description"`
}

func NewZpkg() *Zpkg {
	return &Zpkg{}
}

func (z *Zpkg) Key() string {
	return z.Name
}

func (z *Zpkg) Columns() string {
	return strings.Join(
		[]string{
			z.Name,
			z.Arch,
		},
		"|",
	)
}

func (z *Zpkg) Unique() string {
	key := []string{"zpkg", z.Name}
	return strings.Join(key, ":")
}

func (z *Zpkg) Type() string {
	return "zpkg"
}

func (z *Zpkg) Validate() error {
	var err error = nil

	err = z.validateName()
	if err != nil {
		return err
	}

	err = z.validateVersion()
	if err != nil {
		return err
	}

	err = z.validatePublisher()
	if err != nil {
		return err
	}

	err = z.validateCategory()
	if err != nil {
		return err
	}

	err = z.validateArch()
	if err != nil {
		return err
	}

	err = z.validateOs()
	if err != nil {
		return err
	}

	err = z.validateSummary()
	if err != nil {
		return err
	}

	err = z.validateDescription()
	if err != nil {
		return err
	}

	return err
}

func (z *Zpkg) Valid() bool {
	err := z.Validate()
	if err != nil {
		return false
	}

	return true
}

func (z *Zpkg) validateName() error {
	if z.Name == "" && z.Uri == "" {
		return errors.New("action zpkg:name required")
	}

	return nil
}

func (z *Zpkg) validatePublisher() error {
	if z.Publisher == "" && z.Uri == "" {
		return errors.New("action zpkg:name required")
	}

	return nil
}

func (z *Zpkg) validateVersion() error {
	if z.Version == "" && z.Uri == "" {
		return errors.New("action zpkg:version required")
	}

	return nil
}

func (z *Zpkg) validateCategory() error {
	if z.Category == "" && z.Uri == "" {
		return errors.New("action zpkg:category required")
	}

	return nil
}

func (z *Zpkg) validateArch() error {
	if z.Arch != "x86_64" && z.Arch != "arm64" {
		return errors.New("action zpkg:arch is unsupported")
	}

	return nil
}

func (z *Zpkg) validateOs() error {
	// If only we could lookup a constant in a map?
	if z.Os != "darwin" && z.Os != "linux" && z.Os != "freebsd" {
		return errors.New("action zpkg:os is unsupported")
	}

	return nil
}

func (z *Zpkg) validateSummary() error {
	if z.Summary == "" {
		return errors.New("action zpkg:summary required")
	}

	return nil
}

func (z *Zpkg) validateDescription() error {
	if z.Description == "" {
		return errors.New("action zpkg:description required")
	}

	return nil
}
