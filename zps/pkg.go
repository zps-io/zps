package zps

import (
	"strings"

	"github.com/solvent-io/zps/action"
)

type Pkg struct {
	uri *ZpkgUri

	reqs []*Requirement

	arch        string
	os          string
	summary     string
	description string

	location int
	priority int
}

type JsonPkg struct {
	Uri          string             `json:"uri"`
	Requirements []*JsonRequirement `json:"requirements,omitempty"`

	Arch        string `json:"arch"`
	Os          string `json:"os"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

func NewPkg(uri string, reqs []*Requirement, arch string, os string, summary string, description string) (*Pkg, error) {
	u := NewZpkgUri()
	err := u.Parse(uri)
	if err != nil {
		return nil, err
	}
	return &Pkg{u, reqs, arch, os, summary, description, 0, 0}, nil
}

func NewPkgFromJson(jpkg *JsonPkg) (*Pkg, error) {
	pkg := &Pkg{}

	uri := NewZpkgUri()
	err := uri.Parse(jpkg.Uri)
	if err != nil {
		return nil, err
	}

	pkg.uri = uri
	pkg.arch = jpkg.Arch
	pkg.os = jpkg.Os
	pkg.summary = jpkg.Summary
	pkg.description = jpkg.Description

	for _, jreq := range jpkg.Requirements {
		req, err := NewRequirementFromJson(jreq)
		if err != nil {
			return nil, err
		}

		pkg.reqs = append(pkg.reqs, req)
	}

	return pkg, nil
}

func NewPkgFromManifest(manifest *action.Manifest) (*Pkg, error) {
	pkg := &Pkg{}
	zpkg := manifest.Section("zpkg")[0].(*action.Zpkg)

	uri := NewZpkgUri()
	err := uri.Parse(zpkg.Uri)
	if err != nil {
		return nil, err
	}

	pkg.uri = uri
	pkg.arch = zpkg.Arch
	pkg.os = zpkg.Os
	pkg.summary = zpkg.Summary
	pkg.description = zpkg.Description

	for _, raction := range manifest.Section("requirement") {
		req := NewRequirement(raction.(*action.Requirement).Name, nil)

		if raction.(*action.Requirement).Version != "" {
			version := &Version{}
			err := version.Parse(raction.(*action.Requirement).Version)
			if err != nil {
				return nil, err
			}

			req.Version = version
		}

		switch raction.(*action.Requirement).Method {
		case "depends":
			req = req.Depends().Op(raction.(*action.Requirement).Operation)
		case "conflicts":
			req = req.Conflicts().Op(raction.(*action.Requirement).Operation)
		case "provides":
			req = req.Provides().ANY()
		}

		pkg.reqs = append(pkg.reqs, req)
	}

	return pkg, nil
}

func (p *Pkg) Id() string {
	return strings.Join([]string{p.uri.Name, p.uri.Version.String()}, "@")
}

func (p *Pkg) Uri() *ZpkgUri {
	return p.uri
}

func (p *Pkg) Name() string {
	return p.uri.Name
}

func (p *Pkg) Summary() string {
	return p.summary
}

func (p *Pkg) Description() string {
	return p.description
}

func (p *Pkg) Version() *Version {
	return p.uri.Version
}

func (p *Pkg) Requirements() []*Requirement {
	return p.reqs
}

func (p *Pkg) Arch() string {
	return p.arch
}

func (p *Pkg) Os() string {
	return p.os
}

func (p *Pkg) Location() int {
	return p.location
}

func (p *Pkg) SetLocation(location int) {
	p.location = location
}

func (p *Pkg) Priority() int {
	return p.priority
}

func (p *Pkg) SetPriority(priority int) {
	p.priority = priority
}

func (p *Pkg) Satisfies(req *Requirement) bool {
	switch req.Operation {
	case 3:
		return true
	case 2:
		return p.uri.Version.EXQ(req.Version)
	case 1:
		return p.uri.Version.GTE(req.Version)
	case 0:
		return p.uri.Version.EQ(req.Version)
	case -1:
		return p.uri.Version.LTE(req.Version)
	}

	return false
}

func (p *Pkg) Json() *JsonPkg {
	json := &JsonPkg{}
	json.Arch = p.arch
	json.Os = p.os
	json.Description = p.description
	json.Summary = p.summary
	json.Uri = p.uri.String()

	for index := range p.reqs {
		json.Requirements = append(json.Requirements, p.reqs[index].Json())
	}

	return json
}

func (p *Pkg) Columns() string {
	return strings.Join([]string{
		p.Name(),
		p.Summary(),
		p.Uri().String(),
	}, "|")
}
