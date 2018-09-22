/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package zps

import (
	"errors"
	"strings"
	"time"

	"github.com/blang/semver"
)

type Version struct {
	Semver    semver.Version
	Timestamp time.Time
}

func (v *Version) Parse(version string) error {
	split := strings.Split(version, ":")
	var err error = nil

	if len(split) < 1 {
		return errors.New("zps.Version: error parsing version")
	}

	v.Semver, err = semver.Make(split[0])
	if err != nil {
		return errors.New("zps.Version: error parsing version (semver)")
	}

	if len(split) == 2 {
		v.Timestamp, err = time.Parse("20060102T150405Z", split[1])
	}

	return err
}

func (v *Version) String() string {
	s := []string{v.Semver.String(), v.Timestamp.Format("20060102T150405Z")}
	return strings.Join(s, ":")
}

func (v *Version) Short() string {
	return v.Semver.String()
}

func (v *Version) Compare(ve *Version) int {
	if v.Semver.GT(ve.Semver) {
		return 1
	}

	if v.Semver.LT(ve.Semver) {
		return -1
	}

	if v.Timestamp.After(ve.Timestamp) {
		return 1
	}

	if v.Timestamp.Before(ve.Timestamp) {
		return -1
	}

	if v.Timestamp.Equal(ve.Timestamp) {
		return 2
	}

	return 0
}

func (v *Version) EQ(ve *Version) bool {
	compare := v.Compare(ve)
	return (compare == 0 || compare == 2)
}

func (v *Version) EXQ(ve *Version) bool {
	return (v.Compare(ve) == 2)
}

func (v *Version) GT(ve *Version) bool {
	return (v.Compare(ve) == 1)
}

func (v *Version) GTE(ve *Version) bool {
	return (v.Compare(ve) >= 0)
}

func (v *Version) LT(ve *Version) bool {
	return (v.Compare(ve) == -1)
}

func (v *Version) LTE(ve *Version) bool {
	return (v.Compare(ve) <= 0)
}

func (v *Version) NEQ(ve *Version) bool {
	return (v.Compare(ve) != 0)
}
