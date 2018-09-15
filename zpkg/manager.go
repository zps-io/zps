package zpkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/solvent-io/zps/phase"
	"github.com/solvent-io/zps/zps"
	fpath "path"
	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/provider"
	"sort"
	"strings"
)

type Manager struct {
	*emission.Emitter
}

func NewManager() *Manager {
	return &Manager{Emitter: emission.NewEmitter()}
}

func (m *Manager) Contents(path string) ([]string, error) {
	reader := NewReader(path, "")

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

func (m *Manager) Extract(path string, target string) error {
	reader := NewReader(path, "")

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
		m.Emit("manager.info", fmt.Sprintf("Extracted => %s %s", strings.ToUpper(fsObject.Type()), fpath.Join(target, fsObject.Key())))

		err = factory.Get(fsObject).Realize(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) Info(path string) (string, error) {
	reader := NewReader(path, "")

	err := reader.Read()
	if err != nil {
		return "", err
	}

	pkg, err := zps.NewPkgFromManifest(reader.Manifest)
	if err != nil {
		return "", err
	}

	info := fmt.Sprint("Name: ", pkg.Name(), "\n") +
		fmt.Sprint("Publisher: ", pkg.Publisher(), "\n") +
		fmt.Sprint("Semver: ", pkg.Version().Semver.String(), "\n") +
		fmt.Sprint("Timestamp: ", pkg.Version().Timestamp, "\n") +
		fmt.Sprint("OS: ", pkg.Os(), "\n") +
		fmt.Sprint("Arch: ", pkg.Arch(), "\n") +
		fmt.Sprint("Summary: ", pkg.Summary(), "\n") +
		fmt.Sprint("Description: ", pkg.Description(), "\n")

	return info, nil
}

func (m *Manager) Manifest(path string) (string, error) {
	reader := NewReader(path, "")

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

func (m *Manager) getContext(phase string, options *provider.Options) context.Context {
	ctx := context.WithValue(context.Background(), "phase", phase)
	ctx = context.WithValue(ctx, "options", options)

	return ctx
}