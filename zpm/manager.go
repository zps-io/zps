package zpm

import (
	"errors"
	"strings"

	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/config"
	"github.com/solvent-io/zps/zpm/fetcher"
	"github.com/solvent-io/zps/zpm/publisher"
)

type Manager struct {
	*emission.Emitter
	config *config.ZpsConfig
}

func NewManager(root string, image string) (*Manager, error) {
	var err error
	mgr := &Manager{}

	mgr.Emitter = emission.NewEmitter()

	mgr.config, err = config.LoadConfig(root, image)
	if err != nil {
		return nil, err
	}

	return mgr, nil
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
		fe := fetcher.Get(r.Fetch.Uri)

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
