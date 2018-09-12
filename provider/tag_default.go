package provider

import (
	"context"

	"github.com/chuckpreslar/emission"

	"github.com/solvent-io/zps/action"
)

type TagDefault struct {
	*emission.Emitter
	tag *action.Tag

	phaseMap map[string]string
}

func NewTagDefault(tag action.Action, phaseMap map[string]string, emitter *emission.Emitter) *TagDefault {
	return &TagDefault{emitter, tag.(*action.Tag), phaseMap}
}

func (t *TagDefault) Realize(ctx context.Context) error {
	switch t.phaseMap[Phase(ctx)] {
	default:
		return nil
	}
}
