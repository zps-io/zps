package action

import (
	"errors"
	"strings"
)

type Zpkg struct {
	Uri         string `json:"uri"`
	Arch        string `json:"arch"`
	Os          string `json:"os"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

func NewZpkg() *Zpkg {
	return &Zpkg{}
}

func (z *Zpkg) Key() string {
	return z.Uri
}

func (z *Zpkg) Columns() string {
	return strings.Join(
		[]string{
			z.Uri,
			z.Arch,
		},
		"|",
	)
}

func (z *Zpkg) Unique() string {
	key := []string{"zpkg", z.Uri}
	return strings.Join(key, ":")
}

func (z *Zpkg) Type() string {
	return "zpkg"
}

func (z *Zpkg) Validate() error {
	var err error = nil

	err = z.validateUri()
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

func (z *Zpkg) validateUri() error {
	if z.Uri == "" {
		return errors.New("action zpkg:uri required")
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
