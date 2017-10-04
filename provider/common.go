package provider

import (
	"golang.org/x/net/context"

	"github.com/solvent-io/zps/action"
	"github.com/solvent-io/zps/provider/dir"
	"github.com/solvent-io/zps/provider/file"
	"github.com/solvent-io/zps/provider/meta"
	"github.com/solvent-io/zps/provider/requirement"
	"github.com/solvent-io/zps/provider/symlink"
	"github.com/solvent-io/zps/provider/zpkg"
)

type Provider interface {
	Realize(phase string, ctx context.Context) error
}

// Need to add provider switching
// for now defaults will work on all OSs we care about
func Get(ac action.Action) Provider {
	switch ac.Type() {
	case "zpkg":
		return zpkg.NewDefault(ac)
	case "requirement":
		return requirement.NewDefault(ac)
	case "meta":
		return meta.NewDefault(ac)
	case "dir":
		return dir.NewUnix(ac)
	case "file":
		return file.NewUnix(ac)
	case "symlink":
		return symlink.NewUnix(ac)
	}

	return nil
}
