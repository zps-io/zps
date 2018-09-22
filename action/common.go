/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package action

type Action interface {
	Id() string
	Key() string
	Type() string
	Columns() string
	Condition() *bool
	MayFail() bool
	IsValid() bool
}

type Actions []Action

func (slice Actions) Len() int {
	return len(slice)
}

func (slice Actions) Less(i, j int) bool {
	return slice[i].Key() < slice[j].Key()
}

func (slice Actions) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
