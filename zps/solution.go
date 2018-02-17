package zps

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type Solution struct {
	operations []*Operation
	names      []string

	installGraph *simple.DirectedGraph
	removeGraph  *simple.DirectedGraph

	installRoot graph.Node
	removeRoot  graph.Node

	operationMap map[graph.Node]*Operation
	nameMap      map[string]*Operation
}

func NewSolution() *Solution {
	solution := &Solution{}

	solution.installGraph = simple.NewDirectedGraph()
	solution.removeGraph = simple.NewDirectedGraph()

	solution.installRoot = solution.installGraph.NewNode()
	solution.removeRoot = solution.removeGraph.NewNode()
	solution.operationMap = make(map[graph.Node]*Operation)
	solution.nameMap = make(map[string]*Operation)

	return solution
}

func (s *Solution) AddOperation(operation *Operation) {
	s.operations = append(s.operations, operation)
	if operation.Operation == "noop" || operation.Operation == "install" {
		s.names = append(s.names, operation.Package.Name())
	}

	s.nameMap[operation.Package.Name()] = operation

	switch operation.Operation {
	case "install", "noop":
		operation.Node = s.installGraph.NewNode()
		s.installGraph.AddNode(operation.Node)
		s.operationMap[operation.Node] = operation
	case "remove":
		operation.Node = s.removeGraph.NewNode()
		s.removeGraph.AddNode(operation.Node)
		s.operationMap[operation.Node] = operation
	}
}

func (s *Solution) Get(name string) Solvable {
	for index := range s.operations {
		if s.operations[index].Package.Name() == name && s.operations[index].Operation != "remove" {
			return s.operations[index].Package
		}
	}

	return nil
}

func (s *Solution) Names() []string {
	return s.names
}

func (s *Solution) Noop() bool {
	for _, op := range s.operations {
		if op.Operation != "noop" {
			return false
		}
	}

	return true
}

func (s *Solution) Operations() []*Operation {
	return s.operations
}

func (s *Solution) Graph() ([]*Operation, error) {
	var operations []*Operation

	for _, op := range s.operations {
		for _, req := range op.Package.Requirements() {
			switch op.Operation {
			case "install", "noop":
				edge := s.installGraph.NewEdge(s.nameMap[req.Name].Node, op.Node)
				s.installGraph.SetEdge(edge)
			case "remove":
				edge := s.removeGraph.NewEdge(s.nameMap[req.Name].Node, op.Node)
				s.removeGraph.SetEdge(edge)
			}
		}
	}

	removes, err := topo.Sort(s.removeGraph)
	if err != nil {
		return nil, err
	}

	installs, err := topo.Sort(s.installGraph)
	if err != nil {
		return nil, err
	}

	for _, node := range append(removes, installs...) {
		if s.operationMap[node] == nil {
			continue
		}
		operations = append(operations, s.operationMap[node])
	}

	return operations, nil
}

type Solutions []Solution

func (slice Solutions) Len() int {
	return len(slice)
}

func (slice Solutions) Less(i, j int) bool {

	// TODO eliminate reliance on this sort, policy should select best solution iteratively
	if len(slice[i].Names()) > len(slice[j].Names()) {
		return true
	}

	for index, name := range slice[i].Names() {
		if index+1 == len(slice[i].Names()) {
			return slice[i].Get(name).Version().LT(slice[j].Get(name).Version())
		}

		if slice[i].Get(name).Version().GT(slice[j].Get(name).Version()) {
			return true
		}

		if slice[i].Get(name).Version().LT(slice[j].Get(name).Version()) {
			return false
		}
	}

	return false
}

func (slice Solutions) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
