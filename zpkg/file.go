/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package zpkg

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/solvent-io/zps/action"
	"github.com/zclconf/go-cty/cty"
)

type ZpkgFile struct {
	Bytes []byte
	Path  string

	ctx *hcl.EvalContext
	hcl *hcl.File
}

func (z *ZpkgFile) Load(path string) (*ZpkgFile, error) {
	var err error

	z.Path = path

	z.Bytes, err = ioutil.ReadFile(z.Path)
	if err != nil {
		return nil, err
	}

	z.ctx = &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	return z, nil
}

func (z *ZpkgFile) Eval() (*action.Manifest, error) {
	var diag hcl.Diagnostics

	manifest := action.NewManifest()
	parser := hclparse.NewParser()

	// Parse HCL
	z.hcl, diag = parser.ParseHCL(z.Bytes, z.Path)
	if diag.HasErrors() {
		return nil, diag
	}

	// Populate env namespace
	envs := make(map[string]cty.Value)
	for _, env := range os.Environ() {
		key := strings.Split(env, "=")[0]
		val, _ := os.LookupEnv(key)
		envs[key] = cty.StringVal(val)
	}
	z.ctx.Variables["env"] = cty.ObjectVal(envs)

	// Eval HCL with context
	diag = gohcl.DecodeBody(z.hcl.Body, z.ctx, manifest)
	if diag.HasErrors() {
		return nil, diag
	}

	return manifest, nil
}
