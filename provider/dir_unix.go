/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package provider

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"strconv"
	"syscall"

	"github.com/chuckpreslar/emission"

	"context"

	"github.com/fezz-io/zps/action"
	"github.com/naegelejd/go-acl/os/group"
)

type DirUnix struct {
	*emission.Emitter
	dir *action.Dir

	phaseMap map[string]string
}

func NewDirUnix(dir action.Action, phaseMap map[string]string, emitter *emission.Emitter) Provider {
	return &DirUnix{emitter, dir.(*action.Dir), phaseMap}
}

func (d *DirUnix) Realize(ctx context.Context) error {
	switch d.phaseMap[Phase(ctx)] {
	case "install":
		return d.install(ctx)
	case "package":
		d.Emit("action.info", fmt.Sprintf("%s %s", d.dir.Type(), d.dir.Key()))
		return d.pkg(ctx)
	case "remove":
		return d.remove(ctx)
	default:
		return nil
	}
}

func (d *DirUnix) install(ctx context.Context) error {
	options := Opts(ctx)
	target := path.Join(options.TargetPath, d.dir.Path)

	mode, err := strconv.ParseUint(d.dir.Mode, 0, 0)
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

	owner, _ := user.Lookup(d.dir.Owner)
	grp, _ := user.LookupGroup(d.dir.Group)
	var uid int64
	var gid int64

	if owner != nil && grp != nil {
		uid, _ = strconv.ParseInt(owner.Uid, 0, 0)
		gid, _ = strconv.ParseInt(grp.Gid, 0, 0)
	}

	os.Chown(target, int(uid), int(gid))

	return nil
}

func (d *DirUnix) pkg(ctx context.Context) error {
	options := Opts(ctx)
	target := path.Join(options.TargetPath, d.dir.Path)

	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	if d.dir.Mode == "" {
		d.dir.Mode = fmt.Sprintf("%#o", info.Mode().Perm())
	}

	if d.dir.Owner == "" {
		if options.Secure {
			d.dir.Owner = "root"
		} else if options.Owner != "" {
			d.dir.Owner = options.Owner
		} else {
			usr, err := user.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Uid))
			if err != nil {
				return err
			}
			d.dir.Owner = usr.Username
		}
	}

	if d.dir.Group == "" {
		if options.Secure {
			d.dir.Group = "root"
		} else if options.Group != "" {
			d.dir.Group = options.Group
		} else {
			grp, err := group.LookupId(fmt.Sprint(info.Sys().(*syscall.Stat_t).Gid))
			if err != nil {
				return err
			}
			d.dir.Group = grp.Name
		}
	}

	return err
}

func (d *DirUnix) remove(ctx context.Context) error {
	options := Opts(ctx)
	target := path.Join(options.TargetPath, d.dir.Path)

	empty, err := d.isEmpty(target)
	if err != nil {
		return err
	}

	if empty == false {
		return nil
	}

	return os.Remove(target)
}

func (d *DirUnix) isEmpty(name string) (bool, error) {
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
