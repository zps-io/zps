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

func NewRequirementDefault(req action.Action, phaseMap map[string]string, emitter *emission.Emitter) Provider {
	return &RequirementDefault{emitter, req.(*action.Requirement), phaseMap}
}

func (r *RequirementDefault) Realize(ctx context.Context) error {
	switch r.phaseMap[Phase(ctx)] {
	default:
		return nil
	}
}
