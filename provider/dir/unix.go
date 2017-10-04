package dir

import (
	"fmt"
	"io"
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
	dir *action.Dir
}

func NewUnix(dir action.Action) *Unix {
	return &Unix{dir.(*action.Dir)}
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
	target := path.Join(options.TargetPath, u.dir.Path)

	mode, err := strconv.ParseUint(u.dir.Mode, 0, 0)
	if err != nil {
		return err
	}

	// Allow chmod if exist for now ...
	err = os.Mkdir(target, os.FileMode(mode))
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Silent failures are fine, only a super user can chown to another user
	// Also a given user may not exist on a system though we should catch
	// that elsewhere

	owner, _ := user.Lookup(u.dir.Owner)
	group, _ := group.Lookup(u.dir.Group)
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
	target := path.Join(options.TargetPath, u.dir.Path)

	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	if u.dir.Mode == "" {
		u.dir.Mode = fmt.Sprintf("%#o", info.Mode().Perm())
	}

	if u.dir.Owner == "" {
		if options.Secure {
			u.dir.Owner = "root"
		} else if options.Owner != "" {
			u.dir.Owner = options.Owner
		} else {
			user, err := user.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Uid))
			if err != nil {
				return err
			}
			u.dir.Owner = user.Username
		}
	}

	if u.dir.Group == "" {
		if options.Secure {
			u.dir.Group = "root"
		} else if options.Group != "" {
			u.dir.Group = options.Group
		} else {
			group, err := group.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				return err
			}
			u.dir.Group = group.Name
		}
	}

	return err
}

func (u *Unix) remove(ctx context.Context) error {
	options := ctx.Value("options").(*action.Options)
	target := path.Join(options.TargetPath, u.dir.Path)

	empty, err := u.isEmpty(target)
	if err != nil {
		return err
	}

	if empty == false {
		return nil
	}

	return os.Remove(target)
}

func (u *Unix) isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
