/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package zpm

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fezz-io/zps/action"
	"github.com/fezz-io/zps/phase"
	"github.com/fezz-io/zps/zpkg"
	"github.com/fezz-io/zps/zps"

	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/config"
	"github.com/nightlyone/lockfile"
)

type Manager struct {
	*emission.Emitter
	config *config.ZpsConfig
	state  *State
	cache  *Cache
	lock   lockfile.Lockfile
}

func NewManager(image string) (*Manager, error) {
	var err error
	mgr := &Manager{}

	mgr.Emitter = emission.NewEmitter()

	mgr.config, err = config.LoadConfig(image)
	if err != nil {
		return nil, err
	}

	mgr.lock, err = lockfile.New(filepath.Join(mgr.config.LockPath(), "zpm.lock"))
	if err != nil {
		return nil, err
	}

	mgr.state = NewState(mgr.config.StatePath())
	mgr.cache = NewCache(mgr.config.CachePath())

	return mgr, nil
}

func (m *Manager) CacheClean() error {
	err := m.cache.Clean()
	if err != nil {
		return err
	}

	m.Emit("manager.info", fmt.Sprint("cleaned ", m.cache.path))
	return nil
}

func (m *Manager) CacheClear() error {
	err := m.cache.Clear()
	if err != nil {
		return err
	}

	m.Emit("manager.info", fmt.Sprint("cleared ", m.cache.path))
	return nil
}

func (m *Manager) Channel(repo string, pkg string, channel string) error {
	for _, r := range m.config.Repos {
		if repo == r.Publish.Name && r.Publish.Uri != nil {
			pb := NewPublisher(m.Emitter, m.config.WorkPath(), r.Publish.Uri, r.Publish.Name, r.Publish.Prune)

			err := pb.Channel(pkg, channel)

			return err
		}
	}

	return errors.New("Repo: " + repo + " not found")
}

func (m *Manager) Contents(pkgName string) ([]string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	manifest, err := m.state.Packages.Get(pkgName)
	if err != nil {
		return nil, err
	}

	if manifest == nil {
		return nil, errors.New(fmt.Sprint(pkgName, " not installed"))
	}

	var contents action.Actions
	contents = manifest.Section("Dir", "SymLink", "File")

	sort.Sort(contents)

	var output []string
	for _, fsObject := range contents {
		output = append(output, fsObject.Columns())
	}

	return output, nil
}

func (m *Manager) Freeze(args []string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	pool, err := m.pool()
	if err != nil {
		return err
	}

	for _, arg := range args {
		req, err := zps.NewRequirementFromSimpleString(arg)
		if err != nil {
			return err
		}

		target := pool.Installed(req)
		if target == nil {
			m.Emit("manager.error", fmt.Sprint("Freeze candidate ", arg, " not installed."))
		} else {
			m.state.Frozen.Put(target.Id())
			m.Emit("manager.freeze", fmt.Sprint("froze ", target.Id()))
		}
	}

	return nil
}

func (m *Manager) Info(pkgName string) ([]string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	manifest, err := m.state.Packages.Get(pkgName)
	if err != nil {
		return nil, errors.New(fmt.Sprint(pkgName, " not installed"))
	}

	pkg, err := zps.NewPkgFromManifest(manifest)

	return []string{
		strings.Join([]string{"Name:", pkg.Name()}, "|"),
		strings.Join([]string{"Publisher:", pkg.Publisher()}, "|"),
		strings.Join([]string{"Version:", pkg.Version().Semver.String()}, "|"),
		strings.Join([]string{"Timestamp:", pkg.Version().Timestamp.String()}, "|"),
		strings.Join([]string{"Arch:", pkg.Arch()}, "|"),
		strings.Join([]string{"Summary:", pkg.Summary()}, "|"),
		strings.Join([]string{"Description: ", pkg.Description()}, "|"),
	}, err
}

func (m *Manager) Install(args []string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	reqs, files, err := m.splitReqsFiles(args)

	pool, err := m.pool(files...)
	if err != nil {
		return err
	}

	if pool.RepoCount() <= 1 {
		return errors.New("No repo metadata found. Please run zpm refresh.")
	}

	request := zps.NewRequest()
	for _, arg := range reqs {
		req, err := zps.NewRequirementFromSimpleString(arg)
		if err != nil {
			return err
		}

		if len(pool.WhatProvides(req)) == 0 {
			return errors.New(fmt.Sprint("No candidates found for ", arg))
		}

		request.Install(req)
	}

	// TODO: configure policy
	solver := zps.NewSolver(pool, zps.NewPolicy("updated"))

	solution, err := solver.Solve(request)
	if err != nil {
		return err
	}

	operations, err := solution.Graph()
	if err != nil {
		return err
	}

	for _, op := range operations {
		switch op.Operation {
		case phase.INSTALL:
			uri, _ := url.ParseRequestURI(pool.Location(op.Package.Location()).Uri)
			fe := NewFetcher(uri, m.cache)
			err = fe.Fetch(op.Package.(*zps.Pkg))
			if err != nil {
				return err
			}

			m.Emitter.Emit("manager.fetch", fmt.Sprint("fetching: ", op.Package.Id()))
		case phase.NOOP:
			m.Emit("transaction.noop", fmt.Sprint("using: ", op.Package.Id()))
		}
	}

	if solution.Noop() {
		return nil
	}

	tr := NewTransaction(m.Emitter, m.config.CurrentImage.Path, m.cache, m.state)

	err = tr.Realize(solution)

	return err
}

func (m *Manager) List() ([]string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	pool, err := m.pool()
	if err != nil {
		return nil, err
	}

	var output []string
	for _, pkg := range pool.Image() {

		var line string
		if pool.Frozen(pkg.Id()) {
			line = "[blue]*|" + pkg.(*zps.Pkg).Columns()
		} else {
			line = "[white]~|" + pkg.(*zps.Pkg).Columns()
		}

		output = append(output, line)
	}

	if len(output) == 0 {
		m.Emitter.Emit("manager.warn", "No packages installed.")
		return nil, nil
	}

	return output, nil
}

func (m *Manager) Plan(action string, args []string) (*zps.Solution, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	if action != phase.INSTALL && action != phase.REMOVE {
		return nil, errors.New("action must be either: install or remove")
	}

	pool, err := m.pool()
	if err != nil {
		return nil, err
	}

	if pool.RepoCount() <= 1 {
		return nil, errors.New("No repo metadata found. Please run zpm refresh.")
	}

	request := zps.NewRequest()
	for _, arg := range args {
		req, err := zps.NewRequirementFromSimpleString(arg)
		if err != nil {
			return nil, err
		}

		if len(pool.WhatProvides(req)) == 0 {
			return nil, errors.New(fmt.Sprint("No candidates found for ", arg))
		}

		switch action {
		case phase.INSTALL:
			request.Install(req)
		case phase.REMOVE:
			request.Remove(req)
		}
	}

	// TODO: configure policy
	solver := zps.NewSolver(pool, zps.NewPolicy("updated"))

	solution, err := solver.Solve(request)
	if err != nil {
		return nil, err
	}

	operations, err := solution.Graph()
	if err != nil {
		return nil, err
	}

	for _, op := range operations {
		switch op.Operation {
		case "noop":
			m.Emitter.Emit("transaction.noop", op.Package.Id())
		case "install":
			m.Emitter.Emit("transaction.install", op.Package.Id())
		case "remove":
			m.Emitter.Emit("transaction.remove", op.Package.Id())
		}

	}

	return solution, nil
}

func (m *Manager) Publish(repo string, pkgs ...string) error {
	for _, r := range m.config.Repos {
		if repo == r.Publish.Name && r.Publish.Uri != nil {
			pb := NewPublisher(m.Emitter, m.config.WorkPath(), r.Publish.Uri, r.Publish.Name, r.Publish.Prune)

			err := pb.Publish(pkgs...)

			return err
		}
	}

	return errors.New("Repo: " + repo + " not found")
}

func (m *Manager) Refresh() error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	for _, r := range m.config.Repos {
		fe := NewFetcher(r.Fetch.Uri, m.cache)

		err = fe.Refresh()
		if err == nil {
			m.Emit("manager.refresh", fmt.Sprint("refreshed: ", r.Fetch.Uri.String()))
		} else {
			m.Emit("manager.warn", fmt.Sprint("no metadata: ", r.Fetch.Uri.String()))
		}
	}

	return nil
}

func (m *Manager) Remove(args []string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	pool, err := m.pool()
	if err != nil {
		return err
	}

	request := zps.NewRequest()
	for _, arg := range args {
		req, err := zps.NewRequirementFromSimpleString(arg)
		if err != nil {
			return err
		}

		if pool.Installed(req) == nil {
			return errors.New(fmt.Sprint("No removal candidates found for ", arg))
		}

		request.Remove(req)
	}

	// TODO: configure policy
	solver := zps.NewSolver(pool, zps.NewPolicy("updated"))

	solution, err := solver.Solve(request)
	if err != nil {
		return err
	}

	tr := NewTransaction(m.Emitter, m.config.CurrentImage.Path, m.cache, m.state)

	err = tr.Realize(solution)

	return err
}

func (m *Manager) RepoInit(name string) error {
	for _, repo := range m.config.Repos {
		if name == repo.Publish.Name && repo.Publish.Uri != nil {
			pb := NewPublisher(m.Emitter, m.config.WorkPath(), repo.Publish.Uri, repo.Publish.Name, repo.Publish.Prune)

			return pb.Init()
		}
	}

	return errors.New("Repo: " + name + " not found")
}

func (m *Manager) RepoContents(name string) ([]string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	for _, repo := range m.config.Repos {

		repoConfig, _ := m.repoConfig(repo.Fetch.Uri.String())
		if repoConfig == nil {
			continue
		}

		if name == repoConfig["name"] && repo.Fetch.Uri != nil {
			var contents []string
			osArches := zps.ExpandOsArch(&zps.OsArch{m.config.CurrentImage.Os, m.config.CurrentImage.Arch})

			for _, osarch := range osArches {
				metafile := m.cache.GetMeta(osarch.String(), repo.Fetch.Uri.String())

				repo := &zps.Repo{}

				metadata := NewMetadata(metafile)
				if !metadata.Exists() {
					continue
				}

				meta, err := metadata.All()
				if err != nil {
					return nil, err
				}
				repo.Load(meta)

				for _, pkg := range repo.Solvables() {
					contents = append(contents, strings.Join([]string{pkg.(*zps.Pkg).Name(), pkg.(*zps.Pkg).Id()}, "|"))
				}
			}

			if len(contents) == 0 {
				return nil, errors.New("No repo metadata found. Please run zpm refresh.")
			}

			return contents, err
		}
	}

	return nil, errors.New("Repo: " + name + " not found")
}

func (m *Manager) RepoList() ([]string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	if len(m.config.Repos) == 0 {
		return nil, nil
	}

	var repos []string
	for _, repo := range m.config.Repos {
		name := "<no meta>"
		repoConfig, _ := m.repoConfig(repo.Fetch.Uri.String())

		if repoConfig != nil {
			name = repoConfig["name"]
		}

		repos = append(repos, strings.Join([]string{name, repo.Fetch.Uri.String()}, "|"))
	}

	return repos, nil
}

func (m *Manager) RepoUpdate(name string) error {
	for _, repo := range m.config.Repos {
		if name == repo.Publish.Name && repo.Publish.Uri != nil {
			pb := NewPublisher(m.Emitter, m.config.WorkPath(), repo.Publish.Uri, repo.Publish.Name, repo.Publish.Prune)

			err := pb.Update()

			return err
		}
	}

	return errors.New("Repo: " + name + " not found")
}

func (m *Manager) Thaw(args []string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	pool, err := m.pool()
	if err != nil {
		return err
	}

	for _, arg := range args {
		req, err := zps.NewRequirementFromSimpleString(arg)
		if err != nil {
			return err
		}

		target := pool.Installed(req)
		if target == nil {
			m.Emit("manager.error", fmt.Sprint("Thaw candidate ", arg, " not installed."))
		} else {
			m.state.Frozen.Del(target.Id())
			m.Emit("manager.thaw", fmt.Sprint("thawed: ", target.Id()))
		}
	}

	return nil
}

func (m *Manager) Status(query string) (string, []string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return "", nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	pool, err := m.pool()
	if err != nil {
		return "", nil, err
	}

	var packages []string
	var status string

	req, err := zps.NewRequirementFromSimpleString(query)
	if err != nil {
		return "", nil, err
	}

	status = "Uninstalled"

	for _, pkg := range pool.WhatProvides(req) {
		var line string
		if pool.Frozen(pkg.Id()) && pkg.Location() == 0 {
			status = "Frozen"
			line = "[blue]*|" + pkg.(*zps.Pkg).Columns()
		} else if pkg.Priority() == -1 {
			status = "Installed"
			line = "[yellow]~|" + pkg.(*zps.Pkg).Columns()
		} else {
			line = "[white]-|" + pkg.(*zps.Pkg).Columns()
		}
		packages = append(packages, line)
	}

	if len(packages) == 0 {
		return "", nil, errors.New(fmt.Sprint("pkg: ", query, " unavailable"))
	}

	return status, packages, nil
}

func (m *Manager) TransActionList() ([]string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	transactions, err := m.state.Transactions.All()
	if err != nil {
		return nil, err
	}

	var output []string
	seen := make(map[string]bool)
	for index, t := range transactions {
		if !seen[t.Id] {
			if index != 0 {
				output = append(output, "")
			}
			output = append(output, strings.Join([]string{"[white]" + t.Id, t.Date.Format("Mon Jan 2 15:04:05 MST 2006")}, "|"))
			seen[t.Id] = true
		}

		var op string
		if t.Operation == "install" {
			op = "[green]+ "
		}

		if t.Operation == "remove" {
			op = "[red]- "
		}

		output = append(output, fmt.Sprint(op, t.PkgId))
	}

	if len(transactions) == 0 {
		m.Emitter.Emit("manager.warn", "No transactions found.")
		return nil, nil
	}

	return output, nil
}

func (m *Manager) image() (*zps.Repo, error) {
	packages, err := m.state.Packages.All()
	if err != nil {
		return nil, err
	}

	var solvables zps.Solvables
	for _, manifest := range packages {
		pkg, _ := zps.NewPkgFromManifest(manifest)

		solvables = append(solvables, pkg)
	}

	image := zps.NewRepo("installed", -1, true, nil, solvables)

	return image, nil
}

func (m *Manager) fileRepos(files ...string) ([]*zps.Repo, error) {
	var repos []*zps.Repo
	index := make(map[string]*zps.Repo)

	for _, file := range files {
		path := filepath.Dir(file)

		if index[path] == nil {
			index[path] = zps.NewRepo("local://"+path, 0, true, nil, nil)
			repos = append(repos, index[path])
		}

		reader := zpkg.NewReader(file, "")
		err := reader.Read()
		if err != nil {
			return nil, err
		}

		pkg, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return nil, err
		}

		index[path].Add(pkg)
	}

	return repos, nil
}

func (m *Manager) pool(files ...string) (*zps.Pool, error) {
	var repos []*zps.Repo

	image, err := m.image()
	if err != nil {
		return nil, err
	}

	if len(files) > 0 {
		repos, err = m.fileRepos(files...)
		if err != nil {
			return nil, err
		}
	}

	for _, r := range m.config.Repos {
		if r.Enabled == true {
			if !m.cache.HasMeta(r.Fetch.Uri.String()) {
				m.Emit("manager.warn", fmt.Sprintf("missing metadata: %s", r.Fetch.Uri))
				continue
			}

			osArches := zps.ExpandOsArch(&zps.OsArch{m.config.CurrentImage.Os, m.config.CurrentImage.Arch})

			for _, osarch := range osArches {
				repo := zps.NewRepo(r.Fetch.Uri.String(), r.Priority, r.Enabled, r.Channels, []zps.Solvable{})
				metafile := m.cache.GetMeta(osarch.String(), r.Fetch.Uri.String())

				metadata := NewMetadata(metafile)
				if !metadata.Exists() {
					continue
				}

				meta, err := metadata.All()
				if err != nil && !strings.Contains(err.Error(), "no such file") {
					return nil, err
				}
				repo.Load(meta)

				repos = append(repos, repo)
			}
		}
	}

	frozenEntries, err := m.state.Frozen.All()
	if err != nil {
		return nil, err
	}

	frozen := make(map[string]bool)
	for _, entry := range frozenEntries {
		frozen[entry.PkgId] = true
	}

	pool, err := zps.NewPool(image, frozen, repos...)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (m *Manager) repoConfig(uri string) (map[string]string, error) {
	configPath := m.cache.GetConfig(uri)

	if !m.cache.Exists(filepath.Base(configPath)) {
		return nil, errors.New("No repo metadata found. Please run zpm refresh.")
	}

	configDb := NewConfig(configPath)

	config, err := configDb.All()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (m *Manager) splitReqsFiles(args []string) ([]string, []string, error) {
	var reqs []string
	var files []string

	for _, item := range args {
		if _, err := os.Stat(item); err == nil {
			if filepath.Ext(item) == ".zpkg" {
				// TODO an extra package read
				reader := zpkg.NewReader(item, "")
				err := reader.Read()
				if err != nil {
					return reqs, files, err
				}

				pkg, err := zps.NewPkgFromManifest(reader.Manifest)
				if err != nil {
					return reqs, files, err
				}

				reqs = append(reqs, pkg.Id())
				files = append(files, item)

			}
		} else {
			reqs = append(reqs, item)
		}
	}

	return reqs, files, nil
}
