package zps

type Job struct {
	op          string
	requirement *Requirement
}

func NewJob(op string, requirement *Requirement) *Job {
	return &Job{op, requirement}
}

func (j *Job) Op() string {
	return j.op
}

func (j *Job) Requirement() *Requirement {
	return j.requirement
}
