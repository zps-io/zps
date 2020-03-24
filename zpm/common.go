package zpm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"

	"github.com/fezz-io/zps/config"

	"github.com/fezz-io/zps/action"

	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/phase"
	"github.com/fezz-io/zps/provider"
	"github.com/fezz-io/zps/zpkg"

	"github.com/fezz-io/zps/zps"
)

func FilterPackagesByArch(osarch *zps.OsArch, zpkgs map[string]*zps.Pkg) ([]string, []*zps.Pkg) {
	var files []string
	var pkgs []*zps.Pkg

	for k, v := range zpkgs {
		if v.Os() == osarch.Os && v.Arch() == osarch.Arch {
			files = append(files, k)
			pkgs = append(pkgs, zpkgs[k])
		}
	}

	return files, pkgs
}

func PublisherFromUri(uri *url.URL) string {
	parts := strings.Split(uri.Path, "/")

	if len(parts) < 2 {
		return ""
	} else {
		return parts[len(parts)-2]
	}
}

func ValidateFileSignature(security Security, contentPath string, signaturePath string) error {
	contentBytes, err := ioutil.ReadFile(contentPath)
	if err != nil {
		return err
	}

	sigBytes, err := ioutil.ReadFile(signaturePath)
	if err != nil {
		return err
	}

	sig := &action.Signature{}

	err = json.Unmarshal(sigBytes, sig)
	if err != nil {
		return err
	}

	_, err = security.Verify(&contentBytes, []*action.Signature{sig})

	return err
}

// TODO move to higher level zpkg util
func ValidateZpkg(emitter *emission.Emitter, security Security, path string, quiet bool) error {
	reader := zpkg.NewReader(path, "")

	err := reader.Read()
	if err != nil {
		return err
	}
	defer reader.Close()

	var content []byte
	content = []byte(reader.Manifest.ToSigningJson())

	sig, err := security.Verify(&content, reader.Manifest.Signatures)
	if err != nil {
		return err
	}

	if quiet == false {
		emitter.Emit("manager.info", fmt.Sprintf("Manifest signature validated with key fingerpint: %s", sig.FingerPrint))
	}

	// Validate payload
	options := &provider.Options{}

	ctx := context.WithValue(context.Background(), "phase", phase.VALIDATE)
	ctx = context.WithValue(ctx, "payload", reader.Payload)
	ctx = context.WithValue(ctx, "options", options)

	contents := reader.Manifest.Section("File")
	sort.Sort(contents)

	factory := provider.DefaultFactory(emitter)

	if quiet == false {
		emitter.Emit("manager.info", "Validating payload ...")
	}

	for _, fsObject := range contents {
		err = factory.Get(fsObject).Realize(ctx)
		if err != nil {
			return err
		}
	}

	if quiet == false {
		emitter.Emit("manager.info", fmt.Sprintf("Package verified: %s", path))
	}

	return nil
}

func MergeTemplateConfig(pkgTpls []*action.Template, cfgTpls []*config.TplConfig) []*action.Template {
	index := make(map[string]int)
	var tpls []*action.Template

	for i, tpl := range cfgTpls {
		tpls = append(tpls, &action.Template{
			Name:   tpl.Name,
			Source: tpl.Source,
			Output: tpl.Output,
			Owner:  tpl.Owner,
			Group:  tpl.Group,
			Mode:   tpl.Mode,
		})
		index[tpl.Name] = i
	}

	for i := range pkgTpls {
		if _, ok := index[pkgTpls[i].Name]; !ok {
			tpls = append(tpls, pkgTpls[i])
		}
	}

	return tpls
}
