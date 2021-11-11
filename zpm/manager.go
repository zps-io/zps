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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fezz-io/zps/provider"
	"github.com/fezz-io/zps/sec"

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

	state *State
	cache *Cache
	pki   *Pki

	security Security

	lock lockfile.Lockfile
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
	mgr.pki = NewPki(mgr.config.PkiPath())

	mgr.security, err = NewSecurity(mgr.config.Security, mgr.pki)
	if err != nil {
		return nil, err
	}

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
		if r.Publish == nil {
			continue
		}

		if repo == r.Publish.Name && r.Publish.Uri != nil {
			pb := NewPublisher(m.Emitter, m.security, m.config.WorkPath(), r.Publish.Uri, r.Publish.Name, r.Publish.Prune, r.Publish.LockUri)

			err := pb.Channel(pkg, channel)

			return err
		}
	}

	return errors.New("Repo: " + repo + " not found")
}

func (m *Manager) Configure(packages []string, profile string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	pool, err := m.pool()
	if err != nil {
		return err
	}

	options := &provider.Options{TargetPath: m.config.CurrentImage.Path}
	ctx := m.getContext(phase.CONFIGURE, options)
	ctx = context.WithValue(ctx, "hclCtx", m.config.HclContext(profile))

	factory := provider.DefaultFactory(m.Emitter)

	if len(packages) == 0 {
		installed, err := m.state.Packages.All()
		if err != nil {
			return err
		}

		for _, pkg := range installed {
			packages = append(packages, pkg.Zpkg.Name)
		}
	}

	for _, pkg := range packages {
		req, err := zps.NewRequirementFromSimpleString(pkg)
		if err != nil {
			return err
		}

		if pool.Installed(req) != nil {
			tpls, err := m.state.Templates.Get(pkg)
			if err != nil {
				return err
			}

			// Merge templates
			tpls = MergeTemplateConfig(tpls, m.config.TemplatesForPkg(pkg))

			// Run package tpls first
			for _, tpl := range tpls {
				err = factory.Get(tpl).Realize(ctx)
				if err != nil {
					m.Emit("manager.error", fmt.Sprintf("Template %s failed: %s", tpl.Key(), err.Error()))
				}
			}

			// Run services last
			current, err := m.state.Packages.Get(pkg)
			if err != nil {
				return err
			}

			for _, svc := range current.Section("Service") {
				err = factory.Get(svc).Realize(ctx)
				if err != nil {
					m.Emit("manager.error", fmt.Sprintf("Service %s failed: %s", svc.Key(), err.Error()))
				}
			}
		}
	}

	return nil
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

func (m *Manager) Fetch(args []string) error {
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
	policy := zps.NewPolicy("updated")

	for _, job := range request.Jobs() {
		pkg := policy.SelectRequest(pool.WhatProvides(job.Requirement()))

		uri, _ := url.ParseRequestURI(pool.Location(pkg.Location()).Uri)
		fe := NewFetcher(uri, m.cache, m.security, m.config.CloudProvider())
		err = fe.Fetch(pkg.(*zps.Pkg))
		if err != nil {
			return err
		}

		m.Emitter.Emit("manager.fetch", fmt.Sprint("fetching: ", pkg.Id()))

		// Copy from cache to working directory
		wd, err := os.Getwd()
		if err != nil {
			return errors.New("could not get current directory")
		}

		src, err := os.Open(m.cache.GetFile(pkg.FileName()))
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.OpenFile(filepath.Join(wd, pkg.FileName()), os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
	}

	return err
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

// TODO this is trash, refactor it after Imagefile support is added
func (m *Manager) ImageInit(imageFilePath string, name string, imageOs string, arch string, imagePath string, profile string, configure bool, force bool, helper bool) error {
	image := &config.ImageFile{}

	err := image.Load(imageFilePath)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	} else {
		m.Emit("manager.info", fmt.Sprintf("using Imagefile from: %s", image.FilePath))
	}

	// Flags override ImageFile content
	if name != "" {
		image.Name = name
	}

	if image.Name == "" {
		return errors.New("no name found or specified for image")
	}

	if imageOs != "" {
		image.Os = imageOs
	}

	if arch != "" {
		image.Arch = arch
	}

	// TODO move this into OsArch
	switch image.Os {
	case "":
	case "darwin", "linux":
		m.config.CurrentImage.Os = image.Os
	default:
		return fmt.Errorf("unsupported os %s", image.Os)
	}

	switch arch {
	case "":
	case "x86_64":
		m.config.CurrentImage.Os = image.Arch
	default:
		return fmt.Errorf("unsupported arch %s", image.Arch)
	}

	// Attempt to detect images path
	imagesPath := os.Getenv("ZPS_IMAGES_PATH")
	if imagesPath == "" {
		prefix, err := config.InstallPrefix()
		if err != nil {
			return errors.New("could not detect images path")
		}

		imagesPath = path.Dir(prefix)
	}

	// Override if flag is passed
	if imagePath != "" {
		image.Path, err = filepath.Abs(imagePath)
		if err != nil {
			return err
		}
	}

	// If the image path is not empty ensure abs path is set, otherwise detect the path
	if image.Path != "" {
		image.Path, err = filepath.Abs(image.Path)
		if err != nil {
			return err
		}
	} else {
		// If the image.Path is still empty construct it from either ZPS_IMAGES_PATH or one dir up from the current
		// binary
		image.Path, err = filepath.Abs(filepath.Join(imagesPath, image.Name))
		if err != nil {
			return err
		}
	}

	// Look for name conflicts
	for _, img := range m.config.Images {
		if img.Name == image.Name && img.Path != image.Path {
			return fmt.Errorf("name %s conflicts with existing image: %s", image.Name, img.Path)
		}
	}

	// Refresh metadata in current image to ensure latest zps package
	err = m.Refresh()
	if err != nil {
		return err
	}

	// Silently try dir create
	os.MkdirAll(image.Path, 0755)

	// Exit if the image path is not empty
	empty, err := m.IsEmptyImage(image.Path)
	if err != nil {
		return err
	}

	if !empty && !force {
		return fmt.Errorf("selected init path is not empty: %s", image.Path)
	}

	if force {
		m.Emit("manager.warn", fmt.Sprintf("purging image path: %s", image.Path))
		err = m.EmptyImage(image.Path)
		if err != nil {
			return err
		}
	}

	// Start bootstrap
	m.Emit("manager.info", fmt.Sprintf("initializing: %s", image.Path))

	userImagePath := filepath.Join(m.config.UserPath(), config.ImagePath)
	defaultImagePath := filepath.Join(m.config.ConfigPath(), config.ImagePath)

	m.config.CurrentImage.Path = image.Path

	// Create state db
	os.MkdirAll(m.config.StatePath(), 0755)
	os.Chmod(m.config.StatePath(), 0750)
	m.state = NewState(m.config.StatePath())

	// Install zps into changed root
	err = m.Install([]string{"zps"}, nil)
	if err != nil {
		return err
	}

	// Create pki db and import zps trust
	m.pki = NewPki(m.config.PkiPath())

	// Point to new security
	m.security, err = NewSecurity(m.config.Security, m.pki)
	if err != nil {
		return err
	}

	err = m.PkiTrustImport(filepath.Join(m.config.CertPath(), "zps.io", "ca.pem"), "ca")
	if err != nil {
		return err
	}

	err = m.PkiTrustImport(filepath.Join(m.config.CertPath(), "zps.io", "intermediate.pem"), "intermediate")
	if err != nil {
		return err
	}

	err = m.PkiTrustImport(filepath.Join(m.config.CertPath(), "zps.io", "zps.pem"), "user")
	if err != nil {
		return err
	}

	// Import imagefile trusts
	for _, trust := range image.Trusts {
		err := m.PkiTrustFetch(trust.Uri)
		if err != nil {
			m.Emit("manager.error", fmt.Sprintf("failed to fetch certificates for: %s", trust.Publisher))
			continue
		}
	}

	// TODO fix this so we can refresh the cert cache following imports
	m.security, err = NewSecurity(m.config.Security, m.pki)
	if err != nil {
		return err
	}

	// Ensure we have metadata for the new image
	m.cache = NewCache(m.config.CachePath())

	// Write out repo configs from the ImageFile if present
	for _, rc := range image.Repos {
		ioutil.WriteFile(filepath.Join(m.config.ConfigPath(), config.RepoPath, rc.Name+".conf"), rc.ToHclFile().Bytes(), 0640)
	}

	// Reload repos
	err = m.config.LoadRepos()
	if err != nil {
		return err
	}

	err = m.Refresh()
	if err != nil {
		return err
	}

	// Install image packages
	// TODO make all qualifiers work
	request := zps.NewRequest()
	for _, pkg := range image.Packages {
		version := &zps.Version{}
		req := zps.NewRequirement(pkg.Name, version)

		if pkg.Version != "" {
			err = version.Parse(pkg.Version)
			if err != nil {
				m.Emit("manager.warn", fmt.Sprintf("could not parse version for package: %s, skipping", pkg.Name))
			}

			if !version.Timestamp.IsZero() {
				req.EXQ()
			} else {
				req.EQ()
			}
		} else {
			req.ANY()
		}

		request.Install(req)
	}

	if len(request.Jobs()) > 0 {
		err = m.Install(nil, request)
		if err != nil {
			return err
		}
	}

	// Install templates
	for _, tpl := range image.Templates {
		ioutil.WriteFile(filepath.Join(m.config.ConfigPath(), config.TplPath, tpl.Name+".conf"), tpl.ToHclFile().Bytes(), 0640)
		m.Emit("manager.info", fmt.Sprintf("template %s modifies: %s", tpl.Name, tpl.Output))
	}

	// Install config profiles
	for _, cfg := range image.Configs {
		ioutil.WriteFile(filepath.Join(m.config.ConfigPath(), config.CfgPath, cfg.Namespace+".conf"), cfg.ToHclFile().Bytes(), 0640)
		m.Emit("manager.info", fmt.Sprintf("config namespace add: %s", cfg.Namespace))
	}

	// Write config
	conf := &config.ImageConfig{
		Name: image.Name,
		Path: m.config.CurrentImage.Path,
		Os:   m.config.CurrentImage.Os,
		Arch: m.config.CurrentImage.Arch,
	}

	if _, err := os.Stat(userImagePath); !os.IsNotExist(err) {
		ioutil.WriteFile(filepath.Join(userImagePath, image.Name+".conf"), conf.ToHclFile().Bytes(), 0640)
	} else {
		ioutil.WriteFile(filepath.Join(defaultImagePath, image.Name+".conf"), conf.ToHclFile().Bytes(), 0640)
	}

	// Rewrite helper if required
	err = m.config.SetupHelper(helper)
	if err != nil {
		return err
	}

	if configure {
		err = m.Configure(nil, profile)
		if err != nil {
			return err
		}
	} else {
		m.Emit("manager.warn", "skipping configure")
	}

	return nil
}

func (m *Manager) ImageCurrent(image string) error {
	if image == "" {
		m.Emit("manager.out", fmt.Sprintf("%s %s", m.config.CurrentImage.Name, m.config.CurrentImage.Path))

		return nil
	}

	err := m.config.SelectImage(image)
	if err != nil {
		return err
	}

	// Write new image path to ZPS Session
	session := os.Getenv("ZPS_SESSION")

	if !strings.Contains(session, "zps.sess") {
		return errors.New("zps session not found, ensure the shell helper is loaded")
	}

	err = ioutil.WriteFile(session, []byte(m.config.CurrentImage.Path), 0600)
	if err != nil {
		return err
	}

	m.Emit("manager.info", fmt.Sprintf("%s %s", m.config.CurrentImage.Name, m.config.CurrentImage.Path))

	return nil
}

func (m *Manager) ImageDelete(image string) error {
	if len(m.config.Images) <= 2 {
		return errors.New("you need at least one zps image left to default to")
	}

	// Delete passed image and config
	for _, img := range m.config.Images {
		if (img.Name == image || img.Path == image) && img.Name != "zroot" && img.Name != "default" {
			// Basic safety check
			if !m.IsImage(img.Path) {
				return fmt.Errorf("path does not appear to be a zps image: %s", img.Path)
			}

			os.RemoveAll(img.Path)

			// Remove the image config in question
			cfgPath := m.config.ConfigForImage(img.Path)
			if cfgPath != "" {
				os.Remove(cfgPath)
			}

			m.Emit("manager.warn", fmt.Sprintf("removed: %s : %s", img.Name, img.Path))

			// If we just deleted change the current image to default
			if img.Name == m.config.CurrentImage.Name || img.Path == m.config.CurrentImage.Path {
				err := m.config.LoadImages()
				if err != nil {
					return err
				}

				err = m.ImageCurrent(m.config.Images[1].Name)
				if err != nil {
					return err
				}
			}
			break
		}
	}

	return nil
}

func (m *Manager) ImageList() error {
	for _, image := range m.config.Images {
		if image == m.config.CurrentImage {
			m.Emit("manager.out", fmt.Sprintf("* %s %s", image.Name, image.Path))
		} else if image.Name == "zroot" {
			m.Emit("manager.out", fmt.Sprintf("  [yellow](%s) %s", image.Name, image.Path))
		} else {
			m.Emit("manager.out", fmt.Sprintf("  %s %s", image.Name, image.Path))
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
		return nil, err
	}

	if manifest == nil {
		return nil, errors.New(fmt.Sprint(pkgName, " not installed"))
	}

	pkg, err := zps.NewPkgFromManifest(manifest)
	if err != nil {
		return nil, err
	}

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

func (m *Manager) Install(args []string, request *zps.Request) error {
	/*
		f, err := os.Create("zps.pprof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/
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

	if request == nil {
		request = zps.NewRequest()
		for _, arg := range reqs {
			req, err := zps.NewRequirementFromSimpleString(arg)
			if err != nil {
				return err
			}

			if len(pool.WhatProvides(req)) == 0 {
				return errors.New(fmt.Sprint("no candidates found for: ", arg))
			}

			request.Install(req)
		}
	} else {
		for _, job := range request.Jobs() {
			if len(pool.WhatProvides(job.Requirement())) == 0 {
				return errors.New(fmt.Sprint("no candidates found for: ", job.Requirement().Name))
			}
		}
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
			fe := NewFetcher(uri, m.cache, m.security, m.config.CloudProvider())

			m.Emitter.Emit("spin.start", fmt.Sprint("fetching: ", op.Package.Id()))
			err = fe.Fetch(op.Package.(*zps.Pkg))
			if err != nil {
				m.Emitter.Emit("spin.error", fmt.Sprint("failed: ", op.Package.Id()))
				return err
			}

			m.Emitter.Emit("spin.success", fmt.Sprint("fetched: ", op.Package.Id()))
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

func (m *Manager) PkiKeyPairImport(certPath string, keyPath string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	err = sec.SecurityValidateKeyPair(certPath, keyPath)
	if err != nil {
		return err
	}

	certPem, err := ioutil.ReadFile(certPath)
	if err != nil {
		return err
	}

	keyPem, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}

	subject, publisher, fingerprint, err := sec.SecurityCertMetaFromBytes(&certPem)
	if err != nil {
		return err
	}

	kps, err := m.pki.KeyPairs.GetByPublisher(publisher)
	if err != nil {
		return err
	}

	if len(kps) > 0 {
		for index := range kps {
			m.Emitter.Emit("manager.warn",
				fmt.Sprintf("Removing %s due to matching publisher %s",
					kps[index].Fingerprint,
					publisher,
				),
			)

			err := m.pki.KeyPairs.Del(kps[index].Fingerprint)
			if err != nil {
				return err
			}
		}
	}

	err = m.pki.KeyPairs.Put(fingerprint, subject, publisher, certPem, keyPem)
	if err != nil {
		return err
	}

	m.Emit("manager.info", fmt.Sprintf("Imported keypair for publisher %s", publisher))

	return nil
}

func (m *Manager) PkiKeyPairList() ([]string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	kps, err := m.pki.KeyPairs.All()
	if err != nil {
		return nil, err
	}

	var output []string

	for _, entry := range kps {
		output = append(output, strings.Join([]string{
			entry.Subject,
			entry.Publisher,
			entry.Fingerprint,
		}, "|"))
	}

	return output, nil
}

func (m *Manager) PkiKeyPairRemove(fingerprint string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	kp, err := m.pki.KeyPairs.Get(fingerprint)
	if err != nil {
		return err
	}

	err = m.pki.KeyPairs.Del(fingerprint)
	if err != nil {
		return err
	}

	m.Emit("manager.info", fmt.Sprintf("removed keypair: %s", kp.Subject))

	return err
}

func (m *Manager) PkiTrustFetch(uriString string) error {
	uri, err := url.Parse(uriString)
	if err != nil {
		return fmt.Errorf("invalid uri format")
	}

	fe := NewFetcher(uri, m.cache, m.security, m.config.CloudProvider())
	if fe == nil {
		return fmt.Errorf("uri path not found: %s", uri)
	}

	results, err := fe.Keys()
	if err != nil {
		return err
	}

	for _, cert := range results {
		m.Emit("manager.info", fmt.Sprintf("Imported certificate '%s' for publisher: %s", cert[0], cert[1]))
	}

	return err
}

func (m *Manager) PkiTrustImport(certPath string, typ string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	certPem, err := ioutil.ReadFile(certPath)
	if err != nil {
		return err
	}

	subject, publisher, err := m.security.Trust(&certPem, typ)
	if err != nil {
		return err
	}

	m.Emit("manager.info", fmt.Sprintf("Imported certificate '%s' for publisher: %s", subject, publisher))

	return nil
}

func (m *Manager) PkiTrustList() ([]string, error) {
	err := m.lock.TryLock()
	if err != nil {
		return nil, errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	kps, err := m.pki.Certificates.All()
	if err != nil {
		return nil, err
	}

	var output []string

	for _, entry := range kps {
		output = append(output, strings.Join([]string{
			entry.Subject,
			entry.Publisher,
			entry.Type,
			entry.Fingerprint,
		}, "|"))
	}

	return output, nil
}

func (m *Manager) PkiTrustRemove(fingerprint string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	cert, err := m.pki.Certificates.Get(fingerprint)
	if err != nil {
		return err
	}

	err = m.pki.Certificates.Del(fingerprint)
	if err != nil {
		return err
	}

	m.Emit("manager.info", fmt.Sprintf("removed certficate: %s", cert.Subject))

	return err
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
		if r.Publish == nil {
			continue
		}

		if repo == r.Publish.Name && r.Publish.Uri != nil {
			pb := NewPublisher(m.Emitter, m.security, m.config.WorkPath(), r.Publish.Uri, r.Publish.Name, r.Publish.Prune, r.Publish.LockUri)

			return pb.Publish(pkgs...)
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
		if r.Enabled == false {
			m.Emit("manager.warn", fmt.Sprint("skipped disabled: ", SafeURI(r.Fetch.Uri)))
			continue
		}

		fe := NewFetcher(r.Fetch.Uri, m.cache, m.security, m.config.CloudProvider())
		m.Emit("spin.start", fmt.Sprint("refreshing: ", SafeURI(r.Fetch.Uri)))
		err = fe.Refresh()
		if err == nil {
			m.Emit("spin.success", fmt.Sprint("refreshed: ", SafeURI(r.Fetch.Uri)))
		} else if strings.Contains(err.Error(), "no trusted certificates") {
			m.Emit("spin.error", fmt.Sprint("metadata validation failed: ", SafeURI(r.Fetch.Uri)))
		} else if strings.Contains(err.Error(), "refresh failed") {
			m.Emit("spin.error", fmt.Sprint("fetch metadata failed: ", SafeURI(r.Fetch.Uri)))
		} else {
			m.Emit("spin.warn", fmt.Sprint("no metadata: ", SafeURI(r.Fetch.Uri)))
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
		if repo.Publish == nil {
			continue
		}

		if name == repo.Publish.Name && repo.Publish.Uri != nil {
			pb := NewPublisher(m.Emitter, m.security, m.config.WorkPath(), repo.Publish.Uri, repo.Publish.Name, repo.Publish.Prune, repo.Publish.LockUri)

			return pb.Init()
		}
	}

	return errors.New("Repo: " + name + " not found")
}

func (m *Manager) RepoUnlock(name string, removeEtag bool) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	for _, repo := range m.config.Repos {
		if repo.Publish == nil {
			continue
		}
		if name == repo.Publish.Name && repo.Publish.LockUri != nil {
			locker := NewLocker(repo.Publish.LockUri)
			if removeEtag {
				emptyEtag := ""
				return locker.UnlockWithEtag(&emptyEtag)
			}
			return locker.Unlock()
		}
	}

	return nil

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

				zrepo := &zps.Repo{}

				metadata := NewMetadata(metafile)
				if !metadata.Exists() {
					continue
				}

				// Validate metadata signature
				if m.security.Mode() != SecurityModeNone {
					err := ValidateFileSignature(m.security, m.cache.GetMeta(osarch.String(), repo.Fetch.Uri.String()), m.cache.GetMetaSig(osarch.String(), repo.Fetch.Uri.String()))
					if err != nil {
						m.Emit("manager.error", fmt.Sprintf("invalid metadata signature: %s", repo.Fetch.Uri))
						continue
					}
				}

				meta, err := metadata.All()
				if err != nil {
					return nil, err
				}
				zrepo.Load(meta)

				for _, pkg := range zrepo.Solvables() {
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
		if repo.Publish == nil {
			continue
		}

		if name == repo.Publish.Name && repo.Publish.Uri != nil {
			pb := NewPublisher(m.Emitter, m.security, m.config.WorkPath(), repo.Publish.Uri, repo.Publish.Name, repo.Publish.Prune, repo.Publish.LockUri)

			return pb.Update()
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

func (m *Manager) Tpl(tplPath string, profile string) error {
	options := &provider.Options{TargetPath: ""}

	ctx := m.getContext(phase.CONFIGURE, options)
	ctx = context.WithValue(ctx, "hclCtx", m.config.HclContext(profile))

	factory := provider.DefaultFactory(m.Emitter)

	tpl := &action.Template{
		Name:   "",
		Source: tplPath,
		Output: "",
		Owner:  "",
		Group:  "",
		Mode:   "",
	}

	return factory.Get(tpl).Realize(ctx)
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

func (m *Manager) Update(reqs []string) error {
	err := m.lock.TryLock()
	if err != nil {
		return errors.New("zpm: locked by another process")
	}
	defer m.lock.Unlock()

	pool, err := m.pool()
	if err != nil {
		return err
	}

	if pool.RepoCount() <= 1 {
		return errors.New("No repo metadata found. Please run zpm refresh.")
	}

	if len(reqs) == 0 {
		image, err := m.image()
		if err != nil {
			return err
		}

		for _, pkg := range image.Solvables() {
			reqs = append(reqs, pkg.Name())
		}
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

	// Updates require the update policy
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
			fe := NewFetcher(uri, m.cache, m.security, m.config.CloudProvider())
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

func (m *Manager) ZpkgBuild(zpfPath string, targetPath string, workPath string, outputPath string, restrict bool, secure bool) error {
	builder := zpkg.NewBuilder()

	builder.Emitter = m.Emitter

	builder.ZpfPath(zpfPath).
		TargetPath(targetPath).WorkPath(workPath).
		OutputPath(outputPath).Restrict(restrict).
		Secure(secure)

	filename, manifest, err := builder.Build()
	if err != nil {
		return err
	}

	kp, err := m.security.KeyPair(manifest.Zpkg.Publisher)
	if err != nil {
		return err
	}

	if kp == nil {
		m.Emitter.Emit("manager.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", manifest.Zpkg.Publisher))

		return err
	}

	signer := zpkg.NewSigner(filename, workPath)

	rsaKey, err := kp.RSAKey()
	if err != nil {
		return err
	}

	err = signer.Sign(kp.Fingerprint, rsaKey)
	if err == nil {
		m.Emitter.Emit("manager.info", fmt.Sprintf("Signed with keypair: %s", kp.Subject))
	}

	return err
}

// TODO consider merging with Contents command via file path sniffing
func (m *Manager) ZpkgContents(path string) ([]string, error) {
	reader := zpkg.NewReader(path, "")

	err := reader.Read()
	if err != nil {
		return nil, err
	}

	contents := reader.Manifest.Section("Dir", "SymLink", "File")

	sort.Sort(contents)

	var output []string
	for _, fsObject := range contents {
		output = append(output, fsObject.Columns())
	}

	return output, nil
}

func (m *Manager) ZpkgExtract(filepath string, target string) error {
	reader := zpkg.NewReader(filepath, "")

	err := reader.Read()
	if err != nil {
		return err
	}

	options := &provider.Options{TargetPath: target}
	ctx := m.getContext(phase.INSTALL, options)
	ctx = context.WithValue(ctx, "payload", reader.Payload)

	contents := reader.Manifest.Section("Dir", "SymLink", "File")
	sort.Sort(contents)

	factory := provider.DefaultFactory(m.Emitter)

	for _, fsObject := range contents {
		m.Emit("manager.info", fmt.Sprintf("Extracted => %s %s", strings.ToUpper(fsObject.Type()), path.Join(target, fsObject.Key())))

		err = factory.Get(fsObject).Realize(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO consider merging with Info command utilizing file path sniffing
func (m *Manager) ZpkgInfo(path string) (string, error) {
	reader := zpkg.NewReader(path, "")

	err := reader.Read()
	if err != nil {
		return "", err
	}

	pkg, err := zps.NewPkgFromManifest(reader.Manifest)
	if err != nil {
		return "", err
	}

	info := fmt.Sprint("Package: ", pkg.Id(), "\n") +
		fmt.Sprint("Name: ", pkg.Name(), "\n") +
		fmt.Sprint("Publisher: ", pkg.Publisher(), "\n") +
		fmt.Sprint("Semver: ", pkg.Version().Semver.String(), "\n") +
		fmt.Sprint("Timestamp: ", pkg.Version().Timestamp, "\n") +
		fmt.Sprint("OS: ", pkg.Os(), "\n") +
		fmt.Sprint("Arch: ", pkg.Arch(), "\n") +
		fmt.Sprint("Summary: ", pkg.Summary(), "\n") +
		fmt.Sprint("Description: ", pkg.Description(), "\n")

	return info, nil
}

// TODO consider reworking into Manifest command utilizing file path sniffing
func (m *Manager) ZpkgManifest(path string) (string, error) {
	reader := zpkg.NewReader(path, "")

	err := reader.Read()
	if err != nil {
		return "", err
	}

	var manifest bytes.Buffer
	err = json.Indent(&manifest, []byte(reader.Manifest.ToJson()), "", "    ")
	if err != nil {
		return "", err
	}

	return manifest.String(), nil
}

func (m *Manager) ZpkgSign(path string, workPath string) error {
	reader := zpkg.NewReader(path, workPath)

	err := reader.Read()
	if err != nil {
		return err
	}
	reader.Close()

	kp, err := m.security.KeyPair(reader.Manifest.Zpkg.Publisher)
	if err != nil {
		return err
	}

	if kp == nil {
		m.Emitter.Emit("manager.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", reader.Manifest.Zpkg.Publisher))

		return err
	}

	signer := zpkg.NewSigner(path, workPath)

	rsaKey, err := kp.RSAKey()
	if err != nil {
		return err
	}

	err = signer.Sign(kp.Fingerprint, rsaKey)
	if err == nil {
		m.Emitter.Emit("manager.info", fmt.Sprintf("Signed with keypair: %s", kp.Subject))
	}

	return err
}

// TODO also verify file digests
func (m *Manager) ZpkgValidate(path string) error {
	return ValidateZpkg(m.Emitter, m.security, path, false)
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

func (m *Manager) getContext(phase string, options *provider.Options) context.Context {
	ctx := context.WithValue(context.Background(), "phase", phase)
	ctx = context.WithValue(ctx, "options", options)

	return ctx
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

				// Validate metadata signature
				if m.security.Mode() != SecurityModeNone {
					err := ValidateFileSignature(m.security, m.cache.GetMeta(osarch.String(), r.Fetch.Uri.String()), m.cache.GetMetaSig(osarch.String(), r.Fetch.Uri.String()))
					if err != nil {
						m.Emit("manager.error", fmt.Sprintf("invalid metadata signature: %s", r.Fetch.Uri))
						continue
					}
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

	// Validate config signature
	if m.security.Mode() != SecurityModeNone {
		err := ValidateFileSignature(m.security, m.cache.GetConfig(uri), m.cache.GetConfigSig(uri))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Repo config signature validation failed: %s", uri))
		}
	}

	configDb := NewConfig(configPath)

	cfg, err := configDb.All()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (m *Manager) splitReqsFiles(args []string) ([]string, []string, error) {
	var reqs []string
	var files []string

	for _, item := range args {
		if filepath.Ext(item) == ".zpkg" {
			if _, err := os.Stat(item); err == nil {
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

				itemPath, err := filepath.Abs(item)
				if err != nil {
					return nil, nil, err
				}

				files = append(files, itemPath)

			}
		} else {
			reqs = append(reqs, item)
		}
	}

	return reqs, files, nil
}

func (m *Manager) IsImage(imagePath string) bool {
	if _, err := os.Stat(filepath.Join(imagePath, "usr", "bin", "zps")); os.IsNotExist(err) {
		return false
	}

	return true
}

func (m *Manager) IsEmptyImage(imagePath string) (bool, error) {
	entries, err := ioutil.ReadDir(imagePath)
	if err != nil {
		return false, err
	}

	if len(entries) == 0 {
		return true, nil
	}

	if len(entries) == 1 {
		if strings.Contains("Imagefile", entries[0].Name()) {
			return true, nil
		}
	}

	return false, nil
}

func (m *Manager) EmptyImage(imagePath string) error {
	entries, err := ioutil.ReadDir(imagePath)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	for _, entry := range entries {
		if strings.Contains("Imagefile", entry.Name()) || strings.Contains("var", entry.Name()) {
			continue
		} else {
			os.RemoveAll(filepath.Join(imagePath, entry.Name()))
		}
	}

	// TODO implement tmp, or repo data preloading to avoid this
	// Cleanup var leaving only zps cache
	if _, err := os.Stat(filepath.Join(imagePath, "var")); !os.IsNotExist(err) {
		varEntries, err := ioutil.ReadDir(filepath.Join(imagePath, "var"))
		if err != nil {
			return err
		}

		for _, entry := range varEntries {
			if strings.Contains("cache", entry.Name()) || strings.Contains("lib", entry.Name()) {
				continue
			} else {
				os.RemoveAll(filepath.Join(imagePath, "var", entry.Name()))
			}
		}

		if _, err := os.Stat(filepath.Join(imagePath, "var", "cache")); !os.IsNotExist(err) {
			cacheEntries, err := ioutil.ReadDir(filepath.Join(imagePath, "var", "cache"))
			if err != nil {
				return err
			}

			for _, entry := range cacheEntries {
				if strings.Contains("zps", entry.Name()) {
					continue
				} else {
					os.RemoveAll(filepath.Join(imagePath, "var", "cache", entry.Name()))
				}
			}
		}

		if _, err := os.Stat(filepath.Join(imagePath, "var", "lib")); !os.IsNotExist(err) {
			cacheEntries, err := ioutil.ReadDir(filepath.Join(imagePath, "var", "lib"))
			if err != nil {
				return err
			}

			for _, entry := range cacheEntries {
				if strings.Contains("zps", entry.Name()) {
					continue
				} else {
					os.RemoveAll(filepath.Join(imagePath, "var", "lib", entry.Name()))
				}
			}
		}

		os.Remove(filepath.Join(imagePath, "var", "lib", "zps", "image.db"))
	}

	return nil
}
