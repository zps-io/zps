package zpm

import (
	"errors"
	"strings"

	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/config"
	"github.com/solvent-io/zps/zpm/fetcher"
	"github.com/solvent-io/zps/zpm/publisher"
	"github.com/solvent-io/zps/zps"
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"fmt"
	"io/ioutil"
	"os"
)

type Manager struct {
	*emission.Emitter
	config *config.ZpsConfig
	db *Db
}

func NewManager(root string, image string) (*Manager, error) {
	var err error
	mgr := &Manager{}

	mgr.Emitter = emission.NewEmitter()

	mgr.config, err = config.LoadConfig(root, image)
	if err != nil {
		return nil, err
	}

	mgr.db = &Db{mgr.config.DbPath()}

	return mgr, nil
}

// TODO Entire function is a giant WIP
func (m *Manager) Plan(action string, args []string) (*zps.Solution, error) {
	if action != "install" && action != "remove" {
		return nil, errors.New("action must be either: install or remove")
	}

	var repos []*zps.Repo

	// TODO load installed from current image
	image := zps.NewRepo("installed", -1, true, []zps.Solvable{})

	for _, r := range m.config.Repos {
		if r.Enabled == true {
			repo := zps.NewRepo(r.Fetch.Uri.String(), r.Priority, r.Enabled, []zps.Solvable{})

			// Load meta from cache
			hasher := sha256.New()
			hasher.Write([]byte(r.Fetch.Uri.String()))
			repoId := hex.EncodeToString(hasher.Sum(nil))

			// TODO fix
			osarch := &zps.OsArch{m.config.CurrentImage.Os, m.config.CurrentImage.Arch}

			packagesfile := filepath.Join(m.config.CachePath(), fmt.Sprint(repoId, ".", osarch.String(), ".packages.json"))
			meta := &zps.RepoMeta{}

			pkgsbytes, err := ioutil.ReadFile(packagesfile)

			if err == nil {
				err = meta.Load(pkgsbytes)
				if err != nil {
					return nil, err
				}
			} else if !os.IsNotExist(err) {
				return nil, err
			}

			for _, pkg := range meta.Repo.Solvables {
				repo.Solvables = append(repo.Solvables, pkg)
			}

			repos = append(repos, repo)
		}
	}

	pool, err := zps.NewPool(image, repos...)
	if err != nil {
		return nil, err
	}

	request := zps.NewRequest()
	for _, arg := range args {
		req, err := zps.NewRequirementFromSimpleString(arg)
		if err != nil {
			return nil, err
		}

		switch action {
		case "install":
			request.Install(req)
		case "remove":
			request.Remove(req)
		}
	}

	// TODO: configure policy
	solver := zps.NewSolver(pool, zps.NewPolicy("updated"))

	solution, err := solver.Solve(request)
	if err != nil {
		return nil, err
	}

	for _, op := range solution.Operations() {
		switch op.Operation {
		case "install":
			m.Emitter.Emit("install", op.Package.Id())
		case "remove":
			m.Emitter.Emit("remove", op.Package.Id())
		}

	}

	return solution, nil
}

func (m *Manager) Publish(repo string, pkgs ...string) error {
	for _, r := range m.config.Repos {
		if repo == r.Name && r.Publish.Uri != nil {
			pb := publisher.Get(r.Publish.Uri, r.Publish.Prune)

			err := pb.Publish(pkgs...)

			return err
		}
	}

	return errors.New("Repo: " + repo + " not found")
}

func (m *Manager) Refresh() error {
	for _, r := range m.config.Repos {
		fe := fetcher.Get(r.Fetch.Uri, m.config.CachePath())

		err := fe.Refresh()

		return err
	}

	return nil
}

func (m *Manager) RepoList() ([]string, error) {
	if len(m.config.Repos) == 0 {
		return nil, nil
	}

	var repos []string
	for _, repo := range m.config.Repos {
		repos = append(repos, strings.Join([]string{repo.Name, repo.Fetch.Uri.String()}, "|"))
	}

	return repos, nil
}

func (m *Manager) RepoInit(name string) error {
	for _, repo := range m.config.Repos {
		if name == repo.Name && repo.Publish.Uri != nil {
			pb := publisher.Get(repo.Publish.Uri, repo.Publish.Prune)

			err := pb.Init()

			return err
		}
	}

	return errors.New("Repo: " + name + " not found")
}
