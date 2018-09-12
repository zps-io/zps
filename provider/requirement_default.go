package provider

import (
	"context"

	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/action"
)

type RequirementDefault struct {
	*emission.Emitter
	requirement *action.Requirement

	phaseMap map[string]string
}

func NewRequirementDefault(zpkg action.Action, phaseMap map[string]string, emitter *emission.Emitter) *RequirementDefault {
	return &RequirementDefault{emitter, zpkg.(*action.Zpkg), phaseMap}
}

func (r *RequirementDefault) Realize(ctx context.Context) error {
	switch r.phaseMap[Phase(ctx)] {
	default:
		return nil
	}
}
