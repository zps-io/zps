package zps

type Operation struct {
	Operation string
	Package   Solvable
}

func NewOperation(op string, pkg Solvable) *Operation {
	return &Operation{op, pkg}
}
