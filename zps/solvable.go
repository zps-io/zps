/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zps

type Solvable interface {
	Id() string
	Name() string
	Version() *Version
	Requirements() []*Requirement
	Arch() string
	Os() string

	FileName() string

	Location() int
	SetLocation(location int)

	Priority() int
	SetPriority(priority int)

	SetChannels(...string)
	Channels() []string

	Satisfies(*Requirement) bool
}

type Solvables []Solvable

func (slice Solvables) Len() int {
	return len(slice)
}

func (slice Solvables) Less(i, j int) bool {
	if slice[i].Name() < slice[j].Name() {
		return true
	}
	if slice[i].Name() > slice[j].Name() {
		return false
	}

	if slice[i].Priority() < slice[j].Priority() {
		return true
	}
	if slice[i].Priority() > slice[j].Priority() {
		return false
	}

	return slice[i].Version().GT(slice[j].Version())
}

func (slice Solvables) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
