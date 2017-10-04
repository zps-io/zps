package requirement

import (
	"github.com/solvent-io/zps/action"
	"golang.org/x/net/context"
)

type Default struct {
	requirement *action.Requirement
}

func NewDefault(req action.Action) *Default {
	return &Default{req.(*action.Requirement)}
}

func (d *Default) Realize(phase string, ctx context.Context) error {
	switch phase {
	default:
		return nil
	}
}
