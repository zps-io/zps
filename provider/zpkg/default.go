package zpkg

import (
	"golang.org/x/net/context"

	"github.com/solvent-io/zps/action"
)

type Default struct {
	zpkg *action.Zpkg
}

func NewDefault(zpkg action.Action) *Default {
	return &Default{zpkg.(*action.Zpkg)}
}

func (d *Default) Realize(phase string, ctx context.Context) error {
	switch phase {
	default:
		return nil
	}
}
