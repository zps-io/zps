package file

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"syscall"

	"github.com/naegelejd/go-acl/os/group"
	"github.com/solvent-io/zps/action"
	zpayload "github.com/solvent-io/zps/zpkg/payload"
	"golang.org/x/net/context"
)

type Unix struct {
	file *action.File
}

func NewUnix(file action.Action) *Unix {
	return &Unix{file.(*action.File)}
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
	payload := ctx.Value("payload").(*zpayload.Reader)

	target := path.Join(options.TargetPath, u.file.Path)

	mode, err := strconv.ParseUint(u.file.Mode, 0, 0)
	if err != nil {
		return err
	}

	if u.file.Size != 0 {
		var hash string
		var err error

		hash, err = payload.Get(target, int64(u.file.Offset), int64(u.file.Size))
		if err != nil {
			return err
		}

		if hash != u.file.Hash {
			return errors.New(fmt.Sprint("checksum does not match manifest for: ", target))
		}
	} else {
		file, err := os.Create(target)
		if err != nil {
			return err
		}

		file.Close()
	}
	// Silent failures are fine, only a super user can chown to another user
	// Also a given user may not exist on a system though we should catch
	// that elsewhere

	owner, _ := user.Lookup(u.file.Owner)
	group, _ := group.Lookup(u.file.Group)
	var uid int64
	var gid int64

	if owner != nil && group != nil {
		uid, _ = strconv.ParseInt(owner.Uid, 0, 0)
		gid, _ = strconv.ParseInt(owner.Uid, 0, 0)
	}

	os.Chown(target, int(uid), int(gid))
	os.Chmod(target, os.FileMode(mode))
	return nil
}

func (u *Unix) pkg(ctx context.Context) error {
	options := ctx.Value("options").(*action.Options)
	payload := ctx.Value("payload").(*zpayload.Writer)

	target := path.Join(options.TargetPath, u.file.Path)

	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	if u.file.Mode == "" {
		u.file.Mode = fmt.Sprintf("%#o", info.Mode().Perm())
	}

	if u.file.Owner == "" {
		if options.Secure {
			u.file.Owner = "root"
		} else if options.Owner != "" {
			u.file.Owner = options.Owner
		} else {
			user, err := user.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Uid))
			if err != nil {
				return err
			}
			u.file.Owner = user.Username
		}
	}

	if u.file.Group == "" {
		if options.Secure {
			u.file.Group = "root"
		} else if options.Group != "" {
			u.file.Group = options.Group
		} else {
			group, err := group.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				return err
			}
			u.file.Group = group.Name
		}
	}

	u.file.Size = int(info.Size())

	// Add to payload
	if u.file.Size != 0 {
		u.file.Offset, u.file.Csize, u.file.Hash, err = payload.Put(target)
	} else {
		u.file.Offset = 0
		u.file.Csize = 0
		u.file.Hash = ""
	}

	return err
}

func (u *Unix) remove(ctx context.Context) error {
	options := ctx.Value("options").(*action.Options)
	target := path.Join(options.TargetPath, u.file.Path)

	err := os.Remove(target)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}
