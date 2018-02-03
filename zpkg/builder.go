package zpkg

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"time"

	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/action"
	"github.com/solvent-io/zps/namespace"
	"github.com/solvent-io/zps/provider"
	"github.com/solvent-io/zps/zpkg/payload"
	"github.com/solvent-io/zps/zps"
	"golang.org/x/net/context"
)

type Builder struct {
	*emission.Emitter

	options *action.Options

	zpfPath    string
	workPath   string
	outputPath string

	version uint8

	namespaces []namespace.Namespace

	manifest *action.Manifest

	filename string

	header  *Header
	payload *payload.Writer

	writer *Writer
}

func NewBuilder() *Builder {
	builder := &Builder{Emitter: emission.NewEmitter()}

	builder.version = Version

	builder.options = action.NewOptions()

	builder.namespaces = append(builder.namespaces, namespace.Get("vcs"))

	builder.manifest = action.NewManifest()

	builder.header = NewHeader(Version, Compression)
	builder.payload = payload.NewWriter("", 0)
	builder.writer = NewWriter()

	return builder
}

func (b *Builder) ZpfPath(zp string) *Builder {
	b.zpfPath = zp
	return b
}

func (b *Builder) TargetPath(tp string) *Builder {
	b.options.TargetPath = tp
	return b
}

func (b *Builder) Restrict(r bool) *Builder {
	b.options.Restrict = r
	return b
}

func (b *Builder) Secure(s bool) *Builder {
	b.options.Secure = s
	return b
}

func (b *Builder) WorkPath(wp string) *Builder {
	b.workPath = wp
	b.payload.WorkPath = wp
	return b
}

func (b *Builder) OutputPath(op string) *Builder {
	b.outputPath = op
	return b
}

func (b *Builder) Version(version uint8) *Builder {
	b.version = version
	b.header.Version = version
	return b
}

func (b *Builder) Namespace(name string) *Builder {
	ns := namespace.Get(name)
	if ns != nil {
		b.namespaces = append(b.namespaces, namespace.Get(name))
	}
	return b
}

// Set default paths
func (b *Builder) setPaths() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Can be avoided if we error on create
	// but it would break the builder pattern
	if b.zpfPath == "" {
		b.zpfPath = path.Join(wd, DefaultZpfPath)
	}

	// If the path is a directory append the default
	// ZpfPath
	stat, err := os.Stat(b.zpfPath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		b.zpfPath = path.Join(b.zpfPath, DefaultZpfPath)
	}

	if b.options.TargetPath == "" {
		dir, _ := path.Split(b.zpfPath)
		b.options.TargetPath = path.Join(dir, DefaultTargetDir)
	}
	if b.workPath == "" {
		b.workPath = wd
		b.payload.WorkPath = wd
	}
	if b.outputPath == "" {
		b.outputPath = wd
	}

	return err
}

// This isn't efficient, I don't expect these files to be terribly large however
func (b *Builder) loadZpkgfile() error {
	zpfbytes, err := ioutil.ReadFile(b.zpfPath)
	if err != nil {
		return err
	}

	if b.manifest.Load(string(zpfbytes)) != nil {
		return err
	}

	return err
}

// Process options deal with any special cases here
func (b *Builder) processOptions() error {

	return nil
}

// Add FS objects
func (b *Builder) resolve() error {
	// If restrict is set don't walk the target path
	// this will result in only defined file system objects being added
	// to the package
	if b.options.Restrict == true {
		return nil
	}

	err := filepath.Walk(b.options.TargetPath, func(path string, f os.FileInfo, err error) error {
		objectPath := strings.Replace(path, b.options.TargetPath+string(os.PathSeparator), "", 1)

		if objectPath != b.options.TargetPath {
			if f.IsDir() {
				var dir = action.NewDir()
				dir.Path = objectPath

				b.manifest.Add(dir)
			}

			if f.Mode().IsRegular() {
				var file = action.NewFile()
				file.Path = objectPath

				b.manifest.Add(file)
			}

			if f.Mode()&os.ModeSymlink == os.ModeSymlink {
				var symlink = action.NewSymLink()
				symlink.Path = objectPath

				b.manifest.Add(symlink)
			}
		}

		return nil
	})

	// Ensure we don't have differing filesystem actions with the same path
	var actions action.Actions = b.manifest.Section("dir", "file", "symlink")
	sort.Sort(actions)
	for index, action := range actions {
		prev := index - 1
		if prev != -1 {
			if action.Key() == actions[prev].Key() {
				return errors.New(fmt.Sprint(
					"Action Conflicts:\n",
					strings.ToUpper(actions[prev].Type()), " => ", actions[prev].Key(), "\n",
					strings.ToUpper(action.Type()), " => ", action.Key()))
			}
		}
	}

	return err
}

// Set file name and zpkg timestamp
func (b *Builder) set() error {
	zpkg := b.manifest.Section("zpkg")[0].(*action.Zpkg)

	uri := zps.NewZpkgUri()
	uri.Name = zpkg.Name
	uri.Publisher = zpkg.Publisher
	uri.Category = zpkg.Category

	err := uri.Version.Parse(zpkg.Version)
	if err != nil {
		return err
	}

	uri.Version.Timestamp = time.Now()
	zpkg.Uri = uri.String()

	// Unset uri component values
	zpkg.Name = ""
	zpkg.Version = ""
	zpkg.Publisher = ""
	zpkg.Category = ""

	pkg, err := zps.NewPkgFromManifest(b.manifest)
	if err != nil {
		return err
	}

	b.filename = pkg.FileName()

	return nil
}

// Validate the meta namespaces
func (b *Builder) validate() error {
	var err error

	for _, ns := range b.namespaces {
		err = ns.Validate(b.manifest)
		if err != nil {
			return err
		}
	}

	return err
}

// Completes manifest, builds payload
func (b *Builder) realize() error {
	var err error

	// Setup context
	ctx := action.GetContext(b.options, b.manifest)
	ctx = context.WithValue(ctx, "payload", b.payload)

	for _, act := range b.manifest.Actions {
		err = provider.Get(act).Realize("package", ctx)
		if err != nil {
			return err
		}

		b.Emit("action", act)
	}

	return err
}

func (b *Builder) Build() (string, error) {
	err := b.setPaths()
	if err != nil {
		return "", err
	}

	err = b.loadZpkgfile()
	if err != nil {
		return "", err
	}

	err = b.processOptions()
	if err != nil {
		return "", err
	}

	err = b.resolve()
	if err != nil {
		return "", err
	}

	err = b.set()
	if err != nil {
		return "", err
	}

	err = b.validate()
	if err != nil {
		return "", err
	}

	err = b.realize()
	if err != nil {
		return "", err
	}

	// Write the file
	err = b.writer.Write(b.filename, b.header, b.manifest, b.payload)
	if err != nil {
		return "", err
	}

	b.Emit("complete", b.filename)

	return b.filename, err
}
