/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package config

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"

	"github.com/hashicorp/hcl2/hclparse"

	"errors"
	"fmt"

	"net/url"

	"runtime"

	"github.com/hashicorp/hcl2/gohcl"
)

type ZpsConfig struct {
	Mode     string `hcl:"mode"`
	Security string `hcl:"security"`

	Root         string
	CurrentImage *ImageConfig

	Images []*ImageConfig
	Repos  []*RepoConfig

	hclCtx *hcl.EvalContext
}

func LoadConfig(image string) (*ZpsConfig, error) {
	var err error

	config := &ZpsConfig{}

	err = config.SetRoot()
	if err != nil {
		return nil, err
	}

	// Load configured image list based on current root
	err = config.LoadImages()
	if err != nil {
		return nil, err
	}

	// If image name, path, or hash id provided override image prefix
	err = config.SelectImage(image)
	if err != nil {
		return nil, err
	}

	// Setup shell helper
	err = config.SetupHelper(false)
	if err != nil {
		return nil, err
	}

	// Load image zps config
	err = config.LoadMain()
	if err != nil {
		return nil, err
	}

	// Load repository configs
	err = config.LoadRepos()
	if err != nil {
		return nil, err
	}

	// Load HCL Eval Context
	err = config.LoadHclContext()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (z *ZpsConfig) SetRoot() error {
	var err error

	z.Root, err = InstallPrefix()
	if err != nil {
		return err
	}

	return nil
}

func (z *ZpsConfig) SetupHelper(overwrite bool) error {
	if z.UserPath() == "" {
		return nil
	}

	zpsUserPath := z.UserPath()

	if _, err := os.Stat(zpsUserPath); os.IsNotExist(err) {
		os.Mkdir(zpsUserPath, 0700)
	}

	zpsUserSettingsPath := filepath.Join(zpsUserPath, "init.sh")

	_, err := os.Stat(zpsUserSettingsPath)

	if os.IsNotExist(err) || overwrite {
		err := ioutil.WriteFile(zpsUserSettingsPath, []byte(fmt.Sprintf(ZshHelper, z.CurrentImage.Path)), 0600)
		if err != nil {
			return err
		}
	}

	return err
}

func (z *ZpsConfig) ConfigPath() string {
	if z.CurrentImage == nil {
		return filepath.Join(z.Root, "etc", "zps")
	}

	return filepath.Join(z.CurrentImage.Path, "etc", "zps")
}

func (z *ZpsConfig) CachePath() string {
	return filepath.Join(z.CurrentImage.Path, "var", "cache", "zps")
}

func (z *ZpsConfig) CertPath() string {
	return filepath.Join(z.CurrentImage.Path, "usr", "share", "zps", "certs")
}

func (z *ZpsConfig) LockPath() string {
	return filepath.Join(z.CurrentImage.Path, "var", "lib", "zps")
}

func (z *ZpsConfig) StatePath() string {
	return filepath.Join(z.CurrentImage.Path, "var", "lib", "zps")
}

func (z *ZpsConfig) PkiPath() string {
	return filepath.Join(z.CurrentImage.Path, "var", "lib", "zps")
}

func (z *ZpsConfig) WorkPath() string {
	return filepath.Join(z.CurrentImage.Path, "var", "tmp", "zps")
}

func (z *ZpsConfig) UserPath() string {
	if os.Getenv("HOME") == "" {
		return ""
	}

	return filepath.Join(os.Getenv("HOME"), ".zps")
}

func (z *ZpsConfig) LoadImages() error {
	defaultOs := runtime.GOOS
	var defaultArch string
	switch runtime.GOARCH {
	case "amd64":
		defaultArch = "x86_64"
	default:
		defaultArch = runtime.GOARCH
	}

	defaultImage := &ImageConfig{"default", z.Root, defaultOs, defaultArch}

	z.Images = append(z.Images, defaultImage)

	loadPath := z.ConfigPath()

	// Override image load path if external config exists
	if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".zps", "image.d")); !os.IsNotExist(err) {
		loadPath = filepath.Join(os.Getenv("HOME"), ".zps")
	}

	// Load defined images
	imageConfigs, err := filepath.Glob(filepath.Join(loadPath, "image.d", "*.conf"))
	if err != nil {
		return nil
	}

	for _, cfgPath := range imageConfigs {
		image := &ImageConfig{}
		parser := hclparse.NewParser()

		bytes, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			return nil
		}

		// Parse HCL
		ihcl, diag := parser.ParseHCL(bytes, cfgPath)
		if diag.HasErrors() {
			return diag
		}

		// Eval HCL
		diag = gohcl.DecodeBody(ihcl.Body, nil, image)
		if diag.HasErrors() {
			return diag
		}

		z.Images = append(z.Images, image)
	}

	return nil
}

func (z *ZpsConfig) SelectImage(image string) error {
	// Allow fallthrough for named image for matching path
	if image == "" {
		z.CurrentImage = z.Images[0]
		image = z.CurrentImage.Path
	}

	// Select image, do not return early since we want to return the defined name but preserve the default entry
	for index, img := range z.Images {
		if img.Path == image {
			z.CurrentImage = z.Images[index]
		}

		if img.Name == image {
			z.CurrentImage = z.Images[index]
		}
	}

	if z.CurrentImage != nil {
		return nil
	}

	return errors.New(fmt.Sprintf("image not found or configured: %s", image))
}

func (z *ZpsConfig) LoadMain() error {
	parser := hclparse.NewParser()

	bytes, err := ioutil.ReadFile(path.Join(z.ConfigPath(), "main.conf"))
	if err != nil {
		// Generate defaults for now so we don't have to ship a default config
		if os.IsNotExist(err) {
			z.Mode = "ancillary"
			z.Security = "offline"

			return nil
		}

		return err
	}

	// Parse HCL
	mhcl, diag := parser.ParseHCL(bytes, path.Join(z.ConfigPath(), "main.conf"))
	if diag.HasErrors() {
		return diag
	}

	// Eval HCL
	diag = gohcl.DecodeBody(mhcl.Body, nil, z)
	if diag.HasErrors() {
		return diag
	}

	return nil
}

func (z *ZpsConfig) LoadRepos() error {
	// Load defined repos
	repoConfigs, err := filepath.Glob(path.Join(z.ConfigPath(), "repo.d", "*.conf"))
	if err != nil {
		return nil
	}

	// TODO raise a warning for bad file, continue instead of returning
	for _, rconfig := range repoConfigs {
		repo := &RepoConfig{}
		parser := hclparse.NewParser()

		bytes, err := ioutil.ReadFile(rconfig)
		if err != nil {
			continue
		}

		// Parse HCL
		repoHcl, diag := parser.ParseHCL(bytes, rconfig)
		if diag.HasErrors() {
			continue
		}

		// Eval HCL
		diag = gohcl.DecodeBody(repoHcl.Body, nil, repo)
		if diag.HasErrors() {
			continue
		}

		// Validate fetch section
		if repo.Fetch != nil {
			if repo.Fetch.UriString != "" {
				repo.Fetch.Uri, err = url.Parse(repo.Fetch.UriString)
				if err != nil {
					return err
				}
			} else {
				return errors.New(fmt.Sprint("config: repo fetch.uri required in ", rconfig))
			}
		} else {
			return errors.New(fmt.Sprint("config: repo fetch section required in ", rconfig))
		}

		if repo.Publish != nil {
			if repo.Publish.UriString != "" {
				repo.Publish.Uri, err = url.Parse(repo.Publish.UriString)
				if err != nil {
					return err
				}
			} else {
				return errors.New(fmt.Sprint("config: repo publish.uri required in ", rconfig))
			}
		}

		z.Repos = append(z.Repos, repo)
	}

	return nil
}

func (z *ZpsConfig) LoadHclContext() error {
	z.hclCtx = &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Load env namespace
	envs := make(map[string]cty.Value)
	for _, env := range os.Environ() {
		key := strings.Split(env, "=")[0]
		val, _ := os.LookupEnv(key)
		envs[key] = cty.StringVal(val)
	}

	z.hclCtx.Variables["env"] = cty.ObjectVal(envs)
	z.hclCtx.Functions = map[string]function.Function{
		"upper":   stdlib.UpperFunc,
		"lower":   stdlib.LowerFunc,
		"length":  stdlib.LengthFunc,
		"os_test": z.osTest(),
	}

	return nil
}

func (z *ZpsConfig) HclContext() *hcl.EvalContext {
	return z.hclCtx
}
