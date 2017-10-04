package publisher

import (
	"encoding/json"
	"time"

	"sort"

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
	return nil, nil
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
