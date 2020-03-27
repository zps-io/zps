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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

type ImageConfig struct {
	Name string `hcl:"name"`
	Path string `hcl:"path"`
	Os   string `hcl:"os"`
	Arch string `hcl:"arch"`
}

type PkgConfig struct {
	Name      string `hcl:"name,label"`
	Operation string `hcl:"operation,optional"`
	Version   string `hcl:"version,optional"`
}

type ImageFile struct {
	Name string `hcl:"name"`
	Path string `hcl:"path,optional"`
	Os   string `hcl:"os,optional"`
	Arch string `hcl:"arch,optional"`

	Repos     []*RepoConfig  `hcl:"Repo,block"`
	Configs   []*Config      `hcl:"Config,block"`
	Templates []*TplConfig   `hcl:"Template,block"`
	Packages  []*PkgConfig   `hcl:"Package,block"`
	Trusts    []*TrustConfig `hcl:"Trust,block"`

	FilePath string
}

type TrustConfig struct {
	Publisher string `hcl:"publisher,label"`
	Uri       string `hcl:"uri"`
}

func (i *ImageFile) Load(imageFilePath string) error {
	var err error

	i.FilePath = imageFilePath

	if i.FilePath == "" {
		i.FilePath, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	if stat, err := os.Stat(i.FilePath); !os.IsNotExist(err) {
		if stat.IsDir() {
			i.FilePath = filepath.Join(i.FilePath, "Imagefile")
			if _, err := os.Stat(i.FilePath); os.IsNotExist(err) {
				return fmt.Errorf("Imagefile not found: %s", i.FilePath)
			}
		}
	}

	parser := hclparse.NewParser()

	bytes, err := ioutil.ReadFile(i.FilePath)
	if err != nil {
		return nil
	}

	// Parse HCL
	ihcl, diag := parser.ParseHCL(bytes, i.FilePath)
	if diag.HasErrors() {
		return diag
	}

	// Setup context
	// TODO this code is in three places, do something about that
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Load env namespace
	envs := make(map[string]cty.Value)
	for _, env := range os.Environ() {
		key := strings.Split(env, "=")[0]
		val, _ := os.LookupEnv(key)
		envs[key] = cty.StringVal(val)
	}

	ctx.Variables["env"] = cty.ObjectVal(envs)

	// Eval HCL
	diag = gohcl.DecodeBody(ihcl.Body, ctx, i)
	if diag.HasErrors() {
		return diag
	}

	return nil
}
func (i *ImageConfig) ToHclFile() *hclwrite.File {
	file := hclwrite.NewEmptyFile()

	file.Body().SetAttributeValue("name", cty.StringVal(i.Name))
	file.Body().SetAttributeValue("path", cty.StringVal(i.Path))
	file.Body().SetAttributeValue("os", cty.StringVal(i.Os))
	file.Body().SetAttributeValue("arch", cty.StringVal(i.Arch))

	return file
}
