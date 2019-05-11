/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zps

import (
	"sort"
)

type Policy interface {
	PruneProvides(solvables Solvables) Solvables
	SelectRequest(solvables Solvables) Solvable
	SelectSolution(solutions Solutions) *Solution
}

type UpdatedPolicy struct{}
type InstalledPolicy struct{}

func NewPolicy(method string) Policy {
	switch method {
	case "updated":
		return &UpdatedPolicy{}
	case "installed":
		return &InstalledPolicy{}
	}

	return nil
}

func (u *UpdatedPolicy) PruneProvides(solvables Solvables) Solvables {
	if len(solvables) == 0 {
		return solvables
	}

	sort.Sort(solvables)

	if solvables[0].Priority() == -2 {
		return Solvables{solvables[0]}
	}

	return solvables
}

func (u *UpdatedPolicy) SelectRequest(solvables Solvables) Solvable {
	sort.Sort(solvables)

	for _, solvable := range solvables {
		if len(solvables) > 1 && solvable.Priority() == -1 {
			if solvables[1].Version().GT(solvables[0].Version()) {
				return solvables[1]
			}
		}

		return solvable
	}

	return nil
}

func (u *UpdatedPolicy) SelectSolution(solutions Solutions) *Solution {

	solution := solutions[0]
	return &solution
}

func (i *InstalledPolicy) PruneProvides(solvables Solvables) Solvables {
	return solvables
}

func (i *InstalledPolicy) SelectRequest(solvables Solvables) Solvable {
	sort.Sort(solvables)

	for _, solvable := range solvables {
		if solvable.Priority() <= -1 {
			return solvable
		}
	}

	for _, solvable := range solvables {
		return solvable
	}

	return nil
}

func (i *InstalledPolicy) SelectSolution(solutions Solutions) *Solution {
	return nil
}
