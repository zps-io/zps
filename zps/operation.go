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
