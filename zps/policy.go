package zps

import (
	"sort"
)

// TODO the policy method will be responsible for pruning the proposed solver solution
// installed vs latest, repository priority, organization priority

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
	var result Solvables

	seen := -1
	for index := range solvables {
		if seen == -1 {
			seen = solvables[index].Priority()
		}

		result = append(result, solvables[index])

		if len(solvables)-1 >= index+1 {
			if seen != solvables[index+1].Priority() && seen != -1 {
				break
			}
		}

	}

	return result
}

func (u *UpdatedPolicy) SelectRequest(solvables Solvables) Solvable {
	sort.Sort(solvables)

	for _, solvable := range solvables {
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
		if solvable.Priority() == -1 {
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
