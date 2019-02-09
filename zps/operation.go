/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zps

import "gonum.org/v1/gonum/graph"

type Operation struct {
	Operation string
	Package   Solvable

	Node graph.Node
}

func NewOperation(op string, pkg Solvable) *Operation {
	return &Operation{op, pkg, nil}
}
