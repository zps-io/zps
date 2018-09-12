package provider

import (
	"context"
	"fmt"

	"github.com/chuckpreslar/emission"

	"github.com/solvent-io/zps/action"
)

type TagDefault struct {
	*emission.Emitter
	tag *action.Tag

	phaseMap map[string]string
}

func NewTagDefault(tag action.Action, phaseMap map[string]string, emitter *emission.Emitter) Provider {
	return &TagDefault{emitter, tag.(*action.Tag), phaseMap}
}

func (t *TagDefault) Realize(ctx context.Context) error {
	switch t.phaseMap[Phase(ctx)] {
	default:
		t.Emit("action.info", fmt.Sprintf("%s %s", t.tag.Type(), t.tag.Key()))
		return nil
	}
}
