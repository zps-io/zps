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
	"errors"
	"sort"

	"github.com/fezz-io/sat"
	//"github.com/davecgh/go-spew/spew"
)

type Solver struct {
	pool         *Pool
	policy       Policy
	request      *Request
	satSolutions []*sat.Solution
	solutions    Solutions
	solver       *sat.Solver
}

func NewSolver(pool *Pool, policy Policy) *Solver {
	solver := &Solver{pool, policy, nil, nil, nil, sat.NewSolver()}
	return solver
}

func (s *Solver) Solve(request *Request) (*Solution, error) {
	var satisfiable bool
	s.request = request
	s.solutions = nil

	s.addClauses()

	satisfiable, s.satSolutions = s.solver.Satisfiable()
	if satisfiable == false {
		return nil, errors.New("zps.solver: No solution for requested jobs")
	}

	s.generateSolutions()

	return s.getSolution(), nil
}

func (s *Solver) Cnf() string {
	return s.solver.String()
}

func (s *Solver) SatSolutions() []*sat.Solution {
	return s.satSolutions
}

func (s *Solver) Solutions() Solutions {
	return s.solutions
}

func (s *Solver) addClauses() {
	// Add unary clauses
	for _, job := range s.request.Jobs() {
		switch job.Op() {
		case "install":
			var clause sat.LiteralEncoder
			candidate := s.policy.SelectRequest(s.pool.WhatProvides(job.Requirement()))

			if candidate != nil {
				clause = sat.NewVariable(candidate.Id())
				s.solver.AddClause(clause)
				s.addReqClauses(nil, candidate)
			}
		case "remove":
			var clause sat.LiteralEncoder
			candidate := s.pool.Installed(job.Requirement())

			if candidate != nil {
				clause = sat.NewVariable(candidate.Id()).Not()
				s.solver.AddClause(clause)
				s.addRmClauses(candidate)
			}
		default:
			continue
		}
	}
}

func (s *Solver) addReqClauses(parent Solvable, solvable Solvable) {
	for _, req := range solvable.Requirements() {
		// Continue if a requirement references itself
		if solvable.Name() == req.Name {
			continue
		}

		switch req.Method {
		case "depends":
			var clause []sat.LiteralEncoder

			clause = append(clause, sat.NewVariable(solvable.Id()).Not())
			provides := s.policy.PruneProvides(s.pool.WhatProvides(req))

			for _, candidate := range provides {
				clause = append(clause, sat.NewVariable(candidate.Id()))
				// recurse if candidate is not in fact the parent
				if parent != nil {
					if candidate.Name() == parent.Name() {
						continue
					}
				}

				s.addReqClauses(solvable, candidate)
			}

			s.solver.AddClause(clause...)

			for index, provided := range provides {
				current := index + 1
				for current <= len(provides)-1 {
					if provided.Id() != provides[current].Id() {
						s.solver.AddClause(sat.NewVariable(provided.Id()).Not(), sat.NewVariable(provides[current].Id()).Not())
					}
					current++
				}
			}

		case "conflicts":
			for _, candidate := range s.policy.PruneProvides(s.pool.WhatProvides(req)) {
				s.solver.AddClause(sat.NewVariable(solvable.Id()).Not(), sat.NewVariable(candidate.Id()).Not())
			}
		default:
			continue
		}
	}
}

func (s *Solver) addRmClauses(solvable Solvable) {
	for _, dep := range s.pool.WhatDepends(solvable.Name()) {
		clause := sat.NewVariable(dep.Id()).Not()
		s.solver.AddClause(clause)
		// recurse
		s.addRmClauses(dep)
	}
}

func (s *Solver) generateSolutions() {
	for _, satSol := range s.satSolutions {
		solution := NewSolution()
		keys := make([]string, 0)

		for key := range *satSol {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, key := range keys {
			req, _ := NewRequirementFromSimpleString(key)
			pkg := s.policy.SelectRequest(s.pool.WhatProvides(req))

			if satSol.Value(key) {
				if pkg.Priority() > -1 {
					solution.AddOperation(NewOperation("install", pkg))
				} else {
					solution.AddOperation(NewOperation("noop", pkg))
				}
			} else {
				if pkg.Priority() == -1 {
					solution.AddOperation(NewOperation("remove", pkg))
				}
			}
		}

		s.solutions = append(s.solutions, *solution)
	}

	sort.Sort(s.solutions)
}

func (s *Solver) getSolution() *Solution {
	return s.policy.SelectSolution(s.solutions)
}
