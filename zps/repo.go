package zps

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

type Repo struct {
	Uri       string
	Priority  int
	Enabled   bool
	Solvables Solvables
}

func NewRepo(uri string, priority int, enabled bool, solvables Solvables) *Repo {
	repo := &Repo{uri, priority, enabled, solvables}
	sort.Sort(repo.Solvables)
	return repo
}

type Repos []*Repo

func (r *Repo) Id() string {
	hasher := sha256.New()
	hasher.Write([]byte(r.Uri))

	return hex.EncodeToString(hasher.Sum(nil))
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
