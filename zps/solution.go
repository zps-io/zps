package zps

type Solution struct {
	operations []*Operation
	names      []string
}

func (s *Solution) AddOperation(operation *Operation) {
	s.operations = append(s.operations, operation)
	if operation.Operation == "noop" || operation.Operation == "install" {
		s.names = append(s.names, operation.Package.Name())
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

func (s *Solution) Operations() []*Operation {
	return s.operations
}

type Solutions []Solution

func (slice Solutions) Len() int {
	return len(slice)
}

func (slice Solutions) Less(i, j int) bool {
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
