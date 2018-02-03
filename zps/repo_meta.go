package zps

import (
	"encoding/json"
	"time"

	"sort"
)

type RepoMeta struct {
	Updated *time.Time
	Repo    *Repo
}

type JsonRepoMeta struct {
	Updated  string     `json:"updated"`
	Packages []*JsonPkg `json:"packages"`
}

func (r *RepoMeta) Load(bytes []byte) error {
	meta := &JsonRepoMeta{}
	err := json.Unmarshal(bytes, meta)

	r.Repo = &Repo{}

	for _, jpkg := range meta.Packages {
		pkg, err := NewPkgFromJson(jpkg)
		if err != nil {
			return err
		}

		r.Repo.Solvables = append(r.Repo.Solvables, pkg)
	}

	return err
}

func (r *RepoMeta) Add(pkgs ...*Pkg) {
	if r.Repo == nil {
		r.Repo = &Repo{}
	}

	for _, p := range pkgs {
		r.Repo.Solvables = append(r.Repo.Solvables, p)
	}

	sort.Sort(r.Repo.Solvables)
}

func (r *RepoMeta) Prune(count int) ([]string, error) {
	index := make(map[string]Solvables)

	sort.Sort(r.Repo.Solvables)

	for _, solvable := range r.Repo.Solvables {
		index[solvable.Name()] = append(index[solvable.Name()], solvable)
	}

	var pruned Solvables
	var result Solvables

	for name := range index {
		for len(index[name]) > count {
			var prune Solvable
			prune, index[name] = index[name][len(index[name])-1], index[name][:len(index[name])-1]
			pruned = append(pruned, prune)
		}

		result = append(result, index[name]...)
	}

	var files []string
	for _, file := range pruned {
		files = append(files, file.(*Pkg).FileName())
	}

	r.Repo.Solvables = result

	return files, nil
}

func (r *RepoMeta) Remove(pkg ZpkgUri) error {
	return nil
}

func (r *RepoMeta) Json() ([]byte, error) {
	jsonMeta := &JsonRepoMeta{}
	jsonMeta.Updated = time.Now().Format("20060102T150405Z")

	for _, pkg := range r.Repo.Solvables {
		jsonMeta.Packages = append(jsonMeta.Packages, pkg.(*Pkg).Json())
	}

	result, err := json.Marshal(jsonMeta)
	if err != nil {
		return nil, err
	}

	return result, nil
}
