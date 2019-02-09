/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

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
