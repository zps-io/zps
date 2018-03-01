package zps

import (
	"encoding/json"
	"sort"
	"time"
)

type Repo struct {
	Uri      string
	Priority int
	Enabled  bool
	Updated  time.Time

	channels  map[string]bool
	solvables Solvables
	index     map[string]Solvables
}

type Repos []*Repo

type JsonRepo struct {
	Updated  string     `json:"updated"`
	Packages []*JsonPkg `json:"packages"`
}

func NewRepo(uri string, priority int, enabled bool, channels []string, solvables Solvables) *Repo {
	repo := &Repo{Uri: uri, Priority: priority, Enabled: enabled, channels: make(map[string]bool), solvables: solvables}
	repo.Index()

	for _, ch := range channels {
		repo.channels[ch] = true
	}

	return repo
}

func (r *Repo) Index() {
	r.index = make(map[string]Solvables)
	sort.Sort(r.solvables)

	for _, solvable := range r.solvables {
		r.index[solvable.Name()] = append(r.index[solvable.Name()], solvable)
	}
}

func (r *Repo) Add(pkgs ...*Pkg) Solvables {
	var rejects Solvables

	for _, p := range pkgs {
		if !r.Contains(p) {
			r.solvables = append(r.solvables, p)
		} else {
			rejects = append(rejects, p)
		}
	}

	r.Index()

	return rejects
}

func (r *Repo) Remove(pkg *Pkg) error {
	return nil
}

func (r *Repo) Contains(pkg *Pkg) bool {
	if _, ok := r.index[pkg.Name()]; ok {
		for _, candidate := range r.index[pkg.Name()] {
			if candidate.Version().EXQ(pkg.Version()) {
				return true
			}
		}
	}

	return false
}

func (r *Repo) Prune(count int) (Solvables, error) {
	var pruned Solvables
	var result Solvables

	for name := range r.index {
		for len(r.index[name]) > count {
			var prune Solvable
			prune, r.index[name] = r.index[name][len(r.index[name])-1], r.index[name][:len(r.index[name])-1]
			pruned = append(pruned, prune)
		}

		result = append(result, r.index[name]...)
	}

	r.solvables = result

	return pruned, nil
}

func (r *Repo) Load(bytes []byte) error {
	meta := &JsonRepo{}
	err := json.Unmarshal(bytes, meta)

	for _, jpkg := range meta.Packages {
		pkg, err := NewPkgFromJson(jpkg)
		if err != nil {
			return err
		}

		r.solvables = append(r.solvables, pkg)
	}

	r.Updated, err = time.Parse("20060102T150405Z", meta.Updated)
	if err != nil {
		return err
	}

	r.Index()

	return err
}

func (r *Repo) Solvables() Solvables {
	if len(r.channels) > 0 {
		var filtered Solvables

		for index, solvable := range r.solvables {
			for _, ch := range solvable.Channels() {
				if r.channels[ch] {
					filtered = append(filtered, r.solvables[index])
					break
				}
			}
		}

		return filtered
	}

	return r.solvables
}

func (r *Repo) Json() ([]byte, error) {
	jsonMeta := &JsonRepo{}
	jsonMeta.Updated = time.Now().Format("20060102T150405Z")

	for _, pkg := range r.solvables {
		jsonMeta.Packages = append(jsonMeta.Packages, pkg.(*Pkg).Json())
	}

	result, err := json.Marshal(jsonMeta)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (slice Repos) Len() int {
	return len(slice)
}

func (slice Repos) Less(i, j int) bool {
	return slice[i].Priority < slice[j].Priority
}

func (slice Repos) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
