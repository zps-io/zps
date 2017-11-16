package publisher

import (
	"encoding/json"
	"time"

	"sort"

	"fmt"

	"github.com/solvent-io/zps/zps"
)

type RepoMeta struct {
	Updated *time.Time
	Repo    *zps.Repo
}

type JsonRepoMeta struct {
	Updated  string         `json:"updated"`
	Packages []*zps.JsonPkg `json:"packages"`
}

func (r *RepoMeta) Load(bytes []byte) error {
	meta := &JsonRepoMeta{}
	err := json.Unmarshal(bytes, meta)

	r.Repo = &zps.Repo{}

	for _, jpkg := range meta.Packages {
		pkg, err := zps.NewPkgFromJson(jpkg)
		if err != nil {
			return err
		}

		r.Repo.Solvables = append(r.Repo.Solvables, pkg)
	}

	return err
}

func (r *RepoMeta) Add(pkgs ...*zps.Pkg) {
	if r.Repo == nil {
		r.Repo = &zps.Repo{}
	}

	for _, p := range pkgs {
		r.Repo.Solvables = append(r.Repo.Solvables, p)
	}

	sort.Sort(r.Repo.Solvables)
}

func (r *RepoMeta) Prune(count int) ([]string, error) {
	index := make(map[string]zps.Solvables)

	sort.Sort(r.Repo.Solvables)

	for _, solvable := range r.Repo.Solvables {
		index[solvable.Name()] = append(index[solvable.Name()], solvable)
	}

	var pruned zps.Solvables
	var result zps.Solvables

	for name := range index {
		for len(index[name]) > count {
			var prune zps.Solvable
			prune, index[name] = index[name][len(index[name])-1], index[name][:len(index[name])-1]
			pruned = append(pruned, prune)
		}

		result = append(result, index[name]...)
	}

	var files []string
	for _, file := range pruned {
		files = append(files, fmt.Sprintf("%s@%s-%s-%s.zpkg", file.Name(), file.Version().String(), file.Os(), file.Arch()))
	}

	r.Repo.Solvables = result

	return files, nil
}

func (r *RepoMeta) Remove(pkg zps.ZpkgUri) error {
	return nil
}

func (r *RepoMeta) Json() ([]byte, error) {
	jsonMeta := &JsonRepoMeta{}
	jsonMeta.Updated = time.Now().Format("20060102T150405Z")

	for _, pkg := range r.Repo.Solvables {
		jsonMeta.Packages = append(jsonMeta.Packages, pkg.(*zps.Pkg).Json())
	}

	result, err := json.Marshal(jsonMeta)
	if err != nil {
		return nil, err
	}

	return result, nil
}
