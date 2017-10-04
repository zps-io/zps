package meta

import (
	"golang.org/x/net/context"

	"github.com/solvent-io/zps/action"
)

type Default struct {
	meta *action.Meta
}

func NewDefault(meta action.Action) *Default {
	return &Default{meta.(*action.Meta)}
}

func (d *Default) Realize(phase string, ctx context.Context) error {
	switch phase {
	default:
		return nil
	}
}
