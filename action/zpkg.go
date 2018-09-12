package action

import (
	"errors"
	"fmt"
)

type Zpkg struct {
	Name        string `json:"name" hcl:"name,label"`
	Version     string `json:"version" hcl:"version"`
	Publisher   string `json:"publisher" hcl:"publisher"`
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

func (z *Zpkg) Id() string {
	return fmt.Sprint(z.Type(), ".", z.Key())
}

func (z *Zpkg) Type() string {
	return "Zpkg"
}

func (z *Zpkg) IsValid() bool {
	err := z.validate()
	if err != nil {
		return false
	}

	return true
}

func (z *Zpkg) Condition() *bool {
	return nil
}

func (z *Zpkg) MayFail() bool {
	return false
}

func (z *Zpkg) validate() error {
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

func (z *Zpkg) validateName() error {
	if z.Name == "" {
		return errors.New("action zpkg.name required")
	}

	return nil
}

func (z *Zpkg) validatePublisher() error {
	if z.Publisher == "" {
		return errors.New("action zpkg.name required")
	}

	return nil
}

func (z *Zpkg) validateVersion() error {
	if z.Version == "" {
		return errors.New("action zpkg.version required")
	}

	return nil
}

func (z *Zpkg) validateArch() error {
	if z.Arch != "x86_64" && z.Arch != "arm64" {
		return errors.New("action zpkg.arch is unsupported")
	}

	return nil
}

func (z *Zpkg) validateOs() error {
	// If only we could lookup a constant in a map?
	if z.Os != "darwin" && z.Os != "linux" && z.Os != "freebsd" {
		return errors.New("action zpkg.os is unsupported")
	}

	return nil
}

func (z *Zpkg) validateSummary() error {
	if z.Summary == "" {
		return errors.New("action zpkg.summary required")
	}

	return nil
}

func (z *Zpkg) validateDescription() error {
	if z.Description == "" {
		return errors.New("action zpkg.description required")
	}

	return nil
}
