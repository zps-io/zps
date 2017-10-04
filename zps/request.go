package zps

type Request struct {
	jobs []*Job
}

func NewRequest() *Request {
	request := &Request{}
	return request
}

func (r *Request) Jobs() []*Job {
	return r.jobs
}

func (r *Request) Install(requirement *Requirement) {
	r.addJob(requirement, "install")
}

func (r *Request) Update(requirement *Requirement) {
	r.addJob(requirement, "update")
}

func (r *Request) Remove(requirement *Requirement) {
	r.addJob(requirement, "remove")
}

func (r *Request) Upgrade() {
	r.jobs = append(r.jobs, NewJob("upgrade", nil))
}

func (r *Request) addJob(requirement *Requirement, op string) {
	r.jobs = append(r.jobs, NewJob(op, requirement))
}
