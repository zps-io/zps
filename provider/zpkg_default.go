package provider

import (
	"context"

	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/action"
)

type ZpkgDefault struct {
	*emission.Emitter
	zpkg *action.Zpkg

	phaseMap map[string]string
}

func NewZpkgDefault(zpkg action.Action, phaseMap map[string]string, emitter *emission.Emitter) *ZpkgDefault {
	return &ZpkgDefault{emitter, zpkg.(*action.Zpkg), phaseMap}
}

func (z *ZpkgDefault) Realize(ctx context.Context) error {
	switch z.phaseMap[Phase(ctx)] {
	default:
		return nil
	}
}
