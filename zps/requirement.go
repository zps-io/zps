package zps

import (
	"fmt"
	"strings"
)

type Requirement struct {
	Name      string
	Method    string
	Operation int
	Version   *Version
}

type JsonRequirement struct {
	Name      string `json:"name"`
	Method    string `json:"method"`
	Operation string `json:"operation"`
	Version   string `json:"version,omitempty"`
}

func NewRequirement(name string, version *Version) *Requirement {
	return &Requirement{Name: name, Version: version}
}

func NewRequirementFromJsonRequirement(req *JsonRequirement) (*Requirement, error) {
	newreq := &Requirement{}

	newreq.Name = req.Name
	newreq.Method = req.Method
	newreq.Operation = newreq.OpInt(req.Operation)

	if req.Version != "" {
		version := &Version{}
		err := version.Parse(req.Version)
		if err != nil {
			return nil, err
		}

		newreq.Version = version
	}

	return newreq, nil
}

func NewRequirementFromSimpleString(id string) (*Requirement, error) {
	requirement := &Requirement{}
	requirement.Method = "depends"

	split := strings.Split(id, "@")

	if len(split) < 2 {
		requirement.Name = id
		return requirement.ANY(), nil
	}

	requirement.Name = split[0]

	version := &Version{}
	err := version.Parse(split[1])
	if err != nil {
		return nil, err
	}

	requirement.Version = version

	if requirement.Version.Timestamp.IsZero() {
		return requirement.EQ(), nil
	} else {
		return requirement.EXQ(), nil
	}
}

func (r *Requirement) Depends() *Requirement {
	r.Method = "depends"
	return r
}

func (r *Requirement) Provides() *Requirement {
	r.Method = "provides"
	return r
}

func (r *Requirement) Conflicts() *Requirement {
	r.Method = "conflicts"
	return r
}

func (r *Requirement) ANY() *Requirement {
	r.Operation = 3
	return r
}

func (r *Requirement) GTE() *Requirement {
	r.Operation = 1
	return r
}

func (r *Requirement) LTE() *Requirement {
	r.Operation = -1
	return r
}

func (r *Requirement) EQ() *Requirement {
	r.Operation = 0
	return r
}

func (r *Requirement) EXQ() *Requirement {
	r.Operation = 2
	return r
}

func (r *Requirement) Op(op string) *Requirement {
	switch op {
	case "ANY":
		return r.ANY()
	case "GTE":
		return r.GTE()
	case "LTE":
		return r.LTE()
	case "EQ":
		return r.EQ()
	case "EXQ":
		return r.EXQ()
	}

	return r
}

func (r *Requirement) OpString() string {
	switch r.Operation {
	case 3:
		return "ANY"
	case 1:
		return "GTE"
	case -1:
		return "LTE"
	case 0:
		return "EQ"
	case 2:
		return "EXQ"
	}

	return ""
}

func (r *Requirement) OpInt(op string) int {
	switch op {
	case "ANY":
		return 3
	case "GTE":
		return 1
	case "LTE":
		return -1
	case "EQ":
		return 0
	case "EXQ":
		return 2
	}

	return 3
}

func (r *Requirement) Json() *JsonRequirement {
	json := &JsonRequirement{}

	json.Method = r.Method
	json.Name = r.Name
	json.Version = r.Version.String()
	json.Operation = r.OpString()

	return json
}

func (r *Requirement) String() string {
	switch r.Operation {
	case 3:
		return fmt.Sprint(r.Name, " == ", "*")
	case 2:
		return fmt.Sprint(r.Name, " === ", r.Version.String())
	case 1:
		return fmt.Sprint(r.Name, " >= ", r.Version.Short())
	case 0:
		return fmt.Sprint(r.Name, " == ", r.Version.Short())
	case -1:
		return fmt.Sprint(r.Name, " <= ", r.Version.Short())
	}
	return ""
}
