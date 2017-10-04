package symlink

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"syscall"

	"golang.org/x/net/context"

	"github.com/naegelejd/go-acl/os/group"
	"github.com/solvent-io/zps/action"
)

type Unix struct {
	symlink *action.SymLink
}

func NewUnix(symlink action.Action) *Unix {
	return &Unix{symlink.(*action.SymLink)}
}

func (u *Unix) Realize(phase string, ctx context.Context) error {
	switch phase {
	case "install":
		return u.install(ctx)
	case "package":
		return u.pkg(ctx)
	case "remove":
		return u.remove(ctx)
	default:
		return nil
	}
}

func (u *Unix) install(ctx context.Context) error {
	options := ctx.Value("options").(*action.Options)
	target := path.Join(options.TargetPath, u.symlink.Path)

	err := os.Symlink(u.symlink.Target, target)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Silent failures are fine, only a super user can chown to another user
	// Also a given user may not exist on a system though we should catch
	// that elsewhere

	owner, _ := user.Lookup(u.symlink.Owner)
	group, _ := group.Lookup(u.symlink.Group)
	var uid int64
	var gid int64

	if owner != nil && group != nil {
		uid, _ = strconv.ParseInt(owner.Uid, 0, 0)
		gid, _ = strconv.ParseInt(owner.Uid, 0, 0)
	}

	os.Chown(target, int(uid), int(gid))

	return nil
}

func (u *Unix) pkg(ctx context.Context) error {
	options := ctx.Value("options").(*action.Options)
	target := path.Join(options.TargetPath, u.symlink.Path)

	info, err := os.Lstat(target)
	if err != nil {
		return err
	}

	u.symlink.Target, err = os.Readlink(target)
	if err != nil {
		return err
	}

	if u.symlink.Owner == "" {
		if options.Secure {
			u.symlink.Owner = "root"
		} else if options.Owner != "" {
			u.symlink.Owner = options.Owner
		} else {
			user, err := user.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Uid))
			if err != nil {
				return err
			}
			u.symlink.Owner = user.Username
		}
	}

	if u.symlink.Group == "" {
		if options.Secure {
			u.symlink.Group = "root"
		} else if options.Group != "" {
			u.symlink.Group = options.Group
		} else {
			group, err := group.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				return err
			}
			u.symlink.Group = group.Name
		}
	}

	return err
}

func (u *Unix) remove(ctx context.Context) error {
	options := ctx.Value("options").(*action.Options)
	target := path.Join(options.TargetPath, u.symlink.Path)

	err := os.Remove(target)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}
