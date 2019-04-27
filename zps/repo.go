/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zps

import (
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

func (r *Repo) Load(pkgs []*Pkg)  {

	for _, pkg := range pkgs {
		r.solvables = append(r.solvables, pkg)
	}

	r.Index()
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

func (slice Repos) Len() int {
	return len(slice)
}

func (slice Repos) Less(i, j int) bool {
	return slice[i].Priority < slice[j].Priority
}

func (slice Repos) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
