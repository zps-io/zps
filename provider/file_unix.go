package provider

import (
	"errors"
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
	zpayload "github.com/solvent-io/zps/zpkg/payload"
)

type FileUnix struct {
	*emission.Emitter
	file *action.File

	phaseMap map[string]string
}

func NewFileUnix(file action.Action, phaseMap map[string]string, emitter *emission.Emitter) Provider {
	return &FileUnix{emitter, file.(*action.File), phaseMap}
}

func (f *FileUnix) Realize(ctx context.Context) error {
	switch f.phaseMap[Phase(ctx)] {
	case "install":
		return f.install(ctx)
	case "package":
		return f.pkg(ctx)
	case "remove":
		return f.remove(ctx)
	default:
		return nil
	}
}

func (f *FileUnix) install(ctx context.Context) error {
	options := Opts(ctx)
	payload := ctx.Value("payload").(*zpayload.Reader)

	target := path.Join(options.TargetPath, f.file.Path)

	mode, err := strconv.ParseUint(f.file.Mode, 0, 0)
	if err != nil {
		return err
	}

	if f.file.Size != 0 {
		var hash string
		var err error

		hash, err = payload.Get(target, int64(f.file.Offset), int64(f.file.Size))
		if err != nil {
			return err
		}

		if hash != f.file.Hash {
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

	owner, _ := user.Lookup(f.file.Owner)
	grp, _ := group.Lookup(f.file.Group)
	var uid int64
	var gid int64

	if owner != nil && grp != nil {
		uid, _ = strconv.ParseInt(owner.Uid, 0, 0)
		gid, _ = strconv.ParseInt(owner.Uid, 0, 0)
	}

	os.Chown(target, int(uid), int(gid))
	os.Chmod(target, os.FileMode(mode))

	return nil
}

func (f *FileUnix) pkg(ctx context.Context) error {
	options := Opts(ctx)
	payload := ctx.Value("payload").(*zpayload.Writer)

	target := path.Join(options.TargetPath, f.file.Path)

	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	if f.file.Mode == "" {
		f.file.Mode = fmt.Sprintf("%#o", info.Mode().Perm())
	}

	if f.file.Owner == "" {
		if options.Secure {
			f.file.Owner = "root"
		} else if options.Owner != "" {
			f.file.Owner = options.Owner
		} else {
			usr, err := user.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Uid))
			if err != nil {
				return err
			}
			f.file.Owner = usr.Username
		}
	}

	if f.file.Group == "" {
		if options.Secure {
			f.file.Group = "root"
		} else if options.Group != "" {
			f.file.Group = options.Group
		} else {
			grp, err := group.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				return err
			}
			f.file.Group = grp.Name
		}
	}

	f.file.Size = int(info.Size())

	// Add to payload
	if f.file.Size != 0 {
		f.file.Offset, f.file.Csize, f.file.Hash, err = payload.Put(target)
	} else {
		f.file.Offset = 0
		f.file.Csize = 0
		f.file.Hash = ""
	}

	return err
}

func (f *FileUnix) remove(ctx context.Context) error {
	options := Opts(ctx)
	target := path.Join(options.TargetPath, f.file.Path)

	err := os.Remove(target)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}
