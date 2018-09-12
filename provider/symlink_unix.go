package provider

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"syscall"

	"github.com/chuckpreslar/emission"

	"context"

	"github.com/naegelejd/go-acl/os/group"
	"github.com/solvent-io/zps/action"
)

type SymLinkUnix struct {
	*emission.Emitter
	symlink *action.SymLink

	phaseMap map[string]string
}

func NewSymLinkUnix(symlink action.Action, phaseMap map[string]string, emitter *emission.Emitter) *SymLinkUnix {
	return &SymLinkUnix{emitter, symlink.(*action.SymLink), phaseMap}
}

func (s *SymLinkUnix) Realize(ctx context.Context) error {
	switch s.phaseMap[Phase(ctx)] {
	case "install":
		return s.install(ctx)
	case "package":
		return s.pkg(ctx)
	case "remove":
		return s.remove(ctx)
	default:
		return nil
	}
}

func (s *SymLinkUnix) install(ctx context.Context) error {
	options := Opts(ctx)
	target := path.Join(options.TargetPath, s.symlink.Path)

	err := os.Symlink(s.symlink.Target, target)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Silent failures are fine, only a super user can chown to another user
	// Also a given user may not exist on a system though we should catch
	// that elsewhere

	owner, _ := user.Lookup(s.symlink.Owner)
	grp, _ := group.Lookup(s.symlink.Group)
	var uid int64
	var gid int64

	if owner != nil && grp != nil {
		uid, _ = strconv.ParseInt(owner.Uid, 0, 0)
		gid, _ = strconv.ParseInt(owner.Uid, 0, 0)
	}

	os.Chown(target, int(uid), int(gid))

	return nil
}

func (s *SymLinkUnix) pkg(ctx context.Context) error {
	options := Opts(ctx)
	target := path.Join(options.TargetPath, s.symlink.Path)

	info, err := os.Lstat(target)
	if err != nil {
		return err
	}

	s.symlink.Target, err = os.Readlink(target)
	if err != nil {
		return err
	}

	if s.symlink.Owner == "" {
		if options.Secure {
			s.symlink.Owner = "root"
		} else if options.Owner != "" {
			s.symlink.Owner = options.Owner
		} else {
			usr, err := user.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Uid))
			if err != nil {
				return err
			}
			s.symlink.Owner = usr.Username
		}
	}

	if s.symlink.Group == "" {
		if options.Secure {
			s.symlink.Group = "root"
		} else if options.Group != "" {
			s.symlink.Group = options.Group
		} else {
			grp, err := group.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				return err
			}
			s.symlink.Group = grp.Name
		}
	}

	return err
}

func (s *SymLinkUnix) remove(ctx context.Context) error {
	options := Opts(ctx)
	target := path.Join(options.TargetPath, s.symlink.Path)

	err := os.Remove(target)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}
